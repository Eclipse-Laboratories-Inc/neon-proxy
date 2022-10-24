package proxy

import "context"

type Proxy struct {
	ctx context.Context
}

func NewProxy(
	ctx context.Context,
) *Proxy {
	return &Proxy{
		ctx: ctx,
	}
}
