package server

import (
	"net/http"
	"strings"
)

func (s *Server) handleReleases(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req struct {
			Environment string   `json:"environment"`
			Cluster     string   `json:"cluster"`
			ConfigIDs   []string `json:"config_ids"`
		}
		if err := readJSON(r, &req); err != nil || req.Environment == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		writeJSON(w, http.StatusCreated, s.store.ReleaseCreate(req.Environment, req.Cluster, req.ConfigIDs))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleReleaseByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/releases/")
	if strings.HasSuffix(path, "/approve") {
		s.handleReleaseAction(w, r, strings.TrimSuffix(path, "/approve"), "approve")
		return
	}
	if strings.HasSuffix(path, "/publish") {
		s.handleReleaseAction(w, r, strings.TrimSuffix(path, "/publish"), "publish")
		return
	}
	if strings.HasSuffix(path, "/rollback") {
		s.handleReleaseAction(w, r, strings.TrimSuffix(path, "/rollback"), "rollback")
		return
	}
	if strings.HasSuffix(path, "/status") {
		id := strings.TrimSuffix(path, "/status")
		rls, ok := s.store.ReleaseGet(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": rls.Status})
		return
	}
	id := path
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rls, ok := s.store.ReleaseGet(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, rls)
}

func (s *Server) handleReleaseAction(w http.ResponseWriter, r *http.Request, id, action string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rls, err := s.store.ReleaseAdvance(id, action)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, rls)
}
