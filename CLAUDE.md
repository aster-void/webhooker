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
Receiver ──chan──→ Router ──chan──→ IPC Client (ephemeral)
```

### Actors

| Actor | Package | Description |
|-------|---------|-------------|
| Receiver | `internal/receiver` | HTTP handler, body parsing, escaping |
| Router | `internal/router` | Route matching, dynamic register/unregister |
| IPC Server | `internal/ipc` | Unix socket, ephemeral route management |

### IPC Protocol

JSON over Unix socket:

```
client → {"type":"register"}
server → {"type":"registered","path":"/tmp-abc123","url":"https://example.com/tmp-abc123"}
server → {"type":"webhook","data":"..."}
```

### Environment Variables

All prefixed with `WEBHOOKER_`:
- `PORT` — HTTP port (default: `8080`)
- `DOMAIN` — public base URL for webhook endpoints (e.g., `https://example.com`)

Socket path:
- Server: `/run/webhooker/webhooker.sock` (root) or `$XDG_RUNTIME_DIR/webhooker/webhooker.sock` (user)
- Client: tries both paths

### Security

- Unregistered routes silently ignored (no enumeration hints)
