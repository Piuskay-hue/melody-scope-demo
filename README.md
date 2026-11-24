# melody-scope-demo

A small, complete example on [Melody v3](https://github.com/precision-soft/melody) (v3.14.6):
**one scoped service, resolved from an HTTP route and from a CLI command**, with its own
instance for every request and every command run, and closed when that unit of work
ends, including when the handler or the command fails.

Melody is the only direct dependency. There is no database, message broker or
configuration layer.

## What it shows

- **One instance per unit of work.** Each HTTP request and each CLI run gets its own `WorkStamp`.
- **The same instance within one unit.** Two resolutions inside one request (or one run) return the same pointer.
- **Invisible to the root container.** The application's root container doesn't hold the scoped service.
- **Closed even on failure.** A request or command that fails after resolving the service still closes it, and a close failure is reported, not swallowed.

## Files

| File | What it holds |
|---|---|
| `main.go` | Boots the application, registers both routes and both commands. |
| `services.go` | The singleton `Ledger`, the scoped `WorkStamp`, and both registrations side by side. |
| `http_handler.go` | `GET /scope-report`: resolves the scoped service twice in the request's scope. |
| `cli_command.go` | `scope:report`: the same two resolutions from the console. |
| `error_entry_points.go` | `GET /scope-fail` and `scope:fail`: fail on purpose, after resolving. |
| `error_path.go` | `demo.fragile`, a scoped service whose `Close` refuses. |
| `go.mod`, `go.sum` | One direct dependency: `github.com/precision-soft/melody/v3 v3.14.6`. |
| `.env` | The one key the HTTP process needs to boot (`MELODY_ENV=dev`). |
| `.github/workflows/demo.yml` | Builds the demo and runs every scenario below on Linux and Windows. |

## Requirements

- Go 1.25 or newer
- Linux, macOS or Windows

## Run it

```sh
git clone https://github.com/Piuskay-hue/melody-scope-demo.git
cd melody-scope-demo
go build -o melody-scope-demo .
```

On Windows, build with `go build -o melody-scope-demo.exe .` and run `.\melody-scope-demo.exe`.

**HTTP.** With no arguments the binary starts the HTTP server on `:8080`:

```sh
./melody-scope-demo
```

In another terminal, call the route twice and compare the identifiers:

```sh
curl http://localhost:8080/scope-report
curl http://localhost:8080/scope-report
curl -i http://localhost:8080/scope-fail     # answers 500; the server still prints [closed] for its stamp
```

**CLI.** Any command name runs that command in the console:

```sh
./melody-scope-demo scope:report
./melody-scope-demo scope:fail
./melody-scope-demo scope:fail --broken-close
```

## Observed output

Copied from the workflow run on GitHub's `windows-latest` runner with Go 1.25.14
([run #1](https://github.com/Piuskay-hue/melody-scope-demo/actions/runs/45415191764)).
The `ubuntu-latest` job in the same run gave the same results. The number inside each
identifier (`stamp-<pid>-<n>`) is the process id, so it differs from run to run.

Melody also prints its own startup log lines and a `[started]` / `[finished]` banner around
each command. They are left out below so the demo's own lines are easy to read.

**Two requests to `/scope-report`, in the same server process:**

```json
{"entryPoint":"http","route":"/scope-report","firstResolution":"stamp-7436-1","secondResolution":"stamp-7436-1","sameInstanceWithinOneUnitOfWork":true,"rootContainerSeesScopedService":false,"createdSoFar":["stamp-7436-1"],"closedSoFar":[]}
{"entryPoint":"http","route":"/scope-report","firstResolution":"stamp-7436-2","secondResolution":"stamp-7436-2","sameInstanceWithinOneUnitOfWork":true,"rootContainerSeesScopedService":false,"createdSoFar":["stamp-7436-1","stamp-7436-2"],"closedSoFar":["stamp-7436-1"]}
```

The second request got a new instance, and the first request's instance is already in
`closedSoFar`.

**Then `GET /scope-fail`** answered `500`, and the server printed one `[closed]` line per request,
the failed one included:

```text
[closed] stamp-7436-1
[closed] stamp-7436-2
[closed] stamp-7436-3
```

**`scope:report`:**

```text
entry point:              cli
first resolution:         stamp-6536-1
second resolution:        stamp-6536-1
same instance in one run: true
root container sees it:   false
created so far:           [stamp-6536-1]
closed so far:            []
[closed] stamp-6536-1
```

The `[closed]` line is printed after the command body has returned.

**`scope:fail --broken-close`** (exit code 1):

```text
resolved: stamp-6936-1
resolved: a service that refuses to close
[error] cli command failed with shutdown errors
[closed] stamp-6936-1
```

The command's error and the close failure are reported together, and the stamp is still
closed. The `[error]` line goes to standard error, so its position relative to the other
lines can vary between terminals.

## What this demo does not claim

- **One CLI run is one scope.** A command that processes a thousand rows holds a single
  scope for the whole run. A long-running command that needs fresh scoped services per
  unit of work creates a child scope per unit, as Melody's cron runner does. This demo
  doesn't show that pattern.
- **The identifiers are a teaching device**, not a correlation-id design.
- **This is not an application template.** It is the smallest program that makes the
  scoped lifetime observable.
