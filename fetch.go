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
//
// 取消语义：ctx 已取消时立即返回，既不竞争锁等待，也不向上游发起请求，
// 避免慢上游拖死网关线程、客户端断开后仍占用连接。上游拉取期间不持锁，
// 故并发请求各自受自身 ctx 约束，超时熔断有效。
func (c *Cache) FetchContext(ctx context.Context, req Request) (Response, bool, error) {
	// 已取消则尽快返回，不竞争锁、不拉上游。
	if err := ctx.Err(); err != nil {
		return Response{}, false, wrapCancel(err)
	}

	// 查缓存：仅做内存操作，临界区极短。
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return Response{}, false, ErrClosed
	}
	if c.rt == nil {
		c.mu.Unlock()
		return Response{}, false, ErrNoTransport
	}
	if req.URL == "" {
		c.mu.Unlock()
		return Response{}, false, ErrInvalid
	}
	if err := c.pol.WaitFetch(ctx); err != nil {
		c.mu.Unlock()
		return Response{}, false, wrapCancel(err)
	}
	k := key.Build(req.Method, req.URL, req.Headers)
	if e, ok := c.store.Get(k); ok {
		now := c.clk.Now()
		if now.Before(e.FreshUntil) {
			c.metrics.IncHit()
			c.mu.Unlock()
			return toResponse(e), true, nil
		}
		if now.Before(e.StaleUntil) {
			c.metrics.IncStale()
			c.mu.Unlock()
			go c.revalidateAsync(k, req)
			return toResponse(e), true, nil
		}
	}
	c.metrics.IncMiss()
	c.mu.Unlock()

	// 拉上游不持锁，慢上游不会阻塞其它请求；ctx 取消时立即中止。
	if err := ctx.Err(); err != nil {
		return Response{}, false, wrapCancel(err)
	}
	upRaw, err := upstream.Do(ctx, c.rt, upstream.Req{Method: req.Method, URL: req.URL, Headers: req.Headers})
	if err != nil {
		// ctx 取消是根因时按取消上报，不要伪装成上游故障。
		if cerr := ctx.Err(); cerr != nil {
			return Response{}, false, wrapCancel(cerr)
		}
		return Response{}, false, wrapUpstream(err)
	}
	up := Response{Status: upRaw.Status, Headers: upRaw.Headers, Body: upRaw.Body}
	if !c.pol.CanStore(up.Headers) {
		return up, false, nil
	}

	// 回写缓存：重新加锁，期间可能被 Close。
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return Response{}, false, ErrClosed
	}
	e := entryFrom(k, req.URL, key.VaryNames(req.Headers), up, c.clk.Now(), c.opts.DefaultTTL, c.opts.SWR)
	c.store.Put(k, e)
	if err := c.persistLocked(); err != nil {
		c.store.Delete(k)
		return Response{}, false, err
	}
	if c.audit != nil {
		_ = c.audit.Write("store", k, req.URL)
	}
	return toResponse(e), false, nil
}
