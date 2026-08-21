package httpcachex

// Close 刷盘并释放资源。
//
// 顺序很关键：必须先 persistLocked 把当前内存快照刷盘，再 Clear 清空内存索引。
// 若先 Clear 再 persistLocked，store.List() 返回空，会把空快照写盘，覆盖有效数据；
// 重启后 LoadPersist 恢复的是空缓存，热点全部打穿上游，启动瞬间流量打爆源站。
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
	c.store.Clear()
	if c.audit != nil {
		_ = c.audit.Close()
		c.audit = nil
	}
	return nil
}
