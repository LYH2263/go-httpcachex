package httpcachex

// Close 刷盘并释放资源。
func (c *Cache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return ErrClosed
	}
	c.closed = true
	if err := c.persistLocked(); err != nil {
		return err
	}
	if c.audit != nil {
		_ = c.audit.Close()
		c.audit = nil
	}
	c.store = nil
	return nil
}
