package httpcachex

import (
	"net/http"
	"sync"

	"example.com/httpcachex/internal/audit"
	"example.com/httpcachex/internal/clock"
	"example.com/httpcachex/internal/metrics"
	"example.com/httpcachex/internal/persist"
	"example.com/httpcachex/internal/policy"
	"example.com/httpcachex/internal/store"
)

type Cache struct {
	mu      sync.Mutex
	opts    Options
	store   *store.Mem
	clk     clock.Clock
	pol     policy.Policy
	rt      http.RoundTripper
	persist *persist.Store
	audit   *audit.Logger
	metrics *metrics.Registry
	closed  bool
}

func New(opts Options) (*Cache, error) {
	opts.normalize()
	c := &Cache{
		opts:    opts,
		store:   store.New(opts.MaxEntries),
		clk:     opts.Clock,
		pol:     opts.Policy,
		rt:      opts.Transport,
		metrics: metrics.New(),
	}
	if opts.PersistPath != "" {
		c.persist = persist.New(opts.PersistPath)
	}
	if opts.AuditPath != "" {
		lg, err := audit.Open(opts.AuditPath)
		if err != nil {
			return nil, err
		}
		c.audit = lg
	}
	return c, nil
}
