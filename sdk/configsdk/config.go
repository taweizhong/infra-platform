package configsdk

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var scopePriority = map[string]int{
	"global": 1,
}

func scopeOrder(scope string) int {
	if p, ok := scopePriority[scope]; ok {
		return p
	}
	switch {
	case len(scope) >= 4 && scope[:4] == "env:":
		return 2
	case len(scope) >= 8 && scope[:8] == "cluster:":
		return 3
	case len(scope) >= 4 && scope[:4] == "app:":
		return 4
	case len(scope) >= 9 && scope[:9] == "instance:":
		return 5
	default:
		return 99
	}
}

type FileChangeEvent struct {
	Name    string
	Content string
	Scope   string
	Meta    Metadata
}

func (e FileChangeEvent) Unmarshal(v any) error {
	return decodeTOML(e.Content, v)
}

type configModule struct {
	client *Client

	mu          sync.RWMutex
	scopes      map[string][]ConfigFile
	merged      map[string]string
	fileMeta    map[string]Metadata
	hotMD5      map[string]string
	fileCB      map[string][]func(FileChangeEvent)
	anyCB       []func()
	cancelWatch context.CancelFunc
}

func newConfigModule(c *Client) *configModule {
	return &configModule{
		client:   c,
		scopes:   map[string][]ConfigFile{},
		merged:   map[string]string{},
		fileMeta: map[string]Metadata{},
		hotMD5:   map[string]string{},
		fileCB:   map[string][]func(FileChangeEvent){},
	}
}

func (c *configModule) Init(ctx context.Context) error {
	resp := ConfigFetchResponse{}
	q := mapToQuery(map[string]string{
		"app":      c.client.opts.app,
		"env":      c.client.opts.env,
		"cluster":  c.client.opts.cluster,
		"instance": c.client.opts.instance,
	})
	if err := c.client.tp.get(ctx, "/api/v1/config", q, &resp); err != nil {
		if c.client.opts.fileCachePath != "" {
			if cacheErr := c.loadFromFileCache(); cacheErr == nil {
				return nil
			}
		}
		return err
	}
	if len(resp.Scopes) == 0 {
		return errors.New("empty config scopes")
	}

	c.mu.Lock()
	c.scopes = resp.Scopes
	c.rebuildLocked()
	c.mu.Unlock()

	if c.client.opts.fileCachePath != "" {
		_ = c.persistToFileCache()
	}
	if c.client.opts.enableWatch {
		ctx, cancel := context.WithCancel(context.Background())
		c.cancelWatch = cancel
		go c.watchLoop(ctx)
	}
	return nil
}

func (c *configModule) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelWatch != nil {
		c.cancelWatch()
		c.cancelWatch = nil
	}
}

func (c *configModule) Unmarshal(v any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	joined := ""
	keys := make([]string, 0, len(c.merged))
	for k := range c.merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		joined += c.merged[k] + "\n"
	}
	return decodeTOML(joined, v)
}

func (c *configModule) UnmarshalFile(name string, v any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	content, ok := c.merged[name]
	if !ok {
		return fmt.Errorf("config file not found: %s", name)
	}
	return decodeTOML(content, v)
}

func (c *configModule) OnFileChange(name string, cb func(FileChangeEvent)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fileCB[name] = append(c.fileCB[name], cb)
}

func (c *configModule) OnAnyChange(cb func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.anyCB = append(c.anyCB, cb)
}

func (c *configModule) watchLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		req := ConfigWatchRequest{
			App:      c.client.opts.app,
			Env:      c.client.opts.env,
			Cluster:  c.client.opts.cluster,
			Instance: c.client.opts.instance,
			HotFiles: c.currentHotMD5(),
		}
		callCtx, cancel := context.WithTimeout(ctx, c.client.opts.watchTimeout+time.Second)
		resp := ConfigWatchResponse{}
		err := c.client.tp.post(callCtx, fmt.Sprintf("/api/v1/config/watch?timeout=%s", c.client.opts.watchTimeout.String()), req, &resp)
		cancel()
		if err != nil {
			time.Sleep(c.client.opts.retryInterval)
			continue
		}
		if !resp.Changed || len(resp.Files) == 0 {
			continue
		}
		c.applyChanges(resp.Files)
	}
}

