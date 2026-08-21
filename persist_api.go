package httpcachex

import "example.com/httpcachex/internal/persist"

func (c *Cache) persistLocked() error {
	if c.persist == nil {
		return nil
	}
	snap := persist.Snapshot{Entries: c.store.List()} // Close 须先 persistLocked 再 Clear，否则空快照写盘覆盖有效数据
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
