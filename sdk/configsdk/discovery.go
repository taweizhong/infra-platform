package configsdk

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"
)

type RegisterOption func(*registerOptions)

type registerOptions struct {
	healthPath string
	interval   time.Duration
	ttl        int
	metadata   map[string]string
}

func defaultRegisterOptions() registerOptions {
	return registerOptions{interval: 10 * time.Second, ttl: 30, metadata: map[string]string{}}
}

func WithHealthCheck(path string, interval time.Duration) RegisterOption {
	return func(o *registerOptions) {
		o.healthPath = path
		o.interval = interval
	}
}

func WithDiscoveryMetadata(m map[string]string) RegisterOption {
	return func(o *registerOptions) {
		o.metadata = m
	}
}

type Registration struct {
	ID         string
	service    string
	client     *Client
	cancelFunc context.CancelFunc
	once       sync.Once
}

func (r *Registration) Deregister() error {
	var err error
	r.once.Do(func() {
		if r.cancelFunc != nil {
			r.cancelFunc()
		}
		err = r.client.tp.delete(context.Background(), fmt.Sprintf("/api/v1/services/%s", r.ID), nil)
	})
	return err
}

type discoveryModule struct {
	client *Client
	mu     sync.Mutex
	regs   []*Registration
}

func newDiscoveryModule(c *Client) *discoveryModule { return &discoveryModule{client: c} }

func (d *discoveryModule) Close() {
	d.mu.Lock()
	regs := append([]*Registration(nil), d.regs...)
	d.mu.Unlock()
	for _, r := range regs {
		_ = r.Deregister()
	}
}

func (d *discoveryModule) Register(name, addr string, opts ...RegisterOption) (*Registration, error) {
	ro := defaultRegisterOptions()
	for _, opt := range opts {
		opt(&ro)
	}
	inst := ServiceInstance{Name: name, Address: addr, TTL: ro.ttl, Metadata: ro.metadata}
	resp := ServiceInstance{}
	if err := d.client.tp.post(context.Background(), "/api/v1/services", inst, &resp); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	reg := &Registration{ID: resp.ID, service: name, client: d.client, cancelFunc: cancel}
	go d.heartbeatLoop(ctx, reg.ID, ro.interval)
	d.mu.Lock()
	d.regs = append(d.regs, reg)
	d.mu.Unlock()
	return reg, nil
}

func (d *discoveryModule) Discover(name string) ([]ServiceInstance, error) {
	q := url.Values{}
	q.Set("name", name)
	var out []ServiceInstance
	err := d.client.tp.get(context.Background(), "/api/v1/services", q, &out)
	return out, err
}

func (d *discoveryModule) Watch(ctx context.Context, name string) (<-chan DiscoveryWatchEvent, error) {
	ch := make(chan DiscoveryWatchEvent, 16)
	go func() {
		defer close(ch)
		version := int64(0)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			request := map[string]any{"service": name, "from_version": version}
			resp := struct {
				Changed bool                  `json:"changed"`
				Events  []DiscoveryWatchEvent `json:"events"`
			}{}
			err := d.client.tp.post(ctx, fmt.Sprintf("/api/v1/services/watch?timeout=%s", d.client.opts.watchTimeout.String()), request, &resp)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(d.client.opts.retryInterval):
					continue
				}
			}
			if !resp.Changed {
				continue
			}
			for _, ev := range resp.Events {
				if ev.Version > version {
					version = ev.Version
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

func (d *discoveryModule) heartbeatLoop(ctx context.Context, id string, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_ = d.client.tp.put(context.Background(), fmt.Sprintf("/api/v1/services/%s/heartbeat", id), map[string]any{}, nil)
		}
	}
}
