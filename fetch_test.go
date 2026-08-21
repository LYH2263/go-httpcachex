package httpcachex

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"example.com/httpcachex/internal/clock"
)

// rtFunc 把一个函数适配成 http.RoundTripper。
type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// ok200 返回一条可被缓存的 200 响应（Cache-Control: max-age=30）。
func ok200() *http.Response {
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Cache-Control": []string{"max-age=30"}, "Content-Type": []string{"text/plain"}},
		Body:       http.NoBody,
	}
}

// newCache 构建一个用假时钟、假 transport、且 PersistPath 必然落盘失败的 Cache。
// 失败方式：把 PersistPath 的父目录设成一个普通文件，使 persist.Save 里的
// os.MkdirAll(filepath.Dir(path)) 返回 ENOTDIR。这比依赖目录权限更可靠
// （Windows 不按 Unix 权限位阻止目录写入）。
func newCache(t *testing.T) *Cache {
	t.Helper()
	// 先建一个普通文件，让 PersistPath 的"父目录"是文件而非目录。
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	c, err := New(Options{
		Transport:   rtFunc(func(*http.Request) (*http.Response, error) { return ok200(), nil }),
		Clock:       &clock.Fake{T: time.Unix(1_700_000_000, 0)},
		PersistPath: filepath.Join(blocker, "cache.json"),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// TestFetch_PersistFailRollsBackMemory 复现"PersistPath 配错后 Fetch 返回
// 错误但内存里仍残留条目"的 bug：落盘失败时必须回滚内存索引。
func TestFetch_PersistFailRollsBackMemory(t *testing.T) {
	c := newCache(t)
	defer c.Close()

	resp, stored, err := c.Fetch(Request{Method: "GET", URL: "http://example.com/x"})
	if err == nil {
		t.Fatalf("want persist error, got nil (stored=%v)", stored)
	}
	if !errors.Is(err, ErrPersist) {
		t.Fatalf("want ErrPersist, got %v", err)
	}
	if stored {
		t.Fatalf("should not report stored on persist failure")
	}
	_ = resp

	// 关键断言：落盘失败后内存里不该残留半成功条目，
	// 否则同进程会脏命中、重启又消失。
	if got := c.Snapshot(); len(got) != 0 {
		t.Fatalf("memory not rolled back: %d entries remain (%+v)", len(got), got)
	}
}

// TestFetch_PersistFailNoDirtyHit 验证落盘失败后同进程再次 Fetch 不会脏命中：
// 必须重新请求上游、而非返回残留的旧响应。
func TestFetch_PersistFailNoDirtyHit(t *testing.T) {
	c := newCache(t)
	defer c.Close()

	if _, _, err := c.Fetch(Request{Method: "GET", URL: "http://example.com/x"}); err == nil {
		t.Fatalf("first Fetch: want persist error, got nil")
	}

	// 第二次 Fetch：若内存回滚正确，应当 miss 并再次尝试落盘 → 再次失败，
	// 而非命中残留条目（命中会返回 stored=true 且无错误）。
	resp, stored, err := c.Fetch(Request{Method: "GET", URL: "http://example.com/x"})
	if err == nil {
		t.Fatalf("second Fetch: want persist error again, got nil (stored=%v)", stored)
	}
	if stored {
		t.Fatalf("second Fetch: dirty hit from rolled-back entry (resp=%+v)", resp)
	}
}

// seedCache 建一个能正常落盘的 Cache，种入一条可命中的缓存，返回它。
// 调用方可通过 breakPersist 让后续 persistLocked 必然失败。
func seedCache(t *testing.T) (*Cache, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")
	c, err := New(Options{
		Transport:   rtFunc(func(*http.Request) (*http.Response, error) { return ok200(), nil }),
		Clock:       &clock.Fake{T: time.Unix(1_700_000_000, 0)},
		PersistPath: path,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, _, err := c.Fetch(Request{Method: "GET", URL: "http://example.com/x"}); err != nil {
		t.Fatalf("seed Fetch: %v", err)
	}
	// breakPersist：把持久化目录替换成一个普通文件，
	// 使后续 Save 里的 os.MkdirAll(filepath.Dir(path)) 返回 ENOTDIR。
	breakFn := func() {
		_ = os.RemoveAll(dir)
		_ = os.WriteFile(dir, []byte("x"), 0o644)
	}
	return c, breakFn
}

// TestPurge_PersistFailRollsBackMemory 验证 Purge 落盘失败时把被删条目恢复，
// 不留下"内存已删、磁盘未删"的半成功状态。
func TestPurge_PersistFailRollsBackMemory(t *testing.T) {
	c, breakPersist := seedCache(t)
	defer c.Close()

	before := len(c.Snapshot())
	if before != 1 {
		t.Fatalf("seed: want 1 entry, got %d", before)
	}

	breakPersist()

	// 用 key 取出种入条目的 key（Snapshot 顺序即插入顺序）。
	k := c.Snapshot()[0].Key
	if err := c.Purge(k); err == nil {
		t.Fatalf("Purge: want persist error, got nil")
	}
	// 落盘失败后条目应仍在内存里（被回滚恢复）。
	if got := c.Snapshot(); len(got) != 1 {
		t.Fatalf("Purge not rolled back: want 1 entry, got %d", len(got))
	}
}

// TestRevalidate_PersistFailRollsBackMemory 验证 Revalidate 落盘失败时
// 回滚为旧条目，而非留下一半成功的新条目。
func TestRevalidate_PersistFailRollsBackMemory(t *testing.T) {
	c, breakPersist := seedCache(t)
	defer c.Close()

	breakPersist()

	k := c.Snapshot()[0].Key
	req := Request{Method: "GET", URL: "http://example.com/x"}
	if err := c.Revalidate(context.Background(), k, req); err == nil {
		t.Fatalf("Revalidate: want persist error, got nil")
	}
	// 落盘失败后内存应仍持有旧条目（回滚为新值后再回滚为旧值）。
	if got := c.Snapshot(); len(got) != 1 || got[0].Key != k {
		t.Fatalf("Revalidate not rolled back: got %+v", got)
	}
}
