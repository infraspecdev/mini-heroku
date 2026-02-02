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
        *   **Task 1.2: Architecture Design.** Create a detailed sequence diagram (using Mermaid) showing the request flow from `mini deploy` to a live URL.
        *   **Task 1.3: Design Pitch.** Present the proposed approach to the mentor. Discuss how the CLI will communicate with the Server (REST vs. others).
        *   **Task 1.4: SDK Basics.** Initialize Go module (`go mod init`) and write a script to connect to the Docker client and print the Docker version.
        *   **Task 1.5: Container Lifecycle.** Write functions to list running containers and pull/run an `alpine` container via the SDK.

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
