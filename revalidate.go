package httpcachex

import (
	"context"

	"example.com/httpcachex/internal/upstream"
)

func (c *Cache) revalidateAsync(k string, req Request) {
	_ = c.Revalidate(context.Background(), k, req)
}

// Revalidate 后台条件刷新，须尊重 ctx。
//
// 拉上游不持锁，避免后台慢刷新阻塞前台缓存命中。
func (c *Cache) Revalidate(ctx context.Context, k string, req Request) error {
	if err := ctx.Err(); err != nil {
		return wrapCancel(err)
	}
	if err := c.pol.WaitRevalidate(ctx); err != nil {
		return wrapCancel(err)
	}

	// 上游拉取在锁外进行：慢上游不会拖住前台请求。
	if err := ctx.Err(); err != nil {
		return wrapCancel(err)
	}
	upRaw, err := upstream.Do(ctx, c.rt, upstream.Req{Method: req.Method, URL: req.URL, Headers: req.Headers})
	if err != nil {
		if cerr := ctx.Err(); cerr != nil {
			return wrapCancel(cerr)
		}
		return wrapUpstream(err)
	}
	up := Response{Status: upRaw.Status, Headers: upRaw.Headers, Body: upRaw.Body}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrClosed
	}
	e := entryFrom(k, req.URL, nil, up, c.clk.Now(), c.opts.DefaultTTL, c.opts.SWR)
	c.store.Put(k, e)
	c.metrics.IncRevalidate()
	return c.persistLocked()
}
