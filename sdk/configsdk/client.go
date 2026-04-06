package configsdk

import (
	"context"
	"net/url"
)

type Client struct {
	opts      options
	tp        *transport
	configMod *configModule
	kvMod     *kvModule
	discMod   *discoveryModule
}

func New(opts ...Option) (*Client, error) {
	cfg := defaultOptions()
	for _, opt := range opts {
		opt(&cfg)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	c := &Client{opts: cfg}
	c.tp = newTransport(cfg.agentURL, cfg.requestTimeout, cfg.httpClient)
	c.configMod = newConfigModule(c)
	c.kvMod = newKVModule(c)
	c.discMod = newDiscoveryModule(c)
	if err := c.configMod.Init(context.Background()); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) Close() {
	c.configMod.Close()
	c.discMod.Close()
}

func (c *Client) Config() *configModule       { return c.configMod }
func (c *Client) KV() *kvModule               { return c.kvMod }
func (c *Client) Discovery() *discoveryModule { return c.discMod }

func mapToQuery(m map[string]string) url.Values {
	q := url.Values{}
	for k, v := range m {
		if v != "" {
			q.Set(k, v)
		}
	}
	return q
}
