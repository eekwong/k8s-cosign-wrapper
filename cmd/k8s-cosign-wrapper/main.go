package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/eekwong/k8s-cosign-wrapper/cmd/k8s-cosign-wrapper/api"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httplog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "k8s-cosign-wrapper",
		Version: "v0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "key",
				Usage:   "path to the public key file, KMS URI or Kubernetes Secret",
				EnvVars: []string{"KEY"},
			},
			&cli.BoolFlag{
				Name:    "k8s-keychain",
				Usage:   "whether to use the kubernetes keychain instead of the default keychain",
				EnvVars: []string{"K8S_KEYCHAIN"},
			},
		},
		Action: func(c *cli.Context) error {
			key := strings.TrimSpace(c.String("key"))
			if key == "" {
				return errors.New("key must be present")
			}

			k8sKeychain := c.Bool("k8s-keychain")

			ctx, cancel := context.WithCancel(context.Background())

			server := &http.Server{
				Addr:    ":8080",
				Handler: setupChiRouter(ctx, key, k8sKeychain),
			}

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGILL, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV)

			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := server.ListenAndServe(); err != nil {
					if err != http.ErrServerClosed {
						log.Error().Err(err).Msg("error in http.Server.ListenAndServe")
					}
				}
			}()

			<-sigs
			cancel()

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				log.Error().Err(err).Msg("error in shutting down the HTTP server")
			}

			wg.Wait()

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("error in app.Run()")
	}
}

func setupChiRouter(ctx context.Context, key string, k8sKeychain bool) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Timeout(60 * time.Second))

	logger := httplog.NewLogger("k8s-cosign-wrapper", httplog.Options{
		JSON:    true,
		Concise: false,
	})
	r.Use(httplog.RequestLogger(logger))

	api.SetupRoutes(ctx, r, key, k8sKeychain)

	return r
}
