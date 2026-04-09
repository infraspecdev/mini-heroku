# CLI Reference — `mini`

The `mini` CLI is the user-facing tool for the Mini Heroku platform. It lets you package a local application directory into a compressed archive and deploy it to a self-hosted controller server, with no Docker or server knowledge required on your end.

---

## Table of Contents

1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [Commands](#commands)
   - [mini version](#mini-version)
   - [mini config set-host](#mini-config-set-host)
   - [mini config get-host](#mini-config-get-host)
   - [mini config set-api-key](#mini-config-set-api-key)
   - [mini deploy](#mini-deploy)
   - [mini logs](#mini-logs)
4. [Internal Working](#internal-working)
   - [File Discovery](#file-discovery)
   - [Tarball Creation](#tarball-creation)
   - [HTTP Upload](#http-upload)
   - [Log Streaming](#log-streaming)
5. [Configuration File](#configuration-file)
6. [Design Decisions](#design-decisions)

---

## Installation

**macOS / Linux:**
```bash
curl -sL https://raw.githubusercontent.com/infraspecdev/mini-heroku/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/infraspecdev/mini-heroku/main/install.ps1 | iex
```

The installer detects your OS and architecture, downloads the correct binary, and places it in your `PATH`.

---

## Quick Start
```bash
# 1. Point the CLI at your controller
mini config set-host http://your-vm-ip:8080

# 2. Save your API key
mini config set-api-key your-secret-key

# 3. Deploy an app
mini deploy ./my-app my-app-name

# 4. Stream logs
mini logs my-app-name
```

---

## Commands

### `mini version`

Prints the CLI version string baked in at build time.
```bash
mini version
# mini version v1.2.0
```

**User level:** Run this to confirm the CLI installed correctly and to check which version you have.

**System level:** The version string is injected via `-ldflags` during the GitHub Actions release build, so it always reflects the Git tag that produced the binary.

---

### `mini config set-host`
```
mini config set-host <url>
```

Saves the controller server URL to your local config file.
```bash
mini config set-host http://192.168.1.100:8080
# Host set to: http://192.168.1.100:8080
```

| Argument | Required | Description |
|----------|----------|-------------|
| `<url>`  | Yes      | Full URL of the controller, including scheme and port |

**Validation:** The URL is validated with `url.ParseRequestURI` before saving. Malformed URLs are rejected immediately.
```bash
mini config set-host not-a-url
# Error: invalid URL "not-a-url": ...
```

---

### `mini config get-host`
```
mini config get-host
```

Prints the currently configured controller URL. If no host has been configured yet, prints a helpful prompt instead of an error.

---

### `mini config set-api-key`
```
mini config set-api-key <key>
```

Saves your API key to the OS Keychain. All mutating operations (`deploy`, `logs`) require a valid API key, validated server-side.
```bash
mini config set-api-key my-secret-key
# API key saved to OS Keychain
```

The key is securely stored in OS Keychain.

---

### `mini deploy`
```
mini deploy <folder> <app-name>
```

Packages the contents of `<folder>` and deploys them to the controller under `<app-name>`.
```bash
mini deploy ./my-python-app todo-api
```

| Argument     | Required | Description |
|--------------|----------|-------------|
| `<folder>`   | Yes      | Path to the local application directory |
| `<app-name>` | Yes      | Name to register the app under on the server |

**What it prints:**
```
Discovering files...
Found 8 files
Creating archive...
Archive size: 4312 bytes
Uploading to server...
Status : success
Message: App deployed successfully
App URL: http://todo-api.203.0.113.5.nip.io
```

**Error cases:**

| Situation | Output |
|-----------|--------|
| No host configured | Prints instructions to run `mini config set-host` |
| No API key configured | Returns an error immediately |
| Server rejects the build | Prints the server's error message |
| Folder does not exist | Returns a file-system error |

**Files automatically excluded from the archive:**

| Excluded | Reason |
|----------|--------|
| `.git/` | Version control metadata, not needed at runtime |
| `node_modules/` | Reconstructed from `package.json` inside the container |
| `__pycache__/` | Python bytecode, platform-specific |
| `.vscode/`, `.idea/` | Editor config, not application code |
| `.env` | Secrets — never shipped in an archive |

Everything else, including nested subdirectories, is included.

---

### `mini logs`
```
mini logs <app-name>
```

Streams live stdout/stderr from the running container. Blocks and prints continuously until the container exits or you press `Ctrl-C`.
```bash
mini logs todo-api
# === logs for todo-api (Ctrl-C to stop) ===
# 2024/01/01 00:00:01 server starting
# 2024/01/01 00:00:02 listening on :8080
```

| Argument     | Required | Description |
|--------------|----------|-------------|
| `<app-name>` | Yes      | Name of a previously deployed app |

**Error cases:**

| Situation | HTTP status | Output |
|-----------|-------------|--------|
| App not found in DB | 404 | Prints controller's error message |
| Container not running | 409 | Prints container state info |
| No API key | — | Error before any request is made |

---

## Internal Working

### File Discovery

`mini deploy` walks the target directory using `filepath.Walk`. For every path encountered:

1. If it is a **directory** whose name is in the ignore list, the entire subtree is skipped via `filepath.SkipDir`.
2. If it is a **file** whose name is in the file ignore list (`.env`), it is skipped.
3. Non-regular files (symlinks, devices) are skipped.
4. Everything else is recorded as a relative, forward-slash path.
```
my-app/
├── main.go          → included  (main.go)
├── handler.go       → included  (handler.go)
├── .env             → excluded
├── node_modules/    → excluded (entire tree)
└── src/
    └── utils.go     → included  (src/utils.go)
```

### Tarball Creation

Discovered file contents are assembled into an in-memory `tar.gz` archive:
```
Files on disk
    │
    ▼
tar.Writer  ← writes file headers + content
    │
gzip.Writer ← compresses the tar stream
    │
bytes.Buffer ← in-memory, no temp file
    │
    ▼
[]byte  → sent as the HTTP request body
```

Each entry gets `Name` (relative path), `Mode` (`0644`), and `Size`. Writers are closed in reverse order — `tar.Writer` first, then `gzip.Writer` — to flush all buffers before the bytes are read.

### HTTP Upload

The archive is sent as the raw body of a `POST /upload`:
```
POST /upload HTTP/1.1
Content-Type: application/x-gzip
App-Name: todo-api
X-API-Key: your-secret-key

<raw gzip bytes>
```

The server responds with JSON:
```json
{
  "status": "success",
  "appUrl": "http://todo-api.203.0.113.5.nip.io",
  "message": "App deployed successfully"
}
```

On non-200 responses, the CLI decodes the JSON `message` field for a readable error. If the body is not JSON, the raw text is shown instead.

### Log Streaming

`mini logs` opens a long-lived `GET /apps/<app-name>/logs` connection. The controller streams Docker output via chunked transfer encoding. The CLI copies the response body directly to `stdout` with `io.Copy` — no buffering, bytes appear as soon as the container produces them.

The request is bound to `cmd.Context()`, so `Ctrl-C` cancels the context and cleanly closes the connection.

---

## Configuration File

All settings are stored in `~/.mini/config.json`, created automatically on first use.
```json
{
  "server_url": "http://192.168.1.100:8080",
}
```

- **Permissions:** `0600` (owner read/write only)
- **Location:** resolved via `os.UserHomeDir()`, always `$HOME/.mini/config.json`

You can edit this file manually, but `mini config` subcommands are preferred since they validate before saving.

---

## Design Decisions

**Why a config file instead of environment variables?**
Environment variables are per-shell and per-session. A persistent file means you configure once and every subsequent command just works, even across terminals and reboots.

**Why raw body instead of `multipart/form-data`?**
`multipart/form-data` adds framing overhead and complicates the server-side reader. A raw body with `Content-Type: application/x-gzip` is unambiguous, streams cleanly with `io.Copy` on both sides, and requires no encoding/decoding on either end.

**Why build the archive in memory instead of writing a temp file?**
For typical application directories (a few MB), in-memory assembly avoids temp file management, cleanup logic, and disk I/O latency. If very large deployments become a concern, this is the first place to revisit.

**Why exclude `.env` automatically?**
Secrets should reach the container via environment injection at runtime, not baked into the image at build time. Automatic exclusion prevents accidental leakage even if the developer hasn't `.gitignore`d the file.

**Why `cobra` for the CLI framework?**
`cobra` is the standard Go CLI library (used by Docker, kubectl, Hugo). It handles help text, subcommand routing, argument validation, and context propagation without boilerplate, and its conventions are already familiar to most Go developers.