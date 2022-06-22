package api

import (
	"context"
	"sync"
)

type api struct {
	ctx context.Context
	key string
	mux sync.Mutex
}

func New(ctx context.Context, key string) *api {
	return &api{
		ctx: ctx,
		key: key,
	}
}
