package httpcachex

import (
	"context"

	"example.com/httpcachex/internal/upstream"
)

func (c *Cache) revalidateAsync(k string, req Request) {
	_ = c.Revalidate(context.Background(), k, req)
}

// Revalidate 后台条件刷新，须尊重 ctx。
func (c *Cache) Revalidate(ctx context.Context, k string, req Request) error {

	if err := c.pol.WaitRevalidate(ctx); err != nil {
		return wrapCancel(err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrClosed
	}
	upRaw, err := upstream.Do(ctx, c.rt, upstream.Req{Method: req.Method, URL: req.URL, Headers: req.Headers})
	if err != nil {
		return wrapUpstream(err)
	}
	up := Response{Status: upRaw.Status, Headers: upRaw.Headers, Body: upRaw.Body}
	e := entryFrom(k, req.URL, nil, up, c.clk.Now(), c.opts.DefaultTTL, c.opts.SWR)
	c.store.Put(k, e)
	c.metrics.IncRevalidate()
	return c.persistLocked()
}
