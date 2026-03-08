package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		log.Info("hello request received",
			slog.String("method", r.Method),
			slog.String("remote_addr", r.RemoteAddr),
		)
		fmt.Fprintf(w, "Hello World @ %s\n", time.Now().Format(time.RFC3339))
	})

	srv := &http.Server{Addr: ":8080", Handler: mux}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	// Start HTTP server
	g.Go(func() error {
		log.Info("testserver starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	// Graceful shutdown when context is cancelled
	g.Go(func() error {
		<-ctx.Done()
		log.Info("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil {
		log.Error("server error", slog.Any("error", err))
		os.Exit(1)
	}
}
