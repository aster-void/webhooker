# webhooker

Simple webhook receiver with temporary route support.

## Installation

### Nix (推奨)

```sh
nix profile install github:aster-void/webhooker
```

### Go

```sh
go install github.com/aster-void/webhooker/cmd/webhooker@latest
```

### ソースからビルド

```sh
git clone https://github.com/aster-void/webhooker.git
cd webhooker
go build -o webhooker ./cmd/webhooker
```

## Usage

```sh
# Start daemon
webhooker daemon

# Get temporary webhook URL (streams to stdout)
webhooker
```

## Configuration

| Env | Default | Description |
|-----|---------|-------------|
| `WEBHOOKER_DATA_DIR` | `~/.local/state/webhooker` | Base directory for socket and logs |
| `WEBHOOKER_SOCKET` | `/run/webhooker/webhooker.sock` | Unix socket path |
| `WEBHOOKER_LOG_DIR` | `$DATA_DIR` | Log directory |
| `WEBHOOKER_PORT` | `8080` | HTTP listen port |
| `WEBHOOKER_DOMAIN` | (none) | Public base URL for copyable webhook URLs |
| `WEBHOOKER_ROUTES` | (none) | Persistent route mapping |

## Routes

### Persistent Routes

```sh
WEBHOOKER_ROUTES='secret123:github,abc789:stripe' webhooker daemon
```

- `POST /secret123` → logged to file as `github`
- `POST /abc789` → logged to file as `stripe`
- Unknown paths → silently ignored

### Temporary Routes

```sh
$ webhooker
listening on /tmp-a1b2c3d4
```

With `WEBHOOKER_DOMAIN` configured on the daemon:

```sh
$ webhooker
listening on https://hooks.example.com/tmp-a1b2c3d4
```

- Temporary URL assigned per client
- Webhooks stream to stdout as JSON
- Route deleted when client disconnects
- Not written to persistent log

## Log Format

```
<timestamp> POST <path> <escaped-body>
```

## Limits

- Body: 1MB max
- Header: 8KB max
- Read timeout: 10s
- Write timeout: 5s
- Log rotation: 50MB or 24h idle → truncate

## Build

```sh
go build -o webhooker ./cmd/webhooker
nix build .#webhooker
```

## Test

```sh
nix flake check
```

## NixOS Module

```nix
{
  inputs.webhooker.url = "github:aster-void/webhooker";

  outputs = { nixpkgs, webhooker, ... }: {
    nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        webhooker.nixosModules.webhooker
        {
          services.webhooker = {
            enable = true;
            routes.github = "secret123";
          };
        }
      ];
    };
  };
}
```
