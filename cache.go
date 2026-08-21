package httpcachex

import (
	"context"
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

	// lifecycleCtx 在 Close 时取消，用于停止后台刷新 goroutine。
	lifecycleCtx    context.Context
	lifecycleCancel context.CancelFunc
}

func New(opts Options) (*Cache, error) {
	opts.normalize()
	ctx, cancel := context.WithCancel(context.Background())
	c := &Cache{
		opts:            opts,
		store:           store.New(opts.MaxEntries),
		clk:             opts.Clock,
		pol:             opts.Policy,
		rt:              opts.Transport,
		metrics:         metrics.New(),
		lifecycleCtx:    ctx,
		lifecycleCancel: cancel,
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
