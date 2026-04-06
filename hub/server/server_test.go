package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfigCRUDAndHistory(t *testing.T) {
	s := New()
	h := s.Handler()

	body := map[string]any{"scope": "app:user-svc", "name": "database.toml", "content": "[database]\ndsn=\"a\"", "metadata": map[string]any{"hot": true}}
	cfg := ConfigFile{}
	reqJSON(t, h, http.MethodPost, "/api/v1/configs", body, http.StatusCreated, &cfg)
	if cfg.ID == "" || cfg.Version != 1 {
		t.Fatalf("unexpected create result: %+v", cfg)
	}

	upd := map[string]any{"scope": "app:user-svc", "name": "database.toml", "content": "[database]\ndsn=\"b\"", "metadata": map[string]any{"hot": true}}
	reqJSON(t, h, http.MethodPut, "/api/v1/configs/"+cfg.ID, upd, http.StatusOK, &cfg)
	if cfg.Version != 2 {
		t.Fatalf("version should be 2 got %d", cfg.Version)
	}

	var history []ConfigHistory
	reqJSON(t, h, http.MethodGet, "/api/v1/configs/"+cfg.ID+"/history", nil, http.StatusOK, &history)
	if len(history) != 2 {
		t.Fatalf("history len want 2 got %d", len(history))
	}
}

func TestReleaseStateMachine(t *testing.T) {
	s := New()
	h := s.Handler()

	var rel Release
	reqJSON(t, h, http.MethodPost, "/api/v1/releases", map[string]any{"environment": "prod", "config_ids": []string{"cfg-1"}}, http.StatusCreated, &rel)
	reqJSON(t, h, http.MethodPost, "/api/v1/releases/"+rel.ID+"/approve", nil, http.StatusOK, &rel)
	if rel.Status != ReleaseApproved {
		t.Fatalf("expected approved got %s", rel.Status)
	}
	reqJSON(t, h, http.MethodPost, "/api/v1/releases/"+rel.ID+"/publish", nil, http.StatusOK, &rel)
	if rel.Status != ReleasePublished {
		t.Fatalf("expected published got %s", rel.Status)
	}
	reqJSON(t, h, http.MethodPost, "/api/v1/releases/"+rel.ID+"/rollback", nil, http.StatusOK, &rel)
	if rel.Status != ReleaseRolledBack {
		t.Fatalf("expected rolled_back got %s", rel.Status)
	}
}

func TestEnvironmentAndClusterAPI(t *testing.T) {
	s := New()
	h := s.Handler()

	reqJSON(t, h, http.MethodPost, "/api/v1/environments", map[string]any{"name": "prod"}, http.StatusCreated, nil)
	reqJSON(t, h, http.MethodPost, "/api/v1/clusters", map[string]any{"name": "cn-east-1", "env": "prod"}, http.StatusCreated, nil)

	status := map[string]string{}
	reqJSON(t, h, http.MethodGet, "/api/v1/clusters/cn-east-1/status", nil, http.StatusOK, &status)
	if status["status"] != "up" {
		t.Fatalf("expected up got %+v", status)
	}
}

func reqJSON(t *testing.T, h http.Handler, method, path string, in any, wantCode int, out any) {
	t.Helper()
	var body bytes.Buffer
	if in != nil {
		if err := json.NewEncoder(&body).Encode(in); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &body)
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != wantCode {
		t.Fatalf("%s %s status want %d got %d body=%s", method, path, wantCode, w.Code, w.Body.String())
	}
	if out != nil {
		if err := json.NewDecoder(w.Body).Decode(out); err != nil {
			t.Fatal(err)
		}
	}
}
