package server

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func (s *Server) handleConfigs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		scope := r.URL.Query().Get("scope")
		writeJSON(w, http.StatusOK, s.store.ListConfigs(scope))
	case http.MethodPost:
		var req ConfigFile
		if err := readJSON(r, &req); err != nil || req.Scope == "" || req.Name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		writeJSON(w, http.StatusCreated, s.store.CreateConfig(req))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleConfigByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/configs/")
	if strings.HasSuffix(path, "/history") {
		id := strings.TrimSuffix(path, "/history")
		writeJSON(w, http.StatusOK, s.store.ConfigHistory(id))
		return
	}
	if strings.Contains(path, "/diff") {
		id := strings.TrimSuffix(path, "/diff")
		history := s.store.ConfigHistory(strings.TrimSuffix(id, "/"))
		v1, v2 := r.URL.Query().Get("v1"), r.URL.Query().Get("v2")
		var c1, c2 string
		for _, h := range history {
			if toString(h.Version) == v1 {
				c1 = h.Content
			}
			if toString(h.Version) == v2 {
				c2 = h.Content
			}
		}
		writeJSON(w, http.StatusOK, map[string]string{"v1": c1, "v2": c2})
		return
	}
	id := path
	switch r.Method {
	case http.MethodGet:
		cfg, ok := s.store.GetConfig(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	case http.MethodPut:
		var req ConfigFile
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		cfg, err := s.store.UpdateConfig(id, req)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	case http.MethodDelete:
		s.store.DeleteConfig(id)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleConfigPreview(w http.ResponseWriter, r *http.Request) {
	app := r.URL.Query().Get("app")
	env := r.URL.Query().Get("env")
	cluster := r.URL.Query().Get("cluster")

	scopes := []string{"global", "env:" + env, "cluster:" + cluster, "app:" + app}
	merged := map[string][]ConfigFile{}
	for _, sc := range scopes {
		files := s.store.ListConfigs(sc)
		if len(files) == 0 {
			continue
		}
		merged[sc] = files
	}
	ordered := make([]string, 0, len(merged))
	for k := range merged {
		ordered = append(ordered, k)
	}
	sort.Strings(ordered)
	writeJSON(w, http.StatusOK, map[string]any{"ordered_scopes": ordered, "scopes": merged})
}

func toString(v int64) string {
	return strconv.FormatInt(v, 10)
}
