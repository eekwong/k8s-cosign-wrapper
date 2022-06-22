package api

import (
	"context"

	"github.com/go-chi/chi"
)

func SetupRoutes(ctx context.Context, r *chi.Mux, key string, k8sKeychain bool) {
	api := New(ctx, key, k8sKeychain)
	r.Post("/verify", api.verify())
}
