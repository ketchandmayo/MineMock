# MineMock

MineMock is a minimal TCP mock server for Minecraft (Java Edition), written in Go.

It can:
- respond to **status ping** (server list preview in Minecraft);
- accept **login** and close the connection with a configurable error message;
- simulate a delay before returning an error (useful for launcher/bot/monitoring tests).

## Requirements

- Go 1.22+ (recommended)
- A TCP port (default: `25565`)

## Quick Start

```bash
go run .
```

By default, the server listens on `127.0.0.1:25565`.

Logs are written to:
- stdout;
- `server.log` in the project root.

## Run with Environment Variables

### Linux/macOS (bash/zsh)

```bash
IP=0.0.0.0 \
PORT=25565 \
MOTD='§aMineMock\\n§eTest server' \
VERSION_NAME='1.20.1' \
MAX_PLAYERS=100 \
ONLINE_PLAYERS=7 \
ERROR='§cServer is temporarily unavailable' \
ERROR_DELAY_SECONDS=2 \
FORCE_CONNECTION_LOST_TITLE=true \
go run .
```

### Windows (PowerShell)

```powershell
$env:IP = "0.0.0.0"
$env:PORT = "25565"
$env:MOTD = "§aMineMock\\n§eTest server"
$env:VERSION_NAME = "1.20.1"
$env:MAX_PLAYERS = "100"
$env:ONLINE_PLAYERS = "7"
$env:ERROR = "§cServer is temporarily unavailable"
$env:ERROR_DELAY_SECONDS = "2"
$env:FORCE_CONNECTION_LOST_TITLE = "true"
go run .
```

## Build Binary

```bash
go build -o minemock .
./minemock
```

## Configuration

All settings are configured via environment variables:

| Variable | Default | Description |
|---|---:|---|
| `IP` | `127.0.0.1` | Bind IP address |
| `PORT` | `25565` | Server TCP port |
| `ERROR` | `§r§7MineMock§r\\n§cServer is temporarily unavailable. Try again later.` | Disconnect message used during login |
| `ERROR_DELAY_SECONDS` | `0` | Delay before sending error (seconds) |
| `FORCE_CONNECTION_LOST_TITLE` | `false` | `false`: disconnect directly in login; `true`: login success -> disconnect in play (shows "Connection Lost") |
| `MOTD` | `§c§oMine§4§oMock§r\\n§6Minecraft mock server on golang§r | §eWelcome☺` | MOTD in server status response |
| `VERSION_NAME` | `1.20.1` | Displayed Minecraft version |
| `PROTOCOL` | derived from `VERSION_NAME` | Protocol number used in status ping |
| `MAX_PLAYERS` | `20` | `players.max` in status response |
| `ONLINE_PLAYERS` | `7` | `players.online` in status response |

### `PROTOCOL` Note

If `PROTOCOL` is not set, it is auto-selected from `VERSION_NAME`.
If `VERSION_NAME` is unknown, fallback `763` is used.

## How to Verify

1. Start the server: `go run .`
2. Open Minecraft Java Edition.
3. Add server `127.0.0.1:25565`.
4. Your `MOTD` should appear in the server list.
5. On login attempt, you should receive the message from `ERROR`.

## Tests

```bash
go test ./...
```

## Project Structure

- `main.go` — entry point, logger setup, env config loading, server startup;
- `internal/config` — loading and parsing env-based configuration;
- `internal/server` — TCP server and handshake/status/login handling;
- `internal/protocol` — Minecraft packet encoding/decoding.
