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
	// 已取消则不发起请求：避免对慢上游做无用拉取、占用连接。
	if err := ctx.Err(); err != nil {
		return Result{}, err
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
		// 自定义 RoundTripper 可能不尊重 ctx，显式复核取消。
		if cerr := ctx.Err(); cerr != nil {
			return Result{}, cerr
		}
		return Result{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// 读体期间 ctx 取消是常见情形，按取消上报。
		if cerr := ctx.Err(); cerr != nil {
			return Result{}, cerr
		}
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
