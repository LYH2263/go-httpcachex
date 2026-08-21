package httpcachex

// Close 刷盘并释放资源，并取消后台刷新 goroutine 的生命周期 ctx。
func (c *Cache) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrClosed
	}
	c.closed = true
	// 取消生命周期 ctx：正在进行的后台刷新上游请求不持锁，
	// ctx 取消会经 http.NewRequestWithContext 让 RoundTrip 立即返回，
	// 从而不会阻塞优雅退出。
	if c.lifecycleCancel != nil {
		c.lifecycleCancel()
	}
	if err := c.persistLocked(); err != nil {
		c.mu.Unlock()
		return err
	}
	if c.audit != nil {
		_ = c.audit.Close()
		c.audit = nil
	}
	c.store.Clear()
	c.mu.Unlock()
	return nil
}
