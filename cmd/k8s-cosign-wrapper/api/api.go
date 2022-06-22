package api

import (
	"context"
	"sync"
)

type api struct {
	ctx         context.Context
	key         string
	k8sKeychain bool
	mux         sync.Mutex
}

func New(ctx context.Context, key string, k8sKeychain bool) *api {
	return &api{
		ctx:         ctx,
		key:         key,
		k8sKeychain: k8sKeychain,
	}
}
