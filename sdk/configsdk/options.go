package configsdk

import (
	"errors"
	"net/http"
	"time"
)

type Decryptor interface {
	Decrypt(content string) (string, error)
}

type Option func(*options)

type options struct {
	agentURL          string
	app               string
	env               string
	cluster           string
	instance          string
	enableWatch       bool
	requestTimeout    time.Duration
	watchTimeout      time.Duration
	retryInterval     time.Duration
	httpClient        *http.Client
	decryptor         Decryptor
	fileCachePath     string
	heartbeatInterval time.Duration
}

func defaultOptions() options {
	return options{
		requestTimeout:    5 * time.Second,
		watchTimeout:      30 * time.Second,
		retryInterval:     2 * time.Second,
		heartbeatInterval: 10 * time.Second,
	}
}

func WithAgent(agent string) Option             { return func(o *options) { o.agentURL = agent } }
func WithApp(app string) Option                 { return func(o *options) { o.app = app } }
func WithEnv(env string) Option                 { return func(o *options) { o.env = env } }
func WithCluster(cluster string) Option         { return func(o *options) { o.cluster = cluster } }
func WithInstance(instance string) Option       { return func(o *options) { o.instance = instance } }
func WithWatch(enable bool) Option              { return func(o *options) { o.enableWatch = enable } }
func WithRequestTimeout(d time.Duration) Option { return func(o *options) { o.requestTimeout = d } }
func WithWatchTimeout(d time.Duration) Option   { return func(o *options) { o.watchTimeout = d } }
func WithRetryInterval(d time.Duration) Option  { return func(o *options) { o.retryInterval = d } }
func WithHTTPClient(c *http.Client) Option      { return func(o *options) { o.httpClient = c } }
func WithDecryptor(d Decryptor) Option          { return func(o *options) { o.decryptor = d } }
func WithFileCache(path string) Option          { return func(o *options) { o.fileCachePath = path } }
func WithHeartbeatInterval(d time.Duration) Option {
	return func(o *options) { o.heartbeatInterval = d }
}

func (o options) validate() error {
	if o.agentURL == "" {
		return errors.New("agent endpoint is required")
	}
	if o.app == "" {
		return errors.New("app is required")
	}
	if o.env == "" {
		return errors.New("env is required")
	}
	if o.cluster == "" {
		return errors.New("cluster is required")
	}
	return nil
}
