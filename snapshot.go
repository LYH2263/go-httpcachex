package httpcachex

import "example.com/httpcachex/internal/key"

// Snapshot 导出缓存条目视图。
func (c *Cache) Snapshot() []EntryView {
	c.mu.Lock()
	defer c.mu.Unlock()
	raw := c.store.List()
	out := make([]EntryView, 0, len(raw))
	for _, e := range raw {
		vary := append([]string(nil), e.Vary...)
		out = append(out, EntryView{
			Key:        e.Key,
			URL:        e.URL,
			Vary:       vary,
			Status:     e.Status,
			Hits:       e.Hits,
			BodyLen:    len(e.Body),
			FreshUntil: e.FreshUntil,
			StaleUntil: e.StaleUntil,
		})
	}
	_ = key.Build
	return out
}

// Purge 删除键。
func (c *Cache) Purge(k string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrClosed
	}
	// 记住被删条目，落盘失败时恢复，避免内存删了磁盘没删的半成功状态。
	// 用 Peek 避免触发命中计数。
	prev, ok := c.store.Peek(k)
	c.store.Delete(k)
	if err := c.persistLocked(); err != nil {
		if ok {
			c.store.Put(k, prev)
		}
		return err
	}
	return nil
}

// Stats 计数。
func (c *Cache) Stats() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.metrics.Snapshot()
}
