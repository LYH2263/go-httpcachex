# go-httpcachex

HTTP 反向缓存代理库 + `cached` 管理服务：Cache-Control/Vary 键、fresh/stale、stale-while-revalidate、条件请求。

## Build / Test

```bash
go build ./...
go test ./... -count=1
```

## Daemon

```bash
go run ./cmd/cached -addr :8106 -web web
```
