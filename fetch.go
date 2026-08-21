package httpcachex

import (
	"context"

	"example.com/httpcachex/internal/key"
	"example.com/httpcachex/internal/upstream"
)

// Fetch 按缓存策略取响应。
func (c *Cache) Fetch(req Request) (Response, bool, error) {
	return c.FetchContext(context.Background(), req)
}

// FetchContext 可取消取缓存。
func (c *Cache) FetchContext(ctx context.Context, req Request) (Response, bool, error) {
	if err := ctx.Err(); err != nil {
		return Response{}, false, wrapCancel(err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return Response{}, false, ErrClosed
	}
	if c.rt == nil {
		return Response{}, false, ErrNoTransport
	}
	if req.URL == "" {
		return Response{}, false, ErrInvalid
	}
	if err := c.pol.WaitFetch(ctx); err != nil {
		return Response{}, false, wrapCancel(err)
	}
	k := key.Build(req.Method, req.URL, req.Headers)
	if e, ok := c.store.Get(k); ok {
		now := c.clk.Now()
		if now.Before(e.FreshUntil) {
			c.metrics.IncHit()
			return toResponse(e), true, nil
		}
		if now.Before(e.StaleUntil) {
			c.metrics.IncStale()
			go c.revalidateAsync(k, req)
			return toResponse(e), true, nil
		}
	}
	c.metrics.IncMiss()
	upRaw, err := upstream.Do(ctx, c.rt, upstream.Req{Method: req.Method, URL: req.URL, Headers: req.Headers})
	if err != nil {
		return Response{}, false, wrapUpstream(err)
	}
	up := Response{Status: upRaw.Status, Headers: upRaw.Headers, Body: upRaw.Body}
	if !c.pol.CanStore(up.Headers) {
		return up, false, nil
	}
	e := entryFrom(k, req.URL, key.VaryNames(req.Headers), up, c.clk.Now(), c.opts.DefaultTTL, c.opts.SWR)
	c.store.Put(k, e)
	if err := c.persistLocked(); err != nil {

		return Response{}, false, err
	}
	if c.audit != nil {
		_ = c.audit.Write("store", k, req.URL)
	}
	return toResponse(e), false, nil
}
