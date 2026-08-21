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
	c.store.Delete(k)
	return c.persistLocked()
}

// Stats 计数。
func (c *Cache) Stats() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.metrics.Snapshot()
}