func (c *configModule) applyChanges(files []ConfigFile) {
	c.mu.Lock()
	changedNames := map[string]FileChangeEvent{}
	for _, f := range files {
		existing := c.scopes[f.Scope]
		replaced := false
		for i := range existing {
			if existing[i].Name == f.Name {
				existing[i] = f
				replaced = true
				break
			}
		}
		if !replaced {
			existing = append(existing, f)
		}
		c.scopes[f.Scope] = existing
		changedNames[f.Name] = FileChangeEvent{Name: f.Name, Scope: f.Scope, Content: f.Content, Meta: f.Metadata}
	}
	c.rebuildLocked()
	callbacks := make(map[string][]func(FileChangeEvent), len(changedNames))
	for name := range changedNames {
		callbacks[name] = append(callbacks[name], c.fileCB[name]...)
	}
	anyCallbacks := append([]func(){}, c.anyCB...)
	c.mu.Unlock()

	for name, ev := range changedNames {
		for _, cb := range callbacks[name] {
			cb(ev)
		}
	}
	for _, cb := range anyCallbacks {
		cb()
	}
	if c.client.opts.fileCachePath != "" {
		_ = c.persistToFileCache()
	}
}

func (c *configModule) rebuildLocked() {
	grouped := make(map[string][]ConfigFile)
	for scope, files := range c.scopes {
		for _, f := range files {
			if f.Scope == "" {
				f.Scope = scope
			}
			grouped[f.Name] = append(grouped[f.Name], f)
		}
	}
	newMerged := make(map[string]string, len(grouped))
	newMeta := make(map[string]Metadata, len(grouped))
	newHot := map[string]string{}

	for name, files := range grouped {
		sort.Slice(files, func(i, j int) bool {
			return scopeOrder(files[i].Scope) < scopeOrder(files[j].Scope)
		})
		mergedMap := map[string]any{}
		for _, file := range files {
			content := file.Content
			if decrypt := c.client.opts.decryptor; decrypt != nil && boolFromMeta(file.Metadata, "encrypted") {
				if plain, err := decrypt.Decrypt(content); err == nil {
					content = plain
				}
			}
			if frag, err := parseTOML(content); err == nil {
				deepMerge(mergedMap, frag)
			}
		}
		str, _ := encodeTOML(mergedMap)
		newMerged[name] = str
		meta := files[len(files)-1].Metadata
		newMeta[name] = meta
		if boolFromMeta(meta, "hot") {
			newHot[name] = md5sum(str)
		}
	}
	c.merged = newMerged
	c.fileMeta = newMeta
	c.hotMD5 = newHot
}

func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if srcMap, ok := v.(map[string]any); ok {
			if dstMap, ok := dst[k].(map[string]any); ok {
				deepMerge(dstMap, srcMap)
				dst[k] = dstMap
				continue
			}
		}
		dst[k] = v
	}
}

func (c *configModule) currentHotMD5() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cp := make(map[string]string, len(c.hotMD5))
	for k, v := range c.hotMD5 {
		cp[k] = v
	}
	return cp
}

func md5sum(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func boolFromMeta(meta Metadata, key string) bool {
	if meta == nil {
		return false
	}
	v, ok := meta[key]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

func (c *configModule) persistToFileCache() error {
	c.mu.RLock()
	data, err := json.Marshal(c.scopes)
	c.mu.RUnlock()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(c.client.opts.fileCachePath, 0o755); err != nil {
		return err
	}
	file := filepath.Join(c.client.opts.fileCachePath, "config_cache.json")
	return os.WriteFile(file, data, 0o644)
}

func (c *configModule) loadFromFileCache() error {
	file := filepath.Join(c.client.opts.fileCachePath, "config_cache.json")
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	scopes := map[string][]ConfigFile{}
	if err := json.Unmarshal(b, &scopes); err != nil {
		return err
	}
	c.mu.Lock()
	c.scopes = scopes
	c.rebuildLocked()
	c.mu.Unlock()
	return nil
}
