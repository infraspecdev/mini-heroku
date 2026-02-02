# Detailed Weekly Plan: Mini Heroku (PaaS)

This document provides a granular breakdown of the 8-week internship project. It includes specific learning resources, task divisions for Mentors and Mentees, and guidelines on how work should be assigned.

## General Workflow
*   **Monday:** Sprint Planning. Review the week's goals. Mentor provides a high-level overview of concepts. (Mentor's 1st weekly Sync, 30 mins)
*   **Tue-Thu:** Implementation. Async communication over chat to flag blockers. (Mentor's buddy can help with this)
*   **Friday:** Code Review & Demo. Show what was built to the mentor. (Mentor's 2nd Weekly Sync, 30 mins)

## Phase 1: Getting Started with Docker (Weeks 1-2)

### Week 1: Project Alignment & Docker Fundamentals
**Goal:** Understand the PaaS landscape, define the project architecture, and control the Docker Daemon programmatically using Go.

*   **Learning Resources:**
    *   [How Heroku Works](https://devcenter.heroku.com/articles/how-heroku-works)
    *   [System Design Primer: Proxies & Reverse Proxies](https://github.com/donnemartin/system-design-primer#reverse-proxy)
    *   [Docker Engine API SDK for Go](https://pkg.go.dev/github.com/docker/docker/client)
    *   [Understanding Docker Socket](https://dev.to/piyushbagani15/understanding-varrundockersock-the-key-to-dockers-inner-workings-nm7)
    *   [How CLI's work](https://www.contentful.com/blog/command-line-interfaces-explained/)

*   **Tasks:**
    *   **Mentor:**
        *   Deep dive into the Problem Statement: What is a PaaS? Why Go?
        *   Explain the core components: CLI, Controller, and Proxy.
        *   Verify Mentee's machine environment (Go + Docker Desktop/Engine installed).
    *   **Mentee:**
        *   **Task 1.1: Project Analysis.** Research how Heroku/Vercel handles "git push" deployments and compare it to our approach.
        *   **Task 1.2: SDK Basics.** Initialize Go module (`go mod init`) and write a script to connect to the Docker client and print the Docker version.
        *   **Task 1.3: Container Lifecycle.** Write functions to list running containers and pull/run an `alpine` container via the SDK.

*   **Task Assignment:**
    *   Tasks 1.1 - 1.3: Collaborative. Both mentees work together on the research and design document.
    *   Tasks 1.4 - 1.5: Split individually. One mentee focuses on "Listing/Inspection", the other on "Lifecycle (Start/Stop)".

### Week 2: The Builder (Core Build Pipeline)
**Goal:** Implement the "Build Pipeline"—the logic that turns local source code into a Docker Image.

*   **Learning Resources:**
    *   [Go `archive/tar` Package](https://pkg.go.dev/archive/tar)
    *   [Docker ImageBuild API Reference](https://docs.docker.com/engine/api/v1.43/#tag/Image/operation/ImageBuild)
    *   [Building Docker Images with Go](https://docs.docker.com/guides/golang/build-images/)

*   **Tasks:**
    *   **Mentor:**
        *   Explain "Build Context" in Docker: Why do we need to send the whole folder?
        *   Review the Tar creation logic (a common source of bugs: ensuring relative paths are correct).
        *   Clarify that we are building the *logic* this week, not the polished `mini` CLI command yet.
    *   **Mentee:**
        *   **Task 2.1: The Test Subject (Sample App).** Create a "dummy" project (e.g., `test-app/` containing a `main.go` web server and a `Dockerfile`) to use for testing the pipeline.
        *   **Task 2.2: The Courier (Client Logic).** Implement a Go function `TarFolder(path)` that recursively reads a directory and returns an `io.Reader` (a stream of the tar archive). *Note: This will eventually go into the CLI.*
        *   **Task 2.3: The Receiver (Server Logic).** Create a simple HTTP Server with a `POST /upload` endpoint. It should read the incoming tar stream.
        *   **Task 2.4: The Factory (Docker Integration).** Connect the `POST /upload` handler to the Docker SDK's `ImageBuild()` function.
            *   *Success Criteria:* When the client runs, a new image appears in `docker images`.

*   **Task Assignment:**
    *   Mentee A: **The Courier.** Focuses on File I/O, recursion, and creating a valid Tar stream.
    *   Mentee B: **The Factory.** Focuses on the HTTP Server and strictly typing the Docker `ImageBuild` options.
    *   **Integration:** On Thursday, connect Mentee A's "Tar Stream" into Mentee B's "Upload Handler".

## Phase 2: Cloud & Networking (Weeks 3-5)

### Week 3: Cloud Migration & CI/CD
**Goal:** Move the Controller from "Localhost" to a real Cloud VM.

**Learning Resources**
* **DevOps:** [GitHub Actions for Go](https://github.com/features/actions)
* **Linux:** [Systemd Service Guide](https://www.digitalocean.com/community/tutorials/how-to-use-systemctl-to-manage-systemd-services-and-units)

**Mentee Tasks**
* **Task 3.1: VM Setup (Mentee A)**
    * *Component:* **Infrastructure**
    * Provision Cloud VM AWS. Install Docker, Go. Secure Firewall (Ports 22, 80, 443).
* **Task 3.2: CI/CD Pipeline (Mentee B)**
    * *Component:* **Infrastructure**
    * Create `.github/workflows/deploy.yml`: Build binary -> SCP to VM -> Restart Service.
* **Task 3.3: Systemd Service (Joint)**
    * *Component:* **Infrastructure**
    * Write `mini-heroku.service` for auto-restart on boot/crash.

**Deliverable:** A live server IP that updates automatically on `git push`.

---

### Week 4: The Reverse Proxy
**Goal:** Route incoming traffic (e.g., `app.domain.com`) to the correct container.

**Learning Resources**
* **Go Lib:** [httputil.ReverseProxy](https://pkg.go.dev/net/http/httputil#ReverseProxy)
* **DNS:** [Magic Domain (nip.io)](https://nip.io/)

**Mentee Tasks**
* **Task 4.1: Dynamic Routing (Mentee A)**
    * *Component:* **Reverse Proxy**
    * Create HTTP Handler on Port 80. Inspect `Host` header (e.g., `blog.1.2.3.4.nip.io`).
* **Task 4.2: Proxy Logic (Mentee B)**
    * *Component:* **Reverse Proxy**
    * Implement logic: If Host == X, forward to Container IP Y.
* **Task 4.3: Integration (Joint)**
    * *Component:* **Controller & Proxy**
    * Hardcode a map (`app -> container_ip`) to verify routing works.

**Deliverable:** Accessing a container via a public URL.

---

### Week 5: State & Persistence
**Goal:** Ensure system remembers apps after a restart using a Database.

**Learning Resources**
* **DB:** [GORM with SQLite](https://gorm.io/docs/connecting_to_the_database.html#SQLite)

**🛠 Mentee Tasks**
* **Task 5.1: Schema Design (Mentee A)**
    * *Component:* **Database**
    * Design structs: `Project` (Name, Port, ContainerID, CreatedAt).
* **Task 5.2: DB Integration (Mentee B)**
    * *Component:* **Controller**
    * Initialize SQLite. Update `Deploy` function to save container details to DB.
* **Task 5.3: Reconciliation (Joint)**
    * *Component:* **Controller**
    * On startup: Query DB -> Check Docker -> Restart missing containers.

**Deliverable:** Platform survives server restart without data loss.

---

## Phase 3: The CLI & Polish (Weeks 6-8)

### Week 6: The CLI Tool
**Goal:** Build the `mini` command-line tool for users.

**Learning Resources**
* **Lib:** [Cobra Framework](https://github.com/spf13/cobra)
* **HTTP:** [Multipart Uploads](https://pkg.go.dev/mime/multipart)

**Mentee Tasks**
* **Task 6.1: CLI Skeleton (Mentee A)**
    * *Component:* **CLI**
    * Init Cobra. Create `mini version`, `mini deploy`.
* **Task 6.2: Upload Client (Mentee B)**
    * *Component:* **CLI**
    * Implement client logic: Zip folder -> POST to Server.
* **Task 6.3: Config (Joint)**
    * *Component:* **CLI**
    * Add `mini config set-host <url>` to point CLI to Cloud VM.

**Deliverable:** Deploying from laptop to Cloud VM via terminal.

---

### Week 7: Observability & Security
**Goal:** Allow users to see logs and secure the platform.

**Mentee Tasks**
* **Task 7.1: Log Streaming (Mentee A)**
    * *Component:* **Controller & CLI**
    * Implement `mini logs <app>`. Pipe Docker logs to HTTP Response -> CLI Stdout.
* **Task 7.2: Authentication (Mentee B)**
    * *Component:* **Controller & CLI**
    * Add API Key header check. Reject unauthorized requests.

**Deliverable:** `mini logs` works; Platform is secured.

---

### Week 8: Documentation & Demo
**Goal:** Final polish and presentation.

**Mentee Tasks**
* **Task 8.1: Cleanup (Joint)**
    * *Component:* **All**
    * Remove hardcoded paths, add comments, fix error handling.
* **Task 8.2: Docs (Mentee A)**
    * *Component:* **Docs**
    * Write `README.md` (Installation, Usage).
* **Task 8.3: Demo Prep (Mentee B)**
    * *Component:* **Demo**
    * Prepare script. Practice deploying a "TODO app" live.

**Deliverable:** Polished repo and successful live demo.

---
