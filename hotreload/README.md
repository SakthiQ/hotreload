# HotReload CLI

A simple file‑watching hot reload tool written in Go.  It watches a
project directory, rebuilds when source files change, and restarts the
server process automatically.  The implementation is intentionally small
and self‑contained; the only third‑party dependency is `fsnotify` for
filesystem events and `log/slog` from the standard library is used for
logging.

This repository contains two main components:

* `hotreload/` – the CLI tool source code (`main.go` plus `internal/` packages).
* `testserver/` – a minimal HTTP server used for demonstration and testing.

There is also a `Makefile` at the repo root which can build and run a
demo of the tool.

## Features

- Watch an entire directory tree (recursively) for changes.
- Ignore common noisy directories (`.git`, `node_modules`, `vendor`,
	build artifacts, hidden dirs) plus user‑specified excludes.
- Debounce rapid file events (200 ms) to avoid unnecessary rebuilds.
- Run a build command on change and stream stdout/stderr in real time.
- Start/stop the server process cleanly; kill stubborn processes if
	necessary.
- Avoid restart loops when the server crashes on startup.
- Basic CLI interface with flags: `--root`, `--build`, `--exec`,
	`--exclude`.
- Works on Windows and Unix; uses `cmd` on Windows and `sh` on Unix.

## Quickstart

1. **Build the demo utilities:**

	 ```powershell
	 cd hotreload
	 go build -o ../hotreload.exe .    # produces hotreload.exe at project root
	 cd ..
	 cd testserver && go mod tidy && go build -o server.exe .
	 cd ..
	 ```

2. **Run the hotreload tool** against the sample server:

	 ```powershell
	 ./hotreload.exe --root ./testserver \
			 --build "go build -o server.exe ./testserver" \
			 --exec "./server.exe"
	 ```

	 On first run the tool will immediately trigger a build and start the
	 server.  Edit `testserver/main.go` and save; the watcher will detect
	 the change, rebuild the server binary, and restart it automatically.

3. **Cleanup:**

	 ```sh
	 make clean
	 ```

	 (or manually remove `hotreload.exe`, `testserver/server.exe`, etc.)

	### Platform‑specific build/run

	The commands in the quickstart section work on both Unix and Windows, but
	here is the full sequence spelled out for each shell.

	#### Unix (macOS/Linux)

	```sh
	# compile the CLI and sample server
	cd hotreload && go build -o ../hotreload && cd ..
	cd testserver && go mod tidy && go build -o server && cd ..

	# run hotreload
	./hotreload --root ./testserver \
				--build "go build -o server ./testserver" \
				--exec "./server"

	# press Ctrl+C to stop; use `make clean` to remove binaries
	```

	#### Windows (PowerShell)

	```powershell
	# compile the CLI and sample server
	cd hotreload
	go build -o ../hotreload.exe
	cd ..
	cd testserver
	go mod tidy
	go build -o server.exe
	cd ..

	# run hotreload
	.\hotreload.exe --root ".\testserver" `
		--build "go build -o server.exe .\testserver" `
		--exec ".\server.exe"

	# stop with Ctrl+C; remove binaries via `make clean` or manually
	```


## CLI Flags

```text
--root <path>      project root directory to watch (required)
--build <command>  build command to run on changes (required)
--exec <command>   command used to launch the built binary (required)
--exclude <dirs>   comma-separated list of relative paths to ignore
``` 

## Makefile

The provided `Makefile` automates the demo workflow on platforms with `make`:

```makefile
demo:
		cd testserver && go mod tidy && cd ..
		cd hotreload && go build -o ../hotreload.exe . && cd ..
		./hotreload.exe --root ./testserver \
				--build "go build -o server.exe ./testserver" \
				--exec "./server.exe"

testserver:
		cd testserver && go mod tidy && go build -o server .

clean:
		rm -f server testserver/server hotreload.exe
```

## Implementation Notes

- `internal/watcher` wraps `fsnotify` and handles recursion, debouncing,
	and ignore/exclude logic.
- `internal/builder` executes build commands with context cancellation.
- `internal/runner` starts/stops child processes and handles both graceful
	and forceful termination.
- `internal/app` ties everything together with a main event loop.

This project was built from scratch and does **not** use existing
hot‑reload frameworks such as `air`, `realize`, or `reflex`.