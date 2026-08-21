package entry

import (
	"reflect"
	"testing"
	"time"
)

// TestCloneEntry_VaryIsDeepCopy 防止 CloneEntry 共享 Vary 底层数组：
// 修改克隆体的 Vary 元素不得写穿源条目，否则 Snapshot 展示脱敏会污染缓存。
func TestCloneEntry_VaryIsDeepCopy(t *testing.T) {
	src := Entry{
		Key:    "GET|/x|Accept-Encoding=gzip",
		URL:    "/x",
		Vary:   []string{"Accept-Encoding", "Accept"},
		Status: 200,
		Headers: map[string][]string{
			"Cache-Control": {"max-age=30"},
		},
		Body:       []byte("hello"),
		Hits:       3,
		Stored:     time.Time{},
		FreshUntil: time.Time{},
		StaleUntil: time.Time{},
	}

	cloned := CloneEntry(src)

	// 克隆体应与源值相等（字段语义相等）。
	if !reflect.DeepEqual(cloned.Vary, src.Vary) {
		t.Fatalf("克隆体 Vary = %v, 想要 %v", cloned.Vary, src.Vary)
	}

	// 修改克隆体的 Vary 元素，源条目必须保持原样。
	if len(cloned.Vary) == 0 {
		t.Fatal("测试需要非空 Vary")
	}
	cloned.Vary[0] = "MASKED"
	if src.Vary[0] == "MASKED" {
		t.Fatalf("CloneEntry 泄漏内部切片：修改克隆体写穿了源 Vary，源 = %v", src.Vary)
	}

	// 追加到克隆体亦不得影响源（容量被复用时的边界）。
	cloned.Vary = append(cloned.Vary, "extra")
	if len(src.Vary) != 2 {
		t.Fatalf("克隆体 append 写穿源 Vary，源 len = %d", len(src.Vary))
	}
}
