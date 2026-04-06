package configsdk

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type kvModule struct {
	client *Client
}

func newKVModule(c *Client) *kvModule { return &kvModule{client: c} }

func (k *kvModule) Put(path string, value []byte) (int64, error) {
	return k.put(path, value, nil)
}

func (k *kvModule) PutWithTTL(path string, value []byte, ttl time.Duration) (int64, error) {
	meta := Metadata{"ttl": int(ttl.Seconds())}
	return k.put(path, value, meta)
}

func (k *kvModule) put(p string, value []byte, metadata Metadata) (int64, error) {
	resp := kvPutResponse{}
	err := k.client.tp.put(context.Background(), fmt.Sprintf("/api/v1/kv/%s", trimPath(p)), kvPutRequest{Value: value, Metadata: metadata}, &resp)
	if err != nil {
		return 0, err
	}
	return resp.Version, nil
}

func (k *kvModule) Get(path string) (KVEntry, error) {
	resp := KVEntry{}
	err := k.client.tp.get(context.Background(), fmt.Sprintf("/api/v1/kv/%s", trimPath(path)), nil, &resp)
	return resp, err
}

func (k *kvModule) Delete(path string) error {
	return k.client.tp.delete(context.Background(), fmt.Sprintf("/api/v1/kv/%s", trimPath(path)), nil)
}

func (k *kvModule) List(prefix string) ([]KVEntry, error) {
	var entries []KVEntry
	q := url.Values{}
	q.Set("list", "true")
	q.Set("recursive", "true")
	err := k.client.tp.get(context.Background(), fmt.Sprintf("/api/v1/kv/%s", trimPath(prefix)), q, &entries)
	return entries, err
}

func (k *kvModule) Watch(ctx context.Context, path string) (<-chan KVWatchEvent, error) {
	return k.watchLoop(ctx, KVWatchRequest{Path: path})
}

func (k *kvModule) WatchPrefix(ctx context.Context, prefix string) (<-chan KVWatchEvent, error) {
	return k.watchLoop(ctx, KVWatchRequest{Prefix: prefix})
}

func (k *kvModule) watchLoop(ctx context.Context, req KVWatchRequest) (<-chan KVWatchEvent, error) {
	ch := make(chan KVWatchEvent, 16)
	go func() {
		defer close(ch)
		fromVersion := int64(0)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			req.FromVersion = fromVersion
			resp := KVWatchResponse{}
			err := k.client.tp.post(ctx, fmt.Sprintf("/api/v1/kv/watch?timeout=%s", k.client.opts.watchTimeout.String()), req, &resp)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(k.client.opts.retryInterval):
					continue
				}
			}
			if !resp.Changed {
				continue
			}
			for _, ev := range resp.Events {
				if ev.Entry.Version > fromVersion {
					fromVersion = ev.Entry.Version
				}
				select {
				case ch <- ev:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch, nil
}

func trimPath(p string) string {
	return strings.TrimPrefix(p, "/")
}
