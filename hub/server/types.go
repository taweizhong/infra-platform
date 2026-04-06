package server

import "time"

type Environment struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Cluster struct {
	Name      string    `json:"name"`
	Env       string    `json:"env"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type App struct {
	Name      string    `json:"name"`
	Owner     string    `json:"owner,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ConfigFile struct {
	ID        string         `json:"id"`
	Scope     string         `json:"scope"`
	Name      string         `json:"name"`
	Content   string         `json:"content"`
	Format    string         `json:"format"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Version   int64          `json:"version"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type ConfigHistory struct {
	ConfigID   string         `json:"config_id"`
	Version    int64          `json:"version"`
	Content    string         `json:"content"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	ModifiedAt time.Time      `json:"modified_at"`
}

type Release struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Environment string    `json:"environment"`
	Cluster     string    `json:"cluster,omitempty"`
	ConfigIDs   []string  `json:"config_ids"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	ReleaseDraft      = "draft"
	ReleasePending    = "pending"
	ReleaseApproved   = "approved"
	ReleasePublished  = "published"
	ReleaseRolledBack = "rolled_back"
)
