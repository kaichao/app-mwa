# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test

```bash
# Build a specific router
cd app/router && go build -o ../../main-router .
cd app-base/router && go build -o ../../main-router-base .

# Run tests for a specific package
go test ./internal/vpath/... -v
go test ./internal/vpath0/... -v
go test ./internal/datacube/... -v
go test ./internal/queue/... -v

# Run all tests
go test ./... -v
```

Each module under `modules/` (beam-make, down-sample, fits-merge, fits-redist, pull-unpack, wait-queue) has its own Makefile/Dockerfile for building container images. The `Makefile` at repo root syncs to remote hosts (`dcu`, `p419`, `p419-n0`).

## Architecture

This is a **beam-forming pipeline** for MWA radio astronomy data, built on the [Scalebox](https://github.com/kaichao/scalebox) framework. It processes radio telescope data through a directed acyclic graph of stages.

### Pipeline Flow

```
tar-load → wait-queue → vtask-head → pull-unpack → beam-make → down-sample → fits-redist → fits-merge → fits24ch-move → vtask-tail
```

Each stage runs as a Scalebox module (container). A **main-router** (`app/router/main.go`) dispatches messages between stages.

### Two Router Variants

- **`app/router/`** — Full pipeline router with all stages
- **`app-base/router/`** — Simplified variant with only `default`, `down-sample`, `fits-merge` stages

### Router Pattern

The router receives messages from Scalebox with args: `<task-body> <headers-json>`. The `from_module` header dispatches to the corresponding handler function:

```go
fromFuncs = map[string]func(string, map[string]string) error{
    "":            fromNull,        // entry point
    "wait-queue":  fromWaitQueue,
    "beam-make":   fromBeamMake,
    "down-sample": fromDownSample,
    // ...
}
```

Each `fromXxx()` handler processes the incoming message, optionally operates on semaphores/variables, then calls `toYyy()` to forward to the next stage. `toYyy()` functions create new tasks via `task.AddWithMapHeaders()` or `task.AddTasks()` from the `scalebox` library.

### Message Body Format (strparse)

Messages use a hierarchical path format parsed by `internal/strparse`:
```
<obsID>                                          # full dataset
<obsID>/p<pointingBegin>_<pointingEnd>           # pointing range
<obsID>/p<p0>_<p1>/t<timeBegin>_<timeEnd>        # time cube
<obsID>/p<p0>_<p1>/t<t0>_<t1>/ch<channel>        # single channel
```

### Key Internal Packages

| Package | Purpose |
|---------|---------|
| `internal/vpath/` | Path mapping engine (current). Two-layer config `{classes, pools}`: a **class** maps data to weighted targets (a path or a pool), a **pool** aggregates members by max-free. The virtualization type is not declared — the same targets act as type a (aggregate), b1 (replicate) or b2 (merge) depending on which API the caller invokes (`Allocate` / `AllocateAll`). Copy selection uses stateless rendezvous hashing, so all processes agree without shared state. Pool membership and capacity are read from the `vpath:free-gb:<pool>` semaphore group. |
| `internal/vpath0/` | Legacy vpath implementation, **retained as an archive** — no production callers. Flat `map[category]` config with `static` / `aggregated` path types and a Scalebox-backed aggregator. Kept for reference together with its config `app/vpath0.yaml` and sema definitions `app/{mwa,preload}-vpath0.sema`; none of these archived files are copied into the router image. Its integration tests require a live Scalebox gRPC server. |
| `internal/datacube/` | Data cube configuration. Loads dataset parameters (time range, pointing range, channel count, step sizes) from `dataset.yaml`. Overridable via env vars (`TIME_BEGIN`, `TIME_END`, `POINTING_BEGIN`, `TIME_STEP`). |
| `internal/strparse/` | Parses the hierarchical message body format into obsID, pointing range, time range, and channel components. |
| `internal/queue/` | Redis-backed priority queue (`ZADD`/`ZPOPMIN`) for node selection ordering. |
| `internal/node/` | Node allocation — maps channels and pointings to compute nodes based on group index. |
| `internal/picker/` | Weighted random picker — no current callers (both vpath engines implement their own selection). |
| `internal/cache/` | Database cache helpers (currently commented out, using pgx). |

### Scalebox Integration

The pipeline relies heavily on the `github.com/kaichao/scalebox` library for:
- **task** — `task.AddWithMapHeaders()`, `task.AddTasks()` to enqueue downstream work
- **semaphore/vtask** — Flow control via named semaphores (e.g., `dat-ready`, `dat-done`, `fits-done`, `cube-vtask-done`)
- **variable** — Key-value storage for pointing data root paths

### Configuration

- `dataset.yaml` — Per-observation dataset parameters (time/pointing/channel dimensions)
- `/vpath.yaml` — Path mapping configuration (loaded at startup by `internal/vpath` via `initApp` in `app/router/init.go`)
- `app/vpath0.yaml` — Archived config for `internal/vpath0`, kept in the repo for reference only; not copied into the image
- Environment variables control behavior: `LOG_LEVEL`, `TIME_STEP`, `POINTING_BEGIN`, `TIME_BEGIN`, `TIME_END`, `RUN_MODE`, `USE_GLOBAL_POINTING`, `ORIGIN_ROOT`, `PRESTO_APP_ID`, `SSH_USER`, `SSH_PORT`
