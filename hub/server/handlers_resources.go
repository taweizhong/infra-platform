package server

import (
	"net/http"
	"strings"
)

func (s *Server) handleEnvironments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.store.ListEnvs())
	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := readJSON(r, &req); err != nil || req.Name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		writeJSON(w, http.StatusCreated, s.store.UpsertEnv(req.Name))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleEnvironmentByName(w http.ResponseWriter, r *http.Request) {
	name := pathLast(r.URL.Path)
	switch r.Method {
	case http.MethodPut:
		writeJSON(w, http.StatusOK, s.store.UpsertEnv(name))
	case http.MethodDelete:
		s.store.DeleteEnv(name)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleClusters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.store.ListClusters())
	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
			Env  string `json:"env"`
		}
		if err := readJSON(r, &req); err != nil || req.Name == "" || req.Env == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		writeJSON(w, http.StatusCreated, s.store.UpsertCluster(req.Name, req.Env))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleClusterByName(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/clusters/")
	if strings.HasSuffix(path, "/status") {
		name := strings.TrimSuffix(path, "/status")
		c, ok := s.store.GetCluster(name)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"name": c.Name, "status": c.Status})
		return
	}
	name := path
	switch r.Method {
	case http.MethodGet:
		c, ok := s.store.GetCluster(name)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, c)
	case http.MethodPut:
		var req struct {
			Env string `json:"env"`
		}
		_ = readJSON(r, &req)
		if req.Env == "" {
			req.Env = "default"
		}
		writeJSON(w, http.StatusOK, s.store.UpsertCluster(name, req.Env))
	case http.MethodDelete:
		s.store.DeleteCluster(name)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleApps(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.store.ListApps())
	case http.MethodPost:
		var req struct{ Name, Owner string }
		if err := readJSON(r, &req); err != nil || req.Name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		writeJSON(w, http.StatusCreated, s.store.UpsertApp(req.Name, req.Owner))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAppByName(w http.ResponseWriter, r *http.Request) {
	name := pathLast(r.URL.Path)
	switch r.Method {
	case http.MethodGet:
		a, ok := s.store.GetApp(name)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, a)
	case http.MethodPut:
		var req struct {
			Owner string `json:"owner"`
		}
		_ = readJSON(r, &req)
		writeJSON(w, http.StatusOK, s.store.UpsertApp(name, req.Owner))
	case http.MethodDelete:
		s.store.DeleteApp(name)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
