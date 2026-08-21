package httpcachex

import (
	"net/http"
	"time"

	"example.com/httpcachex/internal/clock"
	"example.com/httpcachex/internal/policy"
)

type Options struct {
	Transport   http.RoundTripper
	Clock       clock.Clock
	DefaultTTL  time.Duration
	SWR         time.Duration
	MaxEntries  int
	PersistPath string
	AuditPath   string
	Policy      policy.Policy
}

func (o *Options) normalize() {
	if o.Clock == nil {
		o.Clock = clock.Real{}
	}
	if o.DefaultTTL <= 0 {
		o.DefaultTTL = 30 * time.Second
	}
	if o.SWR <= 0 {
		o.SWR = 10 * time.Second
	}
	if o.MaxEntries <= 0 {
		o.MaxEntries = 1024
	}
	if o.Policy.DefaultTTL == 0 {
		o.Policy = policy.Default()
	}
}
