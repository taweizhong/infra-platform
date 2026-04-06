package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Server struct {
	store *Store
	mux   *http.ServeMux
}

func New() *Server {
	s := &Server{store: NewStore(), mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	s.mux.HandleFunc("/api/v1/environments", s.handleEnvironments)
	s.mux.HandleFunc("/api/v1/environments/", s.handleEnvironmentByName)
	s.mux.HandleFunc("/api/v1/clusters", s.handleClusters)
	s.mux.HandleFunc("/api/v1/clusters/", s.handleClusterByName)
	s.mux.HandleFunc("/api/v1/apps", s.handleApps)
	s.mux.HandleFunc("/api/v1/apps/", s.handleAppByName)

	s.mux.HandleFunc("/api/v1/configs", s.handleConfigs)
	s.mux.HandleFunc("/api/v1/configs/preview", s.handleConfigPreview)
	s.mux.HandleFunc("/api/v1/configs/", s.handleConfigByID)

	s.mux.HandleFunc("/api/v1/releases", s.handleReleases)
	s.mux.HandleFunc("/api/v1/releases/", s.handleReleaseByID)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, out any) error { return json.NewDecoder(r.Body).Decode(out) }

func pathLast(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
