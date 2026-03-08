package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sakthi-narayan/hotreload/internal/app"
	"github.com/sakthi-narayan/hotreload/internal/logger"
)

func main() {
	root := flag.String("root", "", "project root directory to watch")
	buildCmd := flag.String("build", "", "build command (sh -c)")
	execCmd := flag.String("exec", "", "exec command (sh -c)")
	excludeStr := flag.String("exclude", "", "comma-separated list of relative directories to exclude from watching")

	flag.Parse()

	if *root == "" {
		fmt.Fprintln(os.Stderr, "missing required flag: --root")
		os.Exit(1)
	}
	if *buildCmd == "" {
		fmt.Fprintln(os.Stderr, "missing required flag: --build")
		os.Exit(1)
	}
	if *execCmd == "" {
		fmt.Fprintln(os.Stderr, "missing required flag: --exec")
		os.Exit(1)
	}

	var excludes []string
	if *excludeStr != "" {
		for _, e := range strings.Split(*excludeStr, ",") {
			trimmed := strings.TrimSpace(e)
			if trimmed != "" {
				excludes = append(excludes, trimmed)
			}
		}
	}

	log := logger.NewLogger()
	log.Info("hotreload starting")

	cfg := app.Config{
		Root:     *root,
		BuildCmd: *buildCmd,
		ExecCmd:  *execCmd,
		Excludes: excludes,
	}

	a, err := app.New(log, cfg)
	if err != nil {
		log.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}

	if err := a.Run(); err != nil {
		log.Error("app exited with error", "error", err)
		os.Exit(1)
	}

	log.Info("hotreload exit")
}
