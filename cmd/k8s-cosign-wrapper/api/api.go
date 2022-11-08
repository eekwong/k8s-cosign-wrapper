package api

import (
	"context"
)

type api struct {
	ctx         context.Context
	key         string
	k8sKeychain bool
}

func New(ctx context.Context, key string, k8sKeychain bool) *api {
	return &api{
		ctx:         ctx,
		key:         key,
		k8sKeychain: k8sKeychain,
	}
}
