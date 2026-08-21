package httpcachex

// Snapshot 导出缓存条目视图。返回的切片与字段均为独立拷贝，
// 调用方修改 EntryView（含 Vary 脱敏）不会写穿缓存内部状态。
func (c *Cache) Snapshot() []EntryView {
	c.mu.Lock()
	defer c.mu.Unlock()
	raw := c.store.List()
	out := make([]EntryView, 0, len(raw))
	for _, e := range raw {
		out = append(out, EntryView{
			Key:        e.Key,
			URL:        e.URL,
			Vary:       append([]string(nil), e.Vary...),
			Status:     e.Status,
			Hits:       e.Hits,
			BodyLen:    len(e.Body),
			FreshUntil: e.FreshUntil,
			StaleUntil: e.StaleUntil,
		})
	}
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
