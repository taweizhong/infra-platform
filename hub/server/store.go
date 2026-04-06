package server

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

type Store struct {
	mu sync.RWMutex

	envs     map[string]Environment
	clusters map[string]Cluster
	apps     map[string]App

	configs       map[string]ConfigFile
	configByScope map[string][]string
	configHistory map[string][]ConfigHistory
	releases      map[string]Release

	idSeq atomic.Int64
}

func NewStore() *Store {
	return &Store{
		envs:          map[string]Environment{},
		clusters:      map[string]Cluster{},
		apps:          map[string]App{},
		configs:       map[string]ConfigFile{},
		configByScope: map[string][]string{},
		configHistory: map[string][]ConfigHistory{},
		releases:      map[string]Release{},
	}
}

func (s *Store) nextID(prefix string) string {
	id := s.idSeq.Add(1)
	return fmt.Sprintf("%s-%d", prefix, id)
}

func (s *Store) UpsertEnv(name string) Environment {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := Environment{Name: name, CreatedAt: time.Now().UTC()}
	if old, ok := s.envs[name]; ok {
		e.CreatedAt = old.CreatedAt
	}
	s.envs[name] = e
	return e
}
func (s *Store) ListEnvs() []Environment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Environment, 0, len(s.envs))
	for _, v := range s.envs {
		out = append(out, v)
	}
	return out
}
func (s *Store) DeleteEnv(name string) { s.mu.Lock(); defer s.mu.Unlock(); delete(s.envs, name) }

func (s *Store) UpsertCluster(name, env string) Cluster {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := Cluster{Name: name, Env: env, Status: "up", CreatedAt: time.Now().UTC()}
	if old, ok := s.clusters[name]; ok {
		c.CreatedAt = old.CreatedAt
	}
	s.clusters[name] = c
	return c
}
func (s *Store) GetCluster(name string) (Cluster, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.clusters[name]
	return c, ok
}
func (s *Store) ListClusters() []Cluster {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Cluster, 0, len(s.clusters))
	for _, v := range s.clusters {
		out = append(out, v)
	}
	return out
}
func (s *Store) DeleteCluster(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clusters, name)
}

func (s *Store) UpsertApp(name, owner string) App {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := App{Name: name, Owner: owner, CreatedAt: time.Now().UTC()}
	if old, ok := s.apps[name]; ok {
		a.CreatedAt = old.CreatedAt
	}
	s.apps[name] = a
	return a
}
func (s *Store) ListApps() []App {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]App, 0, len(s.apps))
	for _, v := range s.apps {
		out = append(out, v)
	}
	return out
}
func (s *Store) GetApp(name string) (App, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.apps[name]
	return a, ok
}
func (s *Store) DeleteApp(name string) { s.mu.Lock(); defer s.mu.Unlock(); delete(s.apps, name) }

func (s *Store) CreateConfig(in ConfigFile) ConfigFile {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	in.ID = s.nextID("cfg")
	in.Version = 1
	in.CreatedAt = now
	in.UpdatedAt = now
	if in.Format == "" {
		in.Format = "toml"
	}
	s.configs[in.ID] = in
	s.configByScope[in.Scope] = append(s.configByScope[in.Scope], in.ID)
	s.configHistory[in.ID] = append(s.configHistory[in.ID], ConfigHistory{ConfigID: in.ID, Version: in.Version, Content: in.Content, Metadata: in.Metadata, ModifiedAt: now})
	return in
}

func (s *Store) UpdateConfig(id string, in ConfigFile) (ConfigFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.configs[id]
	if !ok {
		return ConfigFile{}, errors.New("config not found")
	}
	old.Scope = in.Scope
	old.Name = in.Name
	old.Content = in.Content
	old.Format = in.Format
	old.Metadata = in.Metadata
	if old.Format == "" {
		old.Format = "toml"
	}
	old.Version++
	old.UpdatedAt = time.Now().UTC()
	s.configs[id] = old
	s.configHistory[id] = append(s.configHistory[id], ConfigHistory{ConfigID: id, Version: old.Version, Content: old.Content, Metadata: old.Metadata, ModifiedAt: old.UpdatedAt})
	return old, nil
}

func (s *Store) DeleteConfig(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg, ok := s.configs[id]
	if !ok {
		return
	}
	delete(s.configs, id)
	ids := s.configByScope[cfg.Scope]
	for i := range ids {
		if ids[i] == id {
			s.configByScope[cfg.Scope] = slices.Delete(ids, i, i+1)
			break
		}
	}
}

func (s *Store) GetConfig(id string) (ConfigFile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.configs[id]
	return v, ok
}

func (s *Store) ListConfigs(scope string) []ConfigFile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []ConfigFile{}
	if scope == "" {
		for _, c := range s.configs {
			out = append(out, c)
		}
		return out
	}
	for _, id := range s.configByScope[scope] {
		if c, ok := s.configs[id]; ok {
			out = append(out, c)
		}
	}
	return out
}

func (s *Store) ConfigHistory(id string) []ConfigHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]ConfigHistory(nil), s.configHistory[id]...)
}

func (s *Store) ReleaseCreate(environment, cluster string, cfgIDs []string) Release {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	r := Release{ID: s.nextID("rel"), Status: ReleaseDraft, Environment: environment, Cluster: cluster, ConfigIDs: cfgIDs, CreatedAt: now, UpdatedAt: now}
	s.releases[r.ID] = r
	return r
}

func (s *Store) ReleaseGet(id string) (Release, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.releases[id]
	return r, ok
}

func (s *Store) ReleaseAdvance(id, action string) (Release, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.releases[id]
	if !ok {
		return Release{}, errors.New("release not found")
	}
	next, err := nextState(r.Status, action)
	if err != nil {
		return Release{}, err
	}
	r.Status = next
	r.UpdatedAt = time.Now().UTC()
	s.releases[id] = r
	return r, nil
}

func nextState(status, action string) (string, error) {
	switch action {
	case "approve":
		if status == ReleaseDraft || status == ReleasePending {
			return ReleaseApproved, nil
		}
	case "publish":
		if status == ReleaseApproved {
			return ReleasePublished, nil
		}
	case "rollback":
		if status == ReleasePublished {
			return ReleaseRolledBack, nil
		}
	}
	return "", fmt.Errorf("invalid transition: %s -> %s", status, action)
}
