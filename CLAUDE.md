# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test

```sh
go build -o webhooker ./cmd/webhooker
nix build .#webhooker
nix flake check
```

## Architecture

Channel-based actor model:

```
Receiver ──chan──→ Router ──chan──→ File (persistent)
                      │
                      └──chan──→ IPC Client (temporary)
```

### Actors

| Actor | Package | Description |
|-------|---------|-------------|
| Receiver | `internal/receiver` | HTTP handler, body parsing, escaping |
| Router | `internal/router` | Route matching, dynamic register/unregister |
| File | `internal/file` | Persistent log writer with rotation |
| IPC Server | `internal/ipc` | Unix socket, temporary route management |

### IPC Protocol

JSON over Unix socket:

```
client → {"type":"register"}
server → {"type":"registered","path":"/tmp-abc123","url":"https://example.com/tmp-abc123"}
server → {"type":"webhook","data":"..."}
```

### Environment Variables

All prefixed with `WEBHOOKER_`:
- `DATA_DIR` — base directory (default: `~/.local/state/webhooker` or `/var/lib/webhooker` for root)
- `SOCKET` — Unix socket path (default: `/run/webhooker/webhooker.sock`)
- `LOG_DIR` — log directory (default: `$DATA_DIR`)
- `PORT` — HTTP port (default: `8080`)
- `DOMAIN` — public base URL for webhook endpoints (e.g., `https://example.com`)
- `ROUTES` — persistent routes (format: `secret:name,secret:name`)

### Security

- Unregistered routes silently ignored (no enumeration hints)
- Temporary routes not written to persistent log
