package httpc

import (
	"context"
	"net/http"
	"time"

	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
)

type (
	Service interface {
		Call(ctx context.Context, req *http.Request) (*http.Response, error)
	}

	Params struct {
		Config config.Service
	}

	ObserverRoundTripper struct {
		userAgent string
		tripper   http.RoundTripper
	}

	service struct {
		doer *http.Client
	}

	Option func(*http.Transport)

	httpcConfig struct {
		UserAgent string        `yaml:"userAgent"`
		Timeout   time.Duration `yaml:"timeout"`
	}
)

const (
	_cfgKey = "httpc"
)

// New inits a new http service
func New(p Params, options ...Option) (Service, error) {
	var cfg httpcConfig
	if err := p.Config.Populate(_cfgKey, &cfg); err != nil {
		return nil, err
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	for _, o := range options {
		o(t)
	}

	rt := ObserverRoundTripper{
		userAgent: cfg.UserAgent,
		tripper:   t,
	}

	return &service{
		doer: &http.Client{Transport: rt, Timeout: cfg.Timeout},
	}, nil
}

// WithMaxIdleConns returns an option which sets the idle conns per host
func WithMaxIdleConns(conns int) Option {
	return func(t *http.Transport) {
		t.MaxIdleConns = conns
	}
}

// WithMaxConnsPerHost returns an option which sets the max conns per host
func WithMaxConnsPerHost(conns int) Option {
	return func(t *http.Transport) {
		t.MaxConnsPerHost = conns
	}
}

// RoundTrip adds logging and statsCollector to http requests
func (t ObserverRoundTripper) RoundTrip(req *http.Request) (res *http.Response, err error) {
	if req.Header.Get("user-agent") == "" {
		req.Header.Add("user-agent", t.userAgent)
	}

	res, err = t.tripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return res, err
}

func (c *service) Call(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.doer.Do(req)
}
