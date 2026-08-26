# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A caching reverse proxy for Windows debug symbols (PDB files). It fronts `msdl.microsoft.com` (or another upstream symbol server), serving files from a local disk cache and lazily fetching+persisting misses. Deployed as a public symbol mirror node (e.g. `msdl.szdyg.cn`), so the on-disk cache doubles as the served symbol store.

Go 1.22 + gin. ~90 lines of application code total across three packages.

## Commands

```bash
go build -o server .          # build (CI uses: CGO_ENABLED=0 go build -ldflags "-s -w" -o server)
go run .                      # run locally; listens on 0.0.0.0:9000, caches to /pdb
go vet ./...
gofmt -l .
```

There are no tests in this repo. Manual verification is a request against a known symbol:

```bash
curl -v http://localhost:9000/download/symbols/wrpcrt4.pdb/0DBDD41E0805EAAB4F3FE2365B9EC7A91/wrpcrt4.pdb
```

Docker: `docker compose up -d` (mounts `./pdb` → `/pdb`, publishes 9000). The image builds `FROM scratch`, so the binary must stay CGO-free and any new runtime file dependency has to be copied explicitly in the Dockerfile's second stage.

## Architecture

Request flow — one route, one handler, no middleware beyond gin's default logger/recovery:

- `pdb_proxy.go` — registers the symbol route for GET and HEAD on an explicit `http.Server`. Note the **duplicated `:pdbname` param**: this mirrors the symsrv URL layout (`<name>/<hash>/<name>`) and gin permits it; `c.Param("pdbname")` returns the first binding, so the third URL segment is effectively ignored. Preserve this shape — symbol clients (WinDbg, symsrv.dll) construct exactly this path.
- `pdb/pdb_query.go` — validates the params, then cache lookup → hit serves, miss goes through `singleflight` to `DownLoadFile`. Returns **404 with an empty body** on any failure. The empty body is deliberate (commit "包体不为空，symsrv还是会乱下"): symsrv treats a non-empty error body as a valid payload and caches garbage.
- `pdb/pdb_save.go` — `DownLoadFile`: writes to a same-directory temp file, verifies `Content-Length`, then `os.Rename` into place.
- `conf/pdb_conf.go` — all config from env vars in `init()`, no config file: `PDB_DIR` (default `/pdb`), `PDB_SERVER` (default `https://msdl.microsoft.com/download/symbols`), `SERVER_PORT` (default `0.0.0.0:9000`).

## Things to know before changing behavior

These are load-bearing; the obvious "simplifications" reintroduce bugs that were measured and fixed:

- **Never write directly to the final cache path.** The temp-file + `os.Rename` dance in `DownLoadFile` is what guarantees a cached file is either absent or complete. Writing straight to the destination let concurrent requests truncate each other (three concurrent requests for one 4 MB symbol returned 4 MB / 256 KB / 128 KB, all with HTTP 200), and left crash-truncated files that `os.Stat` then served forever as valid.
- **`os.Stat` alone is not a cache hit.** Use `isRegularFile`; a bare existence check happily serves directories.
- **Don't set `WriteTimeout` on the `http.Server`.** Symbol files reach hundreds of MB and a write deadline truncates legitimate downloads. `ReadHeaderTimeout` + `IdleTimeout` cover slowloris.
- `tempFileTTL` (1h) must stay greater than `downloadTimeout` (30m), or `cleanStaleTemps` can delete an in-flight download.
- Cache directories need `0755`, not `0644` — without the execute bit a non-root process cannot create files inside them.
- Param validation in `pdb_query.go` is the real traversal defense. It currently overlaps with `http.ServeFile`'s internal `containsDotDot` check, so removing it looks harmless in testing — it isn't, if `c.File` is ever replaced with a manual `os.Open`.

Deliberately **not** implemented, both because they need a policy decision: case normalization of `pdbname`/`pdbhash` (would invalidate every existing on-disk cache entry and trigger a full re-download on live nodes), and negative caching / cache eviction (needs TTL and size limits). The cache currently only grows.

## Release / deployment

**docker compose is the only supported deployment path.** The systemd `install.sh` / `uninstall.sh` scripts and the entire `.github/` workflow directory were removed deliberately — don't reintroduce CI or a second install mechanism without being asked.

There is no CI in this repo. `docker-compose.yml` pulls `szdyg/pdb_proxy` from Docker Hub, which is published outside this repository; building and pushing that image is a manual step. To build locally instead: `docker build -t szdyg/pdb_proxy .`

The Go toolchain version is pinned in two places that must move together: `go.mod` (`go 1.22`) and `Dockerfile` (`golang:1.22-alpine`). `go get` will silently bump the `go.mod` directive — check it afterwards.

## Conventions

Comments and user-facing docs (`ReadME.md`) are in Chinese; keep new ones consistent with the surrounding file.
