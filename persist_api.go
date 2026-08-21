package httpcachex

import "example.com/httpcachex/internal/persist"

// persistLocked 将当前内存索引落盘。
//
// 调用方负责保证：调用本方法前，已先对内存做了改动；若本方法返回错误，
// 调用方必须回滚相应内存改动，绝不可让"内存已改 + 磁盘未落"的半成功状态留存。
func (c *Cache) persistLocked() error {
	if c.persist == nil {
		return nil
	}
	snap := persist.Snapshot{Entries: c.store.List()}
	if err := c.persist.Save(snap); err != nil {
		return wrapPersist(err)
	}
	return nil
}

func (c *Cache) LoadPersist() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.persist == nil {
		return nil
	}
	snap, err := c.persist.Load()
	if err != nil {
		return wrapPersist(err)
	}
	c.store.Replace(snap.Entries)
	return nil
}
