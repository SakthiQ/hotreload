package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sakthi-narayan/hotreload/internal/builder"
	"github.com/sakthi-narayan/hotreload/internal/runner"
	"github.com/sakthi-narayan/hotreload/internal/watcher"
)

type App struct {
	logger *slog.Logger
	Config
	builder *builder.Builder
	runner  *runner.Runner
	watcher *watcher.Watcher
}

type Config struct {
	Root     string
	BuildCmd string
	ExecCmd  string
	Excludes []string
}

func New(logger *slog.Logger, cfg Config) (*App, error) {
	b := builder.New(logger)
	r := runner.New(logger)
	w, err := watcher.New(logger, cfg.Root, 200*time.Millisecond, cfg.Excludes)
	if err != nil {
		return nil, err
	}

	return &App{
		logger:  logger,
		Config:  cfg,
		builder: b,
		runner:  r,
		watcher: w,
	}, nil
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.triggerRebuild()
	go a.handleSignals(cancel)

	if err := a.watcher.Start(); err != nil {
		a.logger.Error("failed to start watcher", slog.Any("error", err))
		return err
	}
	defer a.watcher.Stop()

	var buildCtx context.Context
	var cancelBuild context.CancelFunc

	for {
		select {
		case <-ctx.Done():
			if cancelBuild != nil {
				cancelBuild()
			}
			if err := a.runner.Stop(); err != nil {
				a.logger.Warn("error stopping process during shutdown", slog.Any("error", err))
			}
			return nil

		case <-a.watcher.Trigger:
			a.logger.Info("file change detected, restarting...")
			if cancelBuild != nil {
				cancelBuild()
			}
			if err := a.runner.Stop(); err != nil {
				a.logger.Warn("error stopping process before restart", slog.Any("error", err))
			}
			buildCtx, cancelBuild = context.WithCancel(ctx)
			go a.runBuild(buildCtx)
		}
	}
}

// handleSignals blocks until SIGINT or SIGTERM is received, then cancels the root context.
func (a *App) handleSignals(cancel context.CancelFunc) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	a.logger.Info("shutting down...")
	cancel()
}

// runBuild executes the build command and, on success, starts the child process.
func (a *App) runBuild(ctx context.Context) {
	if err := a.builder.Build(ctx, a.Root, a.BuildCmd); err != nil {
		if ctx.Err() == nil {
			a.logger.Error("build failed", slog.Any("error", err))
		}
		return
	}
	if ctx.Err() != nil {
		return
	}
	if err := a.runner.Start(a.Root, a.ExecCmd); err != nil {
		a.logger.Error("failed to start process", slog.Any("error", err))
	}
}

func (a *App) triggerRebuild() {
	select {
	case a.watcher.Trigger <- struct{}{}:
	default:
	}
}
