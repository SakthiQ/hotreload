package runner

import (
	"fmt"
	"log/slog"
	"os/exec"
	"sync"
	"time"
)

type Runner struct {
	logger *slog.Logger
	cmd    *exec.Cmd
	mu     sync.Mutex
	done   chan struct{}
}

func New(logger *slog.Logger) *Runner {
	return &Runner{
		logger: logger,
	}
}

func (r *Runner) Start(dir, execCmd string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cmd != nil {
		return fmt.Errorf("process already running")
	}

	r.logger.Info("starting process...")

	// Create OS-aware command (cmd.exe on Windows, sh with ProcessGroups on Unix)
	cmd := createCommand(dir, execCmd)
	if cmd == nil {
		return fmt.Errorf("exec command is empty or invalid")
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	r.cmd = cmd
	r.done = make(chan struct{})

	go func() {
		err := cmd.Wait()
		close(r.done)

		r.mu.Lock()
		if r.cmd == cmd {
			r.cmd = nil
		}
		r.mu.Unlock()

		if err != nil {
			r.logger.Warn("process exited", slog.Any("error", err))
		} else {
			r.logger.Info("process exited cleanly")
		}
	}()

	return nil
}

func (r *Runner) Stop() error {
	r.mu.Lock()
	cmd := r.cmd
	done := r.done
	r.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}

	r.logger.Info("stopping process...")

	// Use OS-specific kill logic (taskkill on Windows, syscall.Kill on Unix)
	killProcess(cmd)

	// Wait for process to exit gracefully
	if done != nil {
		select {
		case <-done:
			// Process exited gracefully
		case <-time.After(5 * time.Second):
			r.logger.Warn("timeout waiting for process to exit; forcefully killing stubborn processes...")
			forceKillProcess(cmd) // Fallback to forceful OS-specific kill

			// Block until the forceful kill is processed by the OS and the wait goroutine closes the channel
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				r.logger.Error("process refused to die even after forceful kill")
			}
		}
	}

	// Brief pause to allow OS to release bound sockets (TIME_WAIT state) before proceeding
	time.Sleep(500 * time.Millisecond)

	return nil
}
