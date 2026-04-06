package configsdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type transport struct {
	baseURL    string
	httpClient *http.Client
}

func newTransport(baseURL string, timeout time.Duration, client *http.Client) *transport {
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	return &transport{baseURL: strings.TrimRight(baseURL, "/"), httpClient: client}
}

func (t *transport) get(ctx context.Context, p string, query url.Values, out any) error {
	u, err := url.Parse(t.baseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, p)
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	return t.do(req, out)
}

func (t *transport) put(ctx context.Context, p string, body any, out any) error {
	return t.withBody(ctx, http.MethodPut, p, body, out)
}

func (t *transport) post(ctx context.Context, p string, body any, out any) error {
	return t.withBody(ctx, http.MethodPost, p, body, out)
}

func (t *transport) delete(ctx context.Context, p string, out any) error {
	u, err := url.Parse(t.baseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, p)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
	if err != nil {
		return err
	}
	return t.do(req, out)
}

func (t *transport) withBody(ctx context.Context, method, p string, body any, out any) error {
	u, err := url.Parse(t.baseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, p)

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return t.do(req, out)
}

func (t *transport) do(req *http.Request, out any) error {
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed: %s: %s", resp.Status, string(raw))
	}
	if out == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
