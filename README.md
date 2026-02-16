# MineMock

**MineMock is a minimal TCP mock server for Minecraft (Java Edition), written in Go.**

It can:

- respond to **status ping** (server list preview in Minecraft);
- accept **login** and close the connection with a configurable error message;
- proxy whitelisted players to a real Minecraft server;
- simulate a delay before returning an error (useful for launcher/bot/monitoring tests).

## Quick Start

### Linux/macOS (bash/zsh)

```bash
./minemock_linux
```

### Windows (PowerShell)

Run the executable file `minemock_win.exe` or:

```powershell
.\minemock_win.exe
```

By default, the server listens on `127.0.0.1:25565`.

Logs are written to:

- stdout;
- `server.log` in the project root.

## Run with Environment Variables

### Linux/macOS (bash/zsh)

```bash
IP=127.0.0.1 \
PORT=25565 \
MOTD='§aMineMock\n§eTest server' \
VERSION_NAME='1.20.1' \
MAX_PLAYERS=100 \
ONLINE_PLAYERS=7 \
ERROR='§r§7MineMock§r\n§cServer is temporarily unavailable. Try again later.' \
ERROR_DELAY_SECONDS=2 \
FORCE_CONNECTION_LOST_TITLE=true \
REAL_SERVER_ADDR='127.0.0.1:25566' \
LOGIN_WHITELIST='Steve,Alex' \
./minemock_linux
```

### Windows (PowerShell)

```powershell
$env:IP = "127.0.0.1"
$env:PORT = "25565"
$env:MOTD = "§aMineMock\n§eTest server"
$env:VERSION_NAME = "1.20.1"
$env:MAX_PLAYERS = "100"
$env:ONLINE_PLAYERS = "7"
$env:ERROR = "§r§7MineMock§r\n§cServer is temporarily unavailable. Try again later."
$env:ERROR_DELAY_SECONDS = "2"
$env:FORCE_CONNECTION_LOST_TITLE = "true"
$env:REAL_SERVER_ADDR = "127.0.0.1:25566"
$env:LOGIN_WHITELIST = "Steve,Alex"
.\minemock_win.exe
```

## Build Binary

Requirements: Go 1.22+ (recommended)

```bash
go build -o minemock .
./minemock
```

## Configuration

All settings are configured via environment variables:

| Variable                      | Description                                                                                                    | Default                                                                   |
|-------------------------------|----------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------|
| `IP`                          | Bind IP address                                                                                                | `127.0.0.1`                                                               |
| `PORT`                        | Server TCP port                                                                                                | `25565`                                                                   |
| `ERROR`                       | Disconnect message used during login                                                                           | `\u00a7c\u00a7oMine\u00a74\u00a7oMock\u00a7r\n\u00a72Server is working` |
| `ERROR_DELAY_SECONDS`         | Delay before sending error (seconds)                                                                           | `0`                                                                       |
| `FORCE_CONNECTION_LOST_TITLE` | `false`: disconnect directly in login; `true`: login success -> disconnect in play (shows "Connection Lost") | `false`                                                                   |
| `MOTD`                        | MOTD in server status response                                                                                 | `§c§oMine§4§oMock§r\\n§6Minecraft mock server on golang§r | §eWelcomeO` |
| `VERSION_NAME`                | Displayed Minecraft version                                                                                    | `1.20.1`                                                                  |
| `PROTOCOL`                    | Protocol number used in status ping                                                                            | derived from `VERSION_NAME`                                               |
| `MAX_PLAYERS`                 | `players.max` in status response                                                                               | `20`                                                                      |
| `ONLINE_PLAYERS`              | `players.online` in status response                                                                            | `7`                                                                       |
| `REAL_SERVER_ADDR`            | Real Minecraft server address (`host:port`) for whitelisted users                                             | empty                                                                      |
| `LOGIN_WHITELIST`             | Comma/semicolon-separated usernames to proxy (example: `Steve,Alex`)                                          | empty                                                                      |

### `PROTOCOL` Note

If `PROTOCOL` is not set, it is auto-selected from `VERSION_NAME`.
If `VERSION_NAME` is unknown, fallback `763` is used.

### Whitelist Proxy Mode

When both `REAL_SERVER_ADDR` and `LOGIN_WHITELIST` are configured:

- usernames from whitelist are transparently proxied to the real server;
- all other users receive the configured login error (`ERROR`).

Username matching is case-insensitive.

## Project Structure

- `main.go` - entry point, logger setup, env config loading, server startup;
- `internal/config` - loading and parsing env-based configuration;
- `internal/server` - TCP server and handshake/status/login/proxy handling;
- `internal/protocol` - Minecraft packet encoding/decoding.
