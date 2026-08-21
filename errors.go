package httpcachex

import "errors"

var (
	ErrClosed      = errors.New("httpcachex: closed")
	ErrInvalid     = errors.New("httpcachex: invalid argument")
	ErrNoTransport = errors.New("httpcachex: no transport")
	ErrUpstream    = errors.New("httpcachex: upstream failed")
	ErrPersist     = errors.New("httpcachex: persist failed")
	ErrCanceled    = errors.New("httpcachex: canceled")
	ErrNotFound    = errors.New("httpcachex: not found")
	ErrNoStore     = errors.New("httpcachex: no-store")
)
