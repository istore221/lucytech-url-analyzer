# lucytech-url-analyzer

A Go application set up with multi-stage Docker builds for minimal image size and automated task execution.

## Prerequisites

Before running the application, make sure you have installed:

* [Docker Desktop](https://docs.docker.com/get-docker/) (Ensure the Docker daemon is running)
* **macOS/Linux:** `make` utility (usually pre-installed)
* **Windows:** [Task](https://taskfile.dev/#/installation) (`winget install Task.Task` or `choco install task`)
* **Local Go (Optional):** Go v1.24+ (if running without Docker)

---

## 1. Running on macOS / Linux (using Makefile)

Using the included `Makefile`:

```bash
# Build and run application in Docker (auto-removes container on exit)
make run
```

### Available Make Targets

| Target | Description |
| :--- | :--- |
| `make run` | **(Default)** Builds the Docker image and executes the app inside a container. |
| `make build` | Builds the optimized multi-stage Docker image. |
| `make clean` | Deletes the local Docker image. |
| `make help` | Lists all available targets and descriptions. |

---

## 2. Running on Windows (using Taskfile.yml)

Using [Task](https://taskfile.dev):

```powershell
# Build and run application in Docker (auto-removes container on exit)
task run
```

### Available Task Commands

| Command | Description |
| :--- | :--- |
| `task run` | **(Default)** Builds and runs the application inside Docker. |
| `task build` | Builds the Docker image. |
| `task clean` | Removes the Docker image from local store. |

---

## 3. Running Manually (Docker & Native Go)

If you do not want to use `make` or `task`, you can run the application directly on any platform (macOS, Linux, or Windows PowerShell/Command Prompt).

### Option A: Using Docker Manually

1. **Build the Docker image:**
   ```bash
   docker build -t lucytech-url-analyzer -f Dockerfile .
   ```

2. **Run the container (auto-cleans on exit):**
   ```bash
   docker run --rm lucytech-url-analyzer
   ```

3. **Clean up image (optional):**
   ```bash
   docker rmi lucytech-url-analyzer
   ```

### Option B: Native Go Execution (Without Docker)

If you have Go ($1.24+$) installed on your host system:

```bash
go run ./example/main.go
```

---