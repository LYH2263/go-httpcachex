package policy

import (
	"context"
	"testing"
	"time"
)

// TestWaitRevalidateRespectsCancel WaitRevalidate 必须在 ctx 已取消时返回其错误，
// 而不是像旧行为那样忽略 ctx 直接返回 nil。
func TestWaitRevalidateRespectsCancel(t *testing.T) {
	p := Default()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)
	go func() { done <- p.WaitRevalidate(ctx) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("期望 WaitRevalidate 对已取消 ctx 返回错误，得到 nil（旧 bug：忽略 ctx）")
		}
		if err != context.Canceled {
			t.Fatalf("期望 context.Canceled，得到 %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("WaitRevalidate 2s 未返回，ctx 被忽略")
	}
}

// TestWaitRevalidateAllowsActive 未取消的 ctx 应立即放行，不阻塞。
func TestWaitRevalidateAllowsActive(t *testing.T) {
	p := Default()
	if err := p.WaitRevalidate(context.Background()); err != nil {
		t.Fatalf("未取消 ctx 应返回 nil，得到 %v", err)
	}
}
