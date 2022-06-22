package api

import (
	"context"

	"github.com/go-chi/chi"
)

func SetupRoutes(ctx context.Context, r *chi.Mux, key string) {
	api := New(ctx, key)
	r.Post("/verify", api.verify())
}
