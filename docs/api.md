# API Reference — Controller

The Mini Heroku controller is an HTTP server running on port `:8080`. It exposes endpoints for deploying apps, streaming logs, managing proxy routes, and health checks. A separate reverse proxy runs on port `:80` and routes public traffic to deployed containers.

---

## Table of Contents

1. [Base URL](#base-url)
2. [Authentication](#authentication)
3. [Common Response Format](#common-response-format)
4. [Endpoints](#endpoints)
   - [POST /upload](#post-upload)
   - [GET /apps/:appName/logs](#get-appsappnamelogs)
   - [POST /register-route](#post-register-route)
   - [GET /health](#get-health)
   - [Reverse Proxy — :80](#reverse-proxy----80)
5. [Error Reference](#error-reference)
6. [Internal Working](#internal-working)
   - [Deploy Pipeline](#deploy-pipeline)
   - [Port Assignment](#port-assignment)
   - [Route Table](#route-table)
   - [Log Streaming Pipeline](#log-streaming-pipeline)
7. [Design Decisions](#design-decisions)

---

## Base URL

| Server | Address | Purpose |
|--------|---------|---------|
| Controller | `http://<vm-ip>:8080` | Deploy, logs, management |
| Proxy | `http://<vm-ip>:80` | Public traffic to apps |

The controller and proxy are separate HTTP servers started in the same process. The controller manages the platform; the proxy forwards end-user traffic.

---

## Authentication

All endpoints except `GET /health` require an API key passed in the request header:

```
X-API-Key: your-secret-key
```

The key is configured on the server via the `API_KEY` environment variable (loaded from `/etc/mini-heroku.env`). It is validated by the `RequireAPIKey` middleware, which wraps all protected routes at registration time in `main.go`.

**On missing or wrong key:**

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "status": "error",
  "message": "unauthorized: invalid or missing API key"
}
```

**Developer level:** Authentication is implemented as a middleware (`auth.RequireAPIKey`) that wraps an `http.Handler`. It reads the `X-API-Key` header, calls `AuthService.Validate`, and either calls `next.ServeHTTP` or short-circuits with a 401. Empty keys always fail, even if `API_KEY` is not set on the server — both sides must be non-empty strings that match exactly.

**System level:** A single shared secret is sufficient for a single-operator platform where the CLI is the only client. There are no user accounts, no token rotation, and no scopes — keeping the auth layer thin means less code to audit and less to go wrong.

---

## Common Response Format

All controller endpoints (except the proxy and the log stream) return JSON with `Content-Type: application/json`.

**Success shape:**

```json
{
  "status": "success",
  "appUrl": "http://my-app.203.0.113.5.nip.io",
  "message": "App deployed successfully"
}
```

**Error shape:**

```json
{
  "status": "error",
  "message": "Docker build failed: manifest not found"
}
```

`appUrl` is omitted on error responses. The `message` field always contains a human-readable description suitable for printing directly to a terminal.

---

## Endpoints

### POST /upload

Deploys an application. Accepts a gzip-compressed tar archive of the application source, builds a Docker image from it, starts a container, registers a proxy route, and returns the public URL.

**Authentication:** Required (`X-API-Key`)

#### Request

```http
POST /upload HTTP/1.1
Host: <vm-ip>:8080
Content-Type: application/x-gzip
App-Name: my-app
X-API-Key: your-secret-key

<raw .tar.gz bytes>
```

**Headers:**

| Header | Required | Description |
|--------|----------|-------------|
| `Content-Type` | Yes | Must be exactly `application/x-gzip` |
| `X-API-Key` | Yes | Server API key |
| `App-Name` | No | Name for the app. Auto-generated if omitted (`app-<pid>`) |

The request body is the raw tarball bytes — no multipart encoding. The tarball must contain a valid `Dockerfile` at its root so Docker can build the image.

#### Response — 200 OK

```json
{
  "status": "success",
  "appUrl": "http://my-app.203.0.113.5.nip.io",
  "message": "App deployed successfully"
}
```

The `appUrl` is constructed as `http://<app-name>.<VM_PUBLIC_IP>.nip.io`. [nip.io](https://nip.io) is a wildcard DNS service — no DNS configuration is required on your part.

#### Response — Error cases

| Condition | Status | Message |
|-----------|--------|---------|
| Wrong HTTP method | `405` | `Only POST allowed` |
| Wrong `Content-Type` | `400` | `Content-Type must be application/x-gzip` |
| Missing API key / wrong key | `401` | `unauthorized: invalid or missing API key` |
| Temp file creation failed | `500` | `Failed to create temp file: <reason>` |
| Body read/save failed | `500` | `Failed to save upload: <reason>` |
| Docker build failed | `500` | `Docker build failed: <reason>` |
| Container start failed | `500` | `Failed to start container: <reason>` |

#### Re-deploy behaviour

If `App-Name` matches an existing app, the old container is stopped and removed before the new one starts. The proxy route is updated in-place. Downtime is limited to the gap between stop and start of the new container.

---

### GET /apps/:appName/logs

Streams live stdout and stderr from the running container for `:appName`. The connection stays open until the container exits or the client disconnects.

**Authentication:** Required (`X-API-Key`)

#### Request

```http
GET /apps/my-app/logs HTTP/1.1
Host: <vm-ip>:8080
X-API-Key: your-secret-key
```

**Path parameter:**

| Parameter | Description |
|-----------|-------------|
| `:appName` | Name of a previously deployed app |

Expected path pattern: `/apps/<appName>/logs`. Paths that do not match this exact three-segment structure return `400`.

#### Response — 200 OK

```
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Cache-Control: no-cache

2024/01/01 00:00:01 server starting
2024/01/01 00:00:02 listening on :8080
2024/01/01 00:00:05 GET /health 200
```

The response body is a plain-text stream. Bytes are flushed to the client incrementally via `http.Flusher` — each write from Docker is forwarded immediately rather than buffered until the connection closes.

#### Response — Error cases

| Condition | Status | Body |
|-----------|--------|------|
| Non-GET method | `405` | JSON error |
| Malformed path | `400` | JSON error |
| App not found in DB | `404` | JSON error with app name |
| Container exists but stopped | `409` | JSON error with app name |
| Runner cannot stream logs | `500` | JSON error |

Note: once the `200` status and headers are written, the server cannot change the status code mid-stream. If Docker closes the connection unexpectedly after streaming has started, the client sees a truncated body rather than an error status.

---

### POST /register-route

Manually registers or updates a proxy route, mapping an app name to a backend container URL. This is used for internal tooling; in normal operation `POST /upload` registers routes automatically.

**Authentication:** Required (`X-API-Key`)

#### Request

```http
POST /register-route HTTP/1.1
Host: <vm-ip>:8080
Content-Type: application/json
X-API-Key: your-secret-key

{
  "app": "my-app",
  "target": "http://172.17.0.5:8080"
}
```

**Body fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `app` | string | Yes | App name used in subdomain routing |
| `target` | string | Yes | Full URL of the container (internal IP + port) |

#### Response — 200 OK

```json
{
  "status": "ok",
  "app": "my-app",
  "target": "http://172.17.0.5:8080"
}
```

#### Response — Error cases

| Condition | Status | Body |
|-----------|--------|------|
| Non-POST method | `405` | Plain text error |
| Invalid JSON body | `400` | Plain text error |
| Missing `app` or `target` | `400` | Plain text error |

---

### GET /health

Returns the server's health status. Used by load balancers, uptime monitors, and deployment pipelines to verify the controller is reachable.

**Authentication:** None — this endpoint is intentionally public.

#### Request

```http
GET /health HTTP/1.1
Host: <vm-ip>:8080
```

#### Response — 200 OK

```json
{
  "status": "ok"
}
```

This endpoint always returns `200` if the process is running. It does not check Docker connectivity or database health — it is a liveness check only.

---

### Reverse Proxy — :80

The proxy server listens on port `:80` and routes incoming requests to the correct container based on the subdomain of the `Host` header.

**This server has no authentication.** It is the public-facing entry point for deployed applications.

#### How routing works

```
Request: GET /api/users
Host: my-app.203.0.113.5.nip.io

→ extracts subdomain: "my-app"
→ looks up route table: "my-app" → "http://172.17.0.5:8080"
→ forwards full request to container
```

The app name is always the first label of the hostname, everything before the first `.`. Port numbers in the `Host` header (e.g. `my-app.1.2.3.4.nip.io:80`) are stripped before extraction.

#### Proxy error responses

| Condition | Status | Body |
|-----------|--------|------|
| Malformed or empty `Host` header | `400` | `{"error": "invalid host header: ..."}` |
| App name not in route table | `404` | `{"error": "app \"my-app\" not found"}` |
| Container unreachable | `502` | `{"error": "could not reach container"}` |

---

## Error Reference

All controller errors follow the same JSON envelope. Below is a consolidated reference across all endpoints.

| Status | Meaning | Common causes |
|--------|---------|---------------|
| `400` | Bad Request | Wrong `Content-Type`, malformed path, invalid JSON body |
| `401` | Unauthorized | Missing or incorrect `X-API-Key` |
| `404` | Not Found | App name not in database; unregistered controller path |
| `405` | Method Not Allowed | Using GET on `/upload`, POST on `/logs`, etc. |
| `409` | Conflict | Container exists in DB but Docker reports it stopped |
| `500` | Internal Server Error | Docker build error, container start failure, I/O error |
| `502` | Bad Gateway | Container is registered but not accepting connections (proxy only) |

---

## Internal Working

### Deploy Pipeline

When `POST /upload` receives a request, the following sequence executes synchronously before a response is returned:

```
Client (mini CLI)
    │
    │  POST /upload
    │  Body: <tar.gz bytes>
    ▼
[1] Save body to temp file
    /tmp/<app-name>.tar.gz
    │
    ▼
[2] builder.BuildImage()
    Sends tarball to Docker daemon as build context
    Streams build output (logs each line via zerolog)
    Returns image name: "<app-name>:latest"
    │
    ▼
[3] Stop + remove old container (if re-deploy)
    ContainerStop  → ContainerRemove
    │
    ▼
[4] runner.GenerateHostPort(appName)
    Deterministic port from name hash: 10000–19999
    │
    ▼
[5] runner.RunContainer()
    ContainerCreate → ContainerStart → ContainerInspect
    Returns ContainerID, ContainerIP, HostPort
    │
    ▼
[6] table.Register(appName, "http://<ContainerIP>:8080")
    In-memory route table updated (thread-safe)
    │
    ▼
[7] db.Upsert(project)
    Persists container state to SQLite
    Non-fatal: app is running even if this fails
    │
    ▼
[8] sendSuccess(w, publicURL, "App deployed successfully")
    publicURL = "http://<appName>.<VM_PUBLIC_IP>.nip.io"
```

The entire pipeline runs in the HTTP handler goroutine. The request does not return until Docker finishes building and the container is confirmed running. For large images this can take tens of seconds — the CLI should not set a short timeout.

### Port Assignment

Each app gets a deterministic host port derived from its name:

```go
func GenerateHostPort(appName string) int {
    hash := 0
    for _, c := range appName {
        hash += int(c)
    }
    return 10000 + (hash % 10000)
}
```

The port range is `10000–19999`. The same app name always produces the same port, which means re-deploys reuse the same host port without needing to persist or look up the previous value. The port is still saved to the database so the reconciler can re-bind it on server restart without recalculating.

### Route Table

The in-memory `RouteTable` is a `sync.RWMutex`-protected `map[string]string`. It maps app name to container target URL:

```
"my-app"   → "http://172.17.0.5:8080"
"todo-api" → "http://172.17.0.6:8080"
```

On server restart, the `reconcile()` function in `main.go` rebuilds this table from the SQLite database before the HTTP servers start accepting traffic. Each live container is re-inspected via Docker to get its current IP, and the route is registered. Stopped containers are restarted; missing containers are recreated from the saved image name.

### Log Streaming Pipeline

```
Client (mini logs)
    │
    │  GET /apps/my-app/logs (HTTP/1.1, keep-alive)
    ▼
LogsHandler
    │
    ├─ db.GetByName(appName) → ContainerID
    ├─ ContainerInspect      → verify Running == true
    │
    ▼
RealRunnerClient.StreamLogs()
    │
    │  docker.ContainerLogs(ctx, containerID, Follow=true)
    │  Returns io.ReadCloser (Docker multiplexed stream)
    │
    ▼
stdcopy.StdCopy(dst, dst, logReader)
    Demultiplexes Docker's stream format
    Writes stdout + stderr to the same flushedWriter
    │
    ▼
flushedWriter.Write()
    Writes chunk to http.ResponseWriter
    Calls http.Flusher.Flush() after every write
    │
    ▼
Client receives log lines in real time
```

Docker's `ContainerLogs` returns a multiplexed binary stream where each frame is prefixed with an 8-byte header encoding the stream type (stdout/stderr) and payload length. `stdcopy.StdCopy` strips these headers and writes the raw text to the destination writer. This is why the output from `mini logs` is clean text rather than binary-framed data.

---

## Design Decisions

**Why is the deploy endpoint synchronous?**
The client (the CLI) needs to know whether the deployment succeeded before it can print the app URL. An async deploy would require polling or webhooks, which adds complexity for no real benefit given that the CLI is already blocking and waiting for the response. The tradeoff is that slow Docker builds tie up the HTTP handler — acceptable for a single-operator tool.

**Why store state in SQLite instead of only in-memory?**
The in-memory route table is rebuilt from the database on every restart. Without persistence, a controller restart would require every app to be redeployed manually. SQLite provides durability with zero operational overhead — no external database process, no connection pooling, no migrations beyond `AutoMigrate`.

**Why is the DB upsert non-fatal?**
The container is already running by the time the upsert happens. Failing the entire request and returning a 500 — after the container has started and the proxy route is live — would be misleading. The app is accessible; only the persistent record is missing. A warning log is emitted instead so the operator is aware.

**Why use nip.io for public URLs?**
nip.io maps `<anything>.<ip>.nip.io` to `<ip>` via DNS. This gives every app a unique public subdomain with zero DNS configuration. The proxy then uses the subdomain to look up the target container. The entire routing chain works on a fresh VM with no domain name or DNS zone required.

**Why does the proxy run on `:80` instead of `:8080`?**
Port `80` is the conventional HTTP port and requires no port number in the URL. Apps are reachable at `http://my-app.1.2.3.4.nip.io` rather than `http://my-app.1.2.3.4.nip.io:8080`. The controller on `:8080` is an internal management API, not user-facing traffic.

**Why does `/health` require no auth?**
Health checks are performed by infrastructure tooling (systemd watchdogs, uptime monitors, load balancers) that may not have the API key. A health endpoint gated behind auth would make infrastructure integration harder for no security benefit — knowing the controller process is alive is not sensitive information.
