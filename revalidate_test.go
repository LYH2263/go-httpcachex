package httpcachex

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/httpcachex/internal/clock"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newTestCache(t *testing.T, rt http.RoundTripper) *Cache {
	t.Helper()
	c, err := New(Options{Transport: rt, Clock: &clock.Fake{T: time.Unix(0, 0)}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// TestRevalidateRespectsCanceledCtx 传入已取消的 ctx 时，Revalidate 必须立即返回，
// 且绝不能向上游打请求。
func TestRevalidateRespectsCanceledCtx(t *testing.T) {
	var called int32
	rt := rtFunc(func(r *http.Request) (*http.Response, error) {
		atomic.AddInt32(&called, 1)
		return &http.Response{StatusCode: 200, Body: http.NoBody, Header: http.Header{}, Request: r}, nil
	})
	c := newTestCache(t, rt)
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 先取消

	done := make(chan error, 1)
	go func() { done <- c.Revalidate(ctx, "k", Request{Method: "GET", URL: "http://x"}) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("期望返回取消错误，得到 nil")
		}
		if !errors.Is(err, ErrCanceled) {
			t.Fatalf("期望 ErrCanceled，得到 %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Revalidate 未尊重已取消的 ctx，2s 后仍未返回")
	}
	if got := atomic.LoadInt32(&called); got != 0 {
		t.Fatalf("取消后不应调用上游，实际调用 %d 次", got)
	}
}

// TestRevalidatePropagatesCancelMidflight 取消 ctx 中途到达时，正在进行的上游请求必须尽快返回。
func TestRevalidatePropagatesCancelMidflight(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	rt := rtFunc(func(r *http.Request) (*http.Response, error) {
		close(started)
		select {
		case <-release:
			return nil, errors.New("should not reach here")
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
	})
	c := newTestCache(t, rt)
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.Revalidate(ctx, "k", Request{Method: "GET", URL: "http://x"}) }()

	<-started // 等到上游请求真正开始
	cancel()  // 中途取消

	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("期望返回错误，得到 nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("中途取消后 Revalidate 2s 内未返回，ctx 未被尊重")
	}
	close(release)
}

// TestRevalidateAsyncStopsOnClose 后台刷新 goroutine 绑定 Cache 生命周期：
// Close 必须取消正在进行的后台刷新上游请求。
func TestRevalidateAsyncStopsOnClose(t *testing.T) {
	var mu sync.Mutex
	busy := make(map[string]chan struct{})
	rt := rtFunc(func(r *http.Request) (*http.Response, error) {
		key := r.URL.String()
		mu.Lock()
		ch, ok := busy[key]
		if !ok {
			ch = make(chan struct{})
			busy[key] = ch
		}
		mu.Unlock()
		select {
		case <-ch:
			return nil, errors.New("released manually")
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
	})
	c := newTestCache(t, rt)

	// 直接构造一个 stuck 的后台刷新。
	go c.revalidateAsync("k", Request{Method: "GET", URL: "http://stuck"})
	// 等到上游请求真正挂起。
	var ch chan struct{}
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	for {
		mu.Lock()
		ch = busy["http://stuck"]
		mu.Unlock()
		if ch != nil {
			break
		}
		select {
		case <-timer.C:
			t.Fatal("上游请求未在 2s 内启动")
		default:
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Close 应能取消后台刷新并返回，而不是被它无限阻塞。
	done := make(chan struct{})
	go func() {
		_ = c.Close()
		close(done)
	}()
	select {
	case <-done:
		// 期望：Close 因后台刷新被取消而及时返回。
	case <-time.After(5 * time.Second):
		t.Fatal("Close 5s 未返回，后台刷新阻塞了优雅退出")
	}
}
