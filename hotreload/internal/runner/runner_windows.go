//go:build windows

package runner

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func createCommand(dir, execCmd string) *exec.Cmd {
	// Parse the string into fields to bypass cmd.exe
	// This ensures cmd.Wait() waits for the ACTUAL application to close, not just the wrapper shell!
	parts := strings.Fields(execCmd)
	if len(parts) == 0 {
		return nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func killProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	// Kill the entire process tree holding the port forcefully immediately
	// Graceful shutdown often fails or holds the port on Windows spawned children
	killCmd := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	return killCmd.Run()
}

func forceKillProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	// Fallback to Go's internal kill if taskkill failed in killProcess.
	// This directly invokes the Windows TerminateProcess API.
	return cmd.Process.Kill()
}
