//go:build !windows

package runner

import (
	"os"
	"os/exec"
	"syscall"
)

func createCommand(dir, execCmd string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", execCmd)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Assign child process to a new process group to allow killing the entire tree later
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

func killProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	// Ask nicely first (SIGTERM)
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
}

func forceKillProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	// Negative PID sends SIGKILL to the entire process group, ensuring tests/servers
	// spawned by the shell are forcefully killed alongside the parent shell.
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}