package configsdk

import "time"

type Metadata map[string]any

type ConfigFile struct {
	Name     string   `json:"name"`
	Scope    string   `json:"scope,omitempty"`
	Content  string   `json:"content"`
	Format   string   `json:"format,omitempty"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type ConfigFetchResponse struct {
	Scopes map[string][]ConfigFile `json:"scopes"`
}

type ConfigWatchRequest struct {
	App      string            `json:"app"`
	Env      string            `json:"env"`
	Cluster  string            `json:"cluster"`
	Instance string            `json:"instance,omitempty"`
	HotFiles map[string]string `json:"hot_files"`
}

type ConfigWatchResponse struct {
	Changed bool         `json:"changed"`
	Files   []ConfigFile `json:"files,omitempty"`
}

type KVEntry struct {
	Path     string   `json:"path"`
	Value    []byte   `json:"value"`
	Version  int64    `json:"version"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type kvPutRequest struct {
	Value    []byte   `json:"value"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type kvPutResponse struct {
	Version int64 `json:"version"`
}

type KVWatchRequest struct {
	Path        string `json:"path,omitempty"`
	Prefix      string `json:"prefix,omitempty"`
	FromVersion int64  `json:"from_version"`
}

type KVWatchEvent struct {
	Type  string  `json:"type"`
	Entry KVEntry `json:"entry"`
}

type KVWatchResponse struct {
	Changed bool           `json:"changed"`
	Events  []KVWatchEvent `json:"events,omitempty"`
}

type ServiceInstance struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Metadata map[string]string `json:"metadata,omitempty"`
	TTL      int               `json:"ttl"`
}

type DiscoveryWatchEvent struct {
	Type      string          `json:"type"`
	Service   string          `json:"service"`
	Instance  ServiceInstance `json:"instance"`
	Version   int64           `json:"version"`
	OccurTime time.Time       `json:"occur_time"`
}
