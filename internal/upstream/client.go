package upstream

import (
	"context"
	"io"
	"net/http"
	"strings"

	"example.com/httpcachex/internal/entry"
)

type Req struct {
	Method  string
	URL     string
	Headers map[string][]string
}

type Result struct {
	Status  int
	Headers map[string][]string
	Body    []byte
}

func Do(ctx context.Context, rt http.RoundTripper, req Req) (Result, error) {
	if rt == nil {
		return Result{}, errNil
	}
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, nil)
	if err != nil {
		return Result{}, err
	}
	for k, vs := range req.Headers {
		for _, v := range vs {
			httpReq.Header.Add(k, v)
		}
	}
	resp, err := rt.RoundTrip(httpReq)
	if err != nil {
		return Result{}, err
	}
	// 无论后续读取/拷贝成功与否，都必须关闭 Body，否则底层 TCP 连接无法
	// 归还连接池，压测下句柄/FD 只增不减，最终连接耗尽、新请求全部卡死。
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, err
	}
	hdrs := map[string][]string{}
	for k, vs := range resp.Header {
		hdrs[k] = append([]string(nil), vs...)
	}
	return Result{
		Status:  resp.StatusCode,
		Headers: hdrs,
		Body:    entry.CloneBytes(body),
	}, nil
}

type nilErr string

func (e nilErr) Error() string { return string(e) }

const errNil = nilErr("upstream: nil transport")

func IsHTML(ct string) bool {
	return strings.Contains(strings.ToLower(ct), "text/html")
}
