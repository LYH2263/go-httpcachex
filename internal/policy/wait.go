package policy

import "context"

// WaitFetch Fetch 前检查取消。已取消则立即返回，避免拉上游。
func (p Policy) WaitFetch(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// WaitRevalidate 后台刷新前检查取消。
func (p Policy) WaitRevalidate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
