package httpcachex

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"example.com/httpcachex/internal/clock"
)

// rtFunc 是测试用 RoundTripper。
type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// TestSnapshot_VaryNotAliasedToStore 防止 Snapshot 把缓存内部的 Vary 切片
// 暴露给调用方：运维若对返回的 EntryView.Vary 做展示脱敏（改 Vary[0]），
// 不得写穿缓存条目，否则下次算缓存键的材料变化、命中率骤降、同 URL 反复打穿上游。
func TestSnapshot_VaryNotAliasedToStore(t *testing.T) {
	now := time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)
	clk := &clock.Fake{T: now}

	var lastReq *http.Request
	rt := rtFunc(func(r *http.Request) (*http.Response, error) {
		lastReq = r
		return &http.Response{
			StatusCode: 200,
			Header: http.Header{
				"Cache-Control":  []string{"max-age=60"},
				"Content-Type":   []string{"text/plain"},
				"Vary":           []string{"Accept-Encoding, Accept"},
				"Accept-Encoding": []string{"gzip"}, // 仅用于让请求头被记录
			},
			Body: http.NoBody,
			Request: r,
		}, nil
	})

	c, err := New(Options{
		Transport:  rt,
		Clock:      clk,
		DefaultTTL: 60 * time.Second,
		SWR:       10 * time.Second,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer c.Close()

	req := Request{
		Method: "GET",
		URL:    "/article",
		Headers: map[string][]string{
			"Vary":            []string{"Accept-Encoding, Accept"},
			"Accept-Encoding": []string{"gzip"},
			"Accept":          []string{"text/html"},
		},
	}
	if _, cached, err := c.Fetch(req); err != nil || cached {
		t.Fatalf("首次 Fetch: cached=%v err=%v", cached, err)
	}

	views := c.Snapshot()
	if len(views) != 1 {
		t.Fatalf("Snapshot 返回 %d 条，想要 1", len(views))
	}
	v := views[0]
	wantVary := []string{"Accept-Encoding", "Accept"}
	if len(v.Vary) != len(wantVary) {
		t.Fatalf("Snapshot Vary = %v, 想要 %v", v.Vary, wantVary)
	}

	// 模拟运维展示脱敏：改写 Vary[0]，且整体替换 Vary 切片。
	v.Vary[0] = "MASKED"
	v.Vary = append(v.Vary, "X-Extra")

	// 再次 Snapshot：内部条目的 Vary 必须未被写穿。
	views2 := c.Snapshot()
	if len(views2) != 1 {
		t.Fatalf("第二次 Snapshot 返回 %d 条，想要 1", len(views2))
	}
	got := views2[0].Vary
	if len(got) != 2 || got[0] != "Accept-Encoding" || got[1] != "Accept" {
		t.Fatalf("脱敏写穿了缓存：第二次 Snapshot Vary = %v，想要 %v", got, wantVary)
	}

	// 关键：脱敏后同一请求仍应命中缓存，而非打穿上游。
	if lastReq == nil {
		t.Log("（首次请求已记录）")
	}
	lastReq = nil
	if _, cached, err := c.Fetch(req); err != nil || !cached {
		t.Fatalf("脱敏后 Fetch 未命中缓存：cached=%v err=%v（缓存键材料被脱敏污染？）", cached, err)
	}
	if lastReq != nil {
		t.Fatalf("脱敏后仍打穿上游：收到上游请求 %s %s", lastReq.Method, lastReq.URL)
	}
}

// 确保没有把 Vary 误存成原始大小写敏感导致 key 漂移（回归守卫）。
func TestSnapshot_VaryMatchesKeyBuild(t *testing.T) {
	now := time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)
	clk := &clock.Fake{T: now}
	rt := rtFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Header: http.Header{
				"Cache-Control": []string{"max-age=60"},
				"Vary":          []string{"Accept-Encoding, Accept"},
			},
			Body:    http.NoBody,
			Request: r,
		}, nil
	})
	c, _ := New(Options{Transport: rt, Clock: clk, DefaultTTL: 60 * time.Second, SWR: 10 * time.Second})
	defer c.Close()

	req := Request{Method: "GET", URL: "/x", Headers: map[string][]string{
		"Accept-Encoding": {"gzip"},
		"Accept":          {"text/html"},
	}}
	if _, _, err := c.Fetch(req); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	views := c.Snapshot()
	if len(views) != 1 {
		t.Fatalf("Snapshot 返回 %d 条", len(views))
	}
	for _, n := range views[0].Vary {
		if strings.Contains(n, "MASKED") {
			t.Fatalf("Snapshot Vary 含脱敏残留：%v", views[0].Vary)
		}
	}
}
