# Troubleshooting Guide

This guide covers the most common failures across the CLI, controller API, Docker pipeline, and infrastructure layer. Each issue includes the exact error message you will see, the root cause, and the specific commands to diagnose and fix it.

---

## Table of Contents

1. [How to Read Logs](#how-to-read-logs)
2. [CLI Issues](#cli-issues)
   - [Controller host not configured](#controller-host-not-configured)
   - [No API key configured](#no-api-key-configured)
   - [deployment failed: controller returned 401](#deployment-failed-controller-returned-401)
   - [deployment failed: controller returned 500](#deployment-failed-controller-returned-500)
   - [connecting to controller: dial tcp … connection refused](#connecting-to-controller-dial-tcp--connection-refused)
   - [exploring directory: …no such file or directory](#exploring-directory-no-such-file-or-directory)
   - [Found 0 files](#found-0-files)
   - [log stream interrupted](#log-stream-interrupted)
3. [API Errors](#api-errors)
   - [401 Unauthorized](#401-unauthorized)
   - [400 Content-Type must be application/x-gzip](#400-content-type-must-be-applicationx-gzip)
   - [500 Docker build failed](#500-docker-build-failed)
   - [500 Failed to start container](#500-failed-to-start-container)
   - [404 app not found](#404-app-not-found)
   - [409 container is not running](#409-container-is-not-running)
   - [502 could not reach container](#502-could-not-reach-container)
4. [Common Failures](#common-failures)
   - [App deployed successfully but URL returns 404](#app-deployed-successfully-but-url-returns-404)
   - [App URL is unreachable from the browser](#app-url-is-unreachable-from-the-browser)
   - [Re-deploy succeeds but app still shows old version](#re-deploy-succeeds-but-app-still-shows-old-version)
   - [Controller crashes on startup](#controller-crashes-on-startup)
   - [All apps disappear after controller restart](#all-apps-disappear-after-controller-restart)
   - [Port conflict on deploy](#port-conflict-on-deploy)
5. [Debug Steps](#debug-steps)
   - [Check controller health](#check-controller-health)
   - [Inspect the controller log](#inspect-the-controller-log)
   - [Inspect a running container](#inspect-a-running-container)
   - [Verify the route table](#verify-the-route-table)
   - [Query the SQLite database](#query-the-sqlite-database)
   - [Check the systemd service](#check-the-systemd-service)
   - [Verify environment variables](#verify-environment-variables)
   - [Full deployment health checklist](#full-deployment-health-checklist)

---

## How to Read Logs

The controller writes structured JSON logs via zerolog. Every line is a JSON object with at minimum `time`, `level`, `service`, and `message` fields. App-specific logs also carry an `app` field.

**Controller log file:**
```bash
tail -f /opt/mini-heroku/logs/controller.log
```

**systemd journal (catches startup crashes before the log file opens):**
```bash
sudo journalctl -u mini-heroku -f
# or last 50 lines without following:
sudo journalctl -u mini-heroku -n 50 --no-pager
```

**Pretty-print JSON logs with `jq`:**
```bash
tail -f /opt/mini-heroku/logs/controller.log | jq '.'
# filter to a specific app only:
tail -f /opt/mini-heroku/logs/controller.log | jq 'select(.app == "my-app")'
# filter to errors only:
tail -f /opt/mini-heroku/logs/controller.log | jq 'select(.level == "error")'
```

---

## CLI Issues

### Controller host not configured

**Symptom:**
```
Controller host not configured
Run: mini config set-host <url>
```

**Cause:** `~/.mini/config.json` either does not exist or has an empty `server_url` field.

**Fix:**
```bash
mini config set-host http://<your-vm-ip>:8080
# Verify:
mini config get-host
```

---

### No API key configured

**Symptom:**
```
Error: no API key configured — run: mini config set-api-key <key>
```



**Fix:**
```bash
mini config set-api-key <your-key>
```

To verify the key was saved:
Check the OS keychain or windows credential.


---

### deployment failed: controller returned 401

**Symptom:**
```
Error: deployment failed: controller returned 401: unauthorized: invalid or missing API key
```

**Cause:** The API key in OS Keychain does not match the `API_KEY` value in `/etc/mini-heroku.env` on the server.

**Fix — client side:**
```bash
mini config set-api-key <correct-key>
```

**Fix — server side (if you need to check what key the server expects):**
```bash
# SSH into the VM, then:
sudo cat /etc/mini-heroku.env
```

If `API_KEY` is blank or missing in that file, update it and restart the service:
```bash
sudo nano /etc/mini-heroku.env
# Add or correct: API_KEY=your-secret-key
sudo systemctl restart mini-heroku
```

---

### deployment failed: controller returned 500

**Symptom:**
```
Error: deployment failed: controller returned 500: Docker build failed: <reason>
```

The `<reason>` part is the raw Docker error. Common values and their fixes are covered in [500 Docker build failed](#500-docker-build-failed) below.

---

### connecting to controller: dial tcp … connection refused

**Symptom:**
```
Error: connecting to controller: dial tcp 192.168.1.100:8080: connect: connection refused
```

**Cause:** The controller process is not running, or the host/port in your config is wrong.

**Debug steps:**

1. Verify the address is correct:
   ```bash
   mini config get-host
   ```

2. Check if the controller is reachable at all:
   ```bash
   curl http://<vm-ip>:8080/health
   # Expected: {"status":"ok"}
   ```

3. SSH into the VM and check the service:
   ```bash
   sudo systemctl status mini-heroku
   ```

4. If stopped, check why it crashed:
   ```bash
   sudo journalctl -u mini-heroku -n 50 --no-pager
   ```

5. If the service is active but port 8080 is not open, check firewall rules:
   ```bash
   sudo ufw status
   # or:
   sudo iptables -L INPUT -n | grep 8080
   ```

---

### exploring directory: …no such file or directory

**Symptom:**
```
Error: exploring directory: lstat ./my-app: no such file or directory
```

**Cause:** The folder path passed to `mini deploy` does not exist on your local machine.

**Fix:** Confirm the path exists and use either an absolute path or a correct relative path:
```bash
ls ./my-app
mini deploy ./my-app my-app-name
# or with absolute path:
mini deploy /home/user/projects/my-app my-app-name
```

---

### Found 0 files

**Symptom:**
```
Discovering files...
Found 0 files
Creating archive...
Archive size: 22 bytes
```

The deploy continues but the server receives an essentially empty tarball and the Docker build will fail with no `Dockerfile` found.

**Cause:** The target folder only contains files/directories in the ignore list (`.git`, `node_modules`, `.env`, etc.), or the folder itself is empty.

**Fix:** Verify what the packager sees:
```bash
ls -la ./my-app
```

Check that you have at least a `Dockerfile` and your source files in the folder root and not inside an ignored subdirectory. The complete ignore list is: `.git/`, `node_modules/`, `__pycache__/`, `.vscode/`, `.idea/`, `.env`.

---

### log stream interrupted

**Symptom:**
```
Error: log stream interrupted: read tcp ...
```

**Cause:** The TCP connection to the controller was dropped mid-stream — typically because the controller process restarted, the VM lost network connectivity, or an idle connection was killed by a network timeout/firewall.

**Fix:** Simply re-run `mini logs <app-name>`. If the interruption keeps happening, check whether the controller is restarting (see [Check the systemd service](#check-the-systemd-service)).

---

## API Errors

### 401 Unauthorized

```json
{
  "status": "error",
  "message": "unauthorized: invalid or missing API key"
}
```

**Cause:** The `X-API-Key` header is absent, empty, or does not match the server's `API_KEY` environment variable. Note that if `API_KEY` is not set on the server, **all requests fail** — an empty server key never matches anything.

**Debug:**
```bash
# Verify the server has a key set:
sudo grep API_KEY /etc/mini-heroku.env

# Test with curl:
curl -H "X-API-Key: your-key" http://<vm-ip>:8080/health
```

---

### 400 Content-Type must be application/x-gzip

```json
{
  "status": "error",
  "message": "Content-Type must be application/x-gzip"
}
```

**Cause:** A request to `POST /upload` was made without `Content-Type: application/x-gzip`. If you are using `mini deploy` this should never happen — it is a sign of a direct API call with the wrong header, or a proxy stripping/rewriting headers.

**Fix:** Ensure `Content-Type: application/x-gzip` is set on the request. If using curl:
```bash
curl -X POST \
  -H "Content-Type: application/x-gzip" \
  -H "X-API-Key: your-key" \
  -H "App-Name: my-app" \
  --data-binary @my-app.tar.gz \
  http://<vm-ip>:8080/upload
```

---

### 500 Docker build failed

```json
{
  "status": "error",
  "message": "Docker build failed: <docker error>"
}
```

The `<docker error>` tells you exactly what went wrong inside Docker. Common cases:

**`dockerfile: no such file or directory` / `failed to read dockerfile`**

Your tarball has no `Dockerfile` at the root level. Check your project structure:
```bash
# List what's in your tarball:
tar -tzf /tmp/<app-name>.tar.gz | head -20
```
The `Dockerfile` must be at the top level of the archive, not inside a subdirectory.

**`manifest unknown` / `manifest not found`**

A `FROM` instruction in your `Dockerfile` references an image that does not exist or cannot be pulled. Check the image name and tag:
```bash
# On the VM, test pulling the base image directly:
docker pull <base-image-name>
```

**`exec format error`**

Your `Dockerfile` copies or runs a binary compiled for the wrong architecture (e.g. a macOS ARM binary on a Linux AMD64 server). Ensure your build produces a `linux/amd64` binary.

**`permission denied`**

The `miniheroku` service user does not have access to the Docker daemon socket:
```bash
# Check docker group membership:
groups miniheroku
# Should include 'docker'. If not:
sudo usermod -aG docker miniheroku
sudo systemctl restart mini-heroku
```

---

### 500 Failed to start container

```json
{
  "status": "error",
  "message": "Failed to start container: starting container: <reason>"
}
```

**`port is already allocated`**

The host port computed for this app name is already bound by another process. Since port assignment is deterministic (`10000 + hash(name) % 10000`), two different app names can hash to the same port in theory, but more commonly the port is held by a previous container that was not cleaned up.

```bash
# Find what's using the port:
sudo ss -tlnp | grep <port>
# or:
sudo lsof -i :<port>

# If it's a zombie container:
docker ps -a | grep <port>
docker rm -f <container-id>
```

Then re-deploy the app.

**`no such image`**

The Docker build step succeeded but the image name returned does not exist in the local daemon. This is a race condition or a daemon-side issue. Check Docker daemon health:
```bash
sudo systemctl status docker
docker images | grep <app-name>
```

---

### 404 app not found

```json
{
  "status": "error",
  "message": "app not found: my-app"
}
```

Returned by `GET /apps/:appName/logs`.

**Cause:** The app name does not exist in the SQLite database. Either the app was never deployed, the name is misspelled, or the database file was deleted/corrupted.

**Debug:**
```bash
# Check what apps are in the database:
sqlite3 /opt/mini-heroku/data/mini.db "SELECT name, status FROM projects;"
```

If the app is missing, it needs to be redeployed:
```bash
mini deploy ./my-app my-app
```

---

### 409 container is not running

```json
{
  "status": "error",
  "message": "container is not running for app: my-app"
}
```

Returned by `GET /apps/:appName/logs`.

**Cause:** The app exists in the database but Docker reports its container as stopped or exited.

**Debug:**
```bash
# Get the container ID from the database:
sqlite3 /opt/mini-heroku/data/mini.db \
  "SELECT name, container_id, status FROM projects WHERE name='my-app';"

# Check the container state:
docker inspect <container-id> | jq '.[0].State'

# See why it exited:
docker logs <container-id> --tail 50
```

**Fix options:**

If the container crashed due to an application error, fix the app and redeploy:
```bash
mini deploy ./my-app my-app
```

If it stopped for an external reason and you want to restart it in-place:
```bash
docker start <container-id>
```

Note: the controller's reconciler will also attempt to restart stopped containers on the next service restart.

---

### 502 could not reach container

```json
{"error": "could not reach container"}
```

Returned by the **proxy** on port `:80`, not the controller.

**Cause:** The proxy has a route registered for this app, but the HTTP request to the container's internal IP and port timed out or was refused. The container may be starting up, crashed, or its internal port (`8080`) is not actually listening.

**Debug:**
```bash
# 1. Check if the container is running:
docker ps | grep <app-name>

# 2. Check the container logs for startup errors:
mini logs <app-name>
# or directly:
docker logs <container-id> --tail 50

# 3. Confirm the app inside the container is listening on port 8080:
docker exec <container-id> ss -tlnp | grep 8080
# or:
docker exec <container-id> curl -s http://localhost:8080/health
```

Your app **must** listen on port `8080` inside the container — this is the port the proxy forwards to. If your app listens on a different port, add an explicit `EXPOSE 8080` and update the app's listen address.

---

## Common Failures

### App deployed successfully but URL returns 404

**Symptom:** `mini deploy` reports success and gives a URL, but visiting that URL returns a 404 from the proxy.

**Cause:** One of two things — either the route was not registered in the proxy's in-memory table, or the `Host` header in your browser request does not match what the proxy expects.

**Debug:**
```bash
# 1. Verify the proxy is routing correctly by hitting with explicit Host:
curl -H "Host: my-app.<vm-ip>.nip.io" http://<vm-ip>/

# 2. Check if the nip.io DNS resolution is working:
nslookup my-app.203.0.113.5.nip.io
# Should resolve to 203.0.113.5

# 3. Check the controller log for route registration:
grep "route registered" /opt/mini-heroku/logs/controller.log | grep "my-app"
```

If the route was never logged as registered, the deploy pipeline likely failed partway through after the container started but before `table.Register` was called. Redeploying should fix it.

---

### App URL is unreachable from the browser

**Symptom:** The app URL resolves in DNS but the browser times out or refuses the connection.

**Cause:** Port `80` on the VM is blocked by a firewall or security group rule.

**Debug:**
```bash
# From your local machine:
nc -zv <vm-ip> 80
# or:
curl -v http://my-app.<vm-ip>.nip.io

# On the VM, check if port 80 is listening:
sudo ss -tlnp | grep :80

# Check firewall:
sudo ufw status
```

**Fix:** Open port 80 in your cloud provider's security group/firewall and in `ufw` if active:
```bash
sudo ufw allow 80/tcp
```

---

### Re-deploy succeeds but app still shows old version

**Symptom:** `mini deploy` returns success, the message says the app was deployed, but requests to the app URL still serve the previous version.

**Cause:** Usually a browser or CDN cache. Occasionally, the old container was not successfully removed and the proxy route was not updated.

**Debug:**
```bash
# 1. Hard-refresh in the browser (Ctrl+Shift+R / Cmd+Shift+R).

# 2. Check which container is actually running:
docker ps | grep <app-name>

# 3. Check the image it's using was rebuilt:
docker inspect <container-id> | jq '.[0].Config.Image'
# Should show the image tagged with your app name

# 4. Check the controller log for the re-deploy sequence:
grep -A5 "stopping old container" /opt/mini-heroku/logs/controller.log | grep "my-app"
```

If the old container was not removed (it will appear twice in `docker ps`), stop and remove it manually then re-deploy:
```bash
docker ps | grep <app-name>
docker stop <old-container-id>
docker rm <old-container-id>
mini deploy ./my-app my-app
```

---

### Controller crashes on startup

**Symptom:** `sudo systemctl status mini-heroku` shows `failed` or `activating` → `failed`.

**Most common causes and fixes:**

**Docker daemon is not running:**
```bash
sudo systemctl status docker
sudo systemctl start docker
sudo systemctl restart mini-heroku
```

**`miniheroku` user is not in the `docker` group:**
```bash
groups miniheroku
# If 'docker' is missing:
sudo usermod -aG docker miniheroku
# Group changes only take effect on next login/service start:
sudo systemctl restart mini-heroku
```

**Missing or unreadable environment file:**
```bash
sudo ls -la /etc/mini-heroku.env
# Should exist and be readable by root/miniheroku
sudo cat /etc/mini-heroku.env
```

**Database directory not writable:**

The controller opens `/opt/mini-heroku/data/mini.db` on startup. If the directory does not exist or is owned by the wrong user, `NewStore` fails fatally.
```bash
sudo ls -la /opt/mini-heroku/data/
sudo chown -R miniheroku:miniheroku /opt/mini-heroku/data
sudo systemctl restart mini-heroku
```

**Port 80 or 8080 already in use:**
```bash
sudo ss -tlnp | grep -E ':80|:8080'
# Kill whatever is holding the port, then:
sudo systemctl restart mini-heroku
```

---

### All apps disappear after controller restart

**Symptom:** After a controller restart, `mini logs <app>` returns 404 and no app URLs work.

**Cause:** The SQLite database file was deleted, the path changed, or the reconciler failed silently.

**Debug:**
```bash
# Check the DB exists and has rows:
ls -lh /opt/mini-heroku/data/mini.db
sqlite3 /opt/mini-heroku/data/mini.db "SELECT count(*) FROM projects;"

# Check reconciler output at startup:
sudo journalctl -u mini-heroku --since "10 minutes ago" | grep reconcil
```

If the DB exists and has rows but apps are still gone from the proxy, check for reconciler errors in the log — a Docker permission error during `ContainerInspect` would cause the reconciler to skip all apps without registering routes.

If the DB was deleted, all apps need to be redeployed. There is no other source of truth.

---

### Port conflict on deploy

**Symptom:**
```
Error: deployment failed: controller returned 500: Failed to start container: starting container: driver failed programming external connectivity: port is already allocated
```

**Cause:** Port assignment is deterministic — `10000 + (sum of ASCII values of app name) % 10000`. Occasionally two app names hash to the same port, or an old container is still holding the port from a failed previous deploy.

**Fix:**
```bash
# Find the conflicting port number first:
# (calculate it yourself: sum the ASCII values of your app name chars, mod 10000, add 10000)
# Or just find what's on the expected range:
docker ps --format "table {{.Names}}\t{{.Ports}}" | grep 1[0-9][0-9][0-9][0-9]

# Remove the zombie container if it belongs to your app:
docker rm -f <container-id>

# Then redeploy:
mini deploy ./my-app my-app
```

If two different apps genuinely hash to the same port, rename one of them slightly — even one character difference changes the hash.

---

## Debug Steps

A set of reusable commands to run when you need to inspect system state.

### Check controller health

The fastest way to confirm the controller is alive and reachable:

```bash
curl -s http://<vm-ip>:8080/health | jq '.'
# Expected: {"status":"ok"}
```

If this fails, the controller is down. See [Controller crashes on startup](#controller-crashes-on-startup).

---

### Inspect the controller log

```bash
# Live tail:
tail -f /opt/mini-heroku/logs/controller.log

# Last 100 lines, pretty-printed:
tail -100 /opt/mini-heroku/logs/controller.log | jq '.'

# Filter to a specific app:
tail -f /opt/mini-heroku/logs/controller.log | jq 'select(.app == "my-app")'

# Filter to errors and warnings only:
tail -f /opt/mini-heroku/logs/controller.log | jq 'select(.level | IN("error","warn"))'

# Show log entries from the last deploy of an app:
grep '"my-app"' /opt/mini-heroku/logs/controller.log | tail -20 | jq '.'
```

---

### Inspect a running container

```bash
# List all containers related to mini-heroku apps:
docker ps --format "table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"

# Get full details for a specific container:
docker inspect <container-id> | jq '.[0] | {
  State: .State,
  IP: .NetworkSettings.IPAddress,
  Ports: .NetworkSettings.Ports
}'

# Live logs from a container (equivalent to mini logs):
docker logs -f <container-id>

# Tail last 50 lines:
docker logs --tail 50 <container-id>

# Open a shell inside a running container (for deep inspection):
docker exec -it <container-id> /bin/sh
```

---

### Verify the route table

There is no dedicated endpoint to dump the route table, but you can infer its state from the controller log and from live proxy responses:

```bash
# Check which routes were registered at startup (reconcile):
grep "route registered" /opt/mini-heroku/logs/controller.log | jq '{app: .app, target: .target}'

# Test the proxy route for a specific app directly:
curl -v -H "Host: my-app.<vm-ip>.nip.io" http://<vm-ip>/
# A 502 means the route exists but the container is unreachable.
# A 404 means the route is not registered at all.
```

---

### Query the SQLite database

```bash
# Open the database:
sqlite3 /opt/mini-heroku/data/mini.db

# Inside sqlite3:
.headers on
.mode column

# List all apps:
SELECT name, status, host_port, container_ip, updated_at FROM projects;

# Check a specific app:
SELECT * FROM projects WHERE name = 'my-app';

# Exit:
.quit
```

Or as a one-liner:
```bash
sqlite3 /opt/mini-heroku/data/mini.db \
  "SELECT name, status, host_port, container_ip FROM projects;"
```

---

### Check the systemd service

```bash
# Current status:
sudo systemctl status mini-heroku

# Full logs since last start:
sudo journalctl -u mini-heroku -n 100 --no-pager

# Follow live (useful during a restart):
sudo journalctl -u mini-heroku -f

# Restart:
sudo systemctl restart mini-heroku

# Check if it's set to start on boot:
sudo systemctl is-enabled mini-heroku
```

---

### Verify environment variables

```bash
# View the env file (contains API_KEY, VM_PUBLIC_IP, BASE_URL):
sudo cat /etc/mini-heroku.env

# Confirm the running process actually received them:
sudo cat /proc/$(pgrep controller)/environ | tr '\0' '\n' | grep -E 'API_KEY|VM_PUBLIC_IP|BASE_URL'
```

If `API_KEY` appears blank in the env file, update it and restart:
```bash
sudo nano /etc/mini-heroku.env
sudo systemctl restart mini-heroku
```

---

### Full deployment health checklist

Run through this list top-to-bottom when a deployment is not behaving as expected:

```bash
# 1. Controller is alive:
curl -s http://<vm-ip>:8080/health

# 2. Service is running:
sudo systemctl is-active mini-heroku

# 3. Docker is running and accessible by the service user:
sudo -u miniheroku docker ps

# 4. Environment variables are set:
sudo grep -c 'API_KEY' /etc/mini-heroku.env

# 5. Database file exists:
ls -lh /opt/mini-heroku/data/mini.db

# 6. App exists in database:
sqlite3 /opt/mini-heroku/data/mini.db "SELECT name, status FROM projects;"

# 7. Container is running:
docker ps | grep <app-name>

# 8. Route was registered (check log):
grep "route registered" /opt/mini-heroku/logs/controller.log | grep <app-name>

# 9. Proxy reaches the container:
curl -s -H "Host: <app-name>.<vm-ip>.nip.io" http://<vm-ip>/
```

All 9 checks passing means the system is healthy end-to-end.
