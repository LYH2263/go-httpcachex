package httpcachex

import (
	"context"

	"example.com/httpcachex/internal/upstream"
)

func (c *Cache) revalidateAsync(k string, req Request) {
	// 继承 Cache 生命周期 ctx：Close 时被取消，让后台刷新尽快退出。
	_ = c.Revalidate(c.lifecycleCtx, k, req)
}

// Revalidate 后台条件刷新，须尊重 ctx。
//
// 关键点：上游网络请求（upstream.Do）不持有 c.mu，这样
//   - ctx 取消能立刻让 RoundTrip 中止返回，无需等锁；
//   - Close 能随时拿到锁并取消生命周期 ctx，触发在途刷新退出，优雅停机不阻塞。
func (c *Cache) Revalidate(ctx context.Context, k string, req Request) error {
	if err := c.pol.WaitRevalidate(ctx); err != nil {
		return wrapCancel(err)
	}
	if err := ctx.Err(); err != nil {
		return wrapCancel(err)
	}
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClosed
	}
	c.mu.Unlock()

	// 不持锁发起上游请求；ctx 取消会经 http.NewRequestWithContext 传入 RoundTrip，
	// 中途取消即可立即返回，不会继续向上游打满请求。
	upRaw, err := upstream.Do(ctx, c.rt, upstream.Req{Method: req.Method, URL: req.URL, Headers: req.Headers})
	if err != nil {
		return wrapUpstream(err)
	}
	up := Response{Status: upRaw.Status, Headers: upRaw.Headers, Body: upRaw.Body}

	c.mu.Lock()
	defer c.mu.Unlock()
	// 二次检查：上游返回期间可能已 Close，避免向已清空的 store 写入。
	if c.closed {
		return ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return wrapCancel(err)
	}
	e := entryFrom(k, req.URL, nil, up, c.clk.Now(), c.opts.DefaultTTL, c.opts.SWR)
	c.store.Put(k, e)
	c.metrics.IncRevalidate()
	return c.persistLocked()
}
