# lucytech-url-analyzer

## Prerequisites

Before running the application, make sure you have installed:

- [Docker Desktop](https://docs.docker.com/get-docker/) (Ensure the Docker daemon is running)
- **macOS/Linux:** `make` utility (usually pre-installed)
- port **3000** and **8080** must be available. if not change **docker-compose.yaml** 

---

## 1. Quick Start (Using Makefile)

If `make` is installed on your system, use the provided shortcuts to manage the application lifecycle:

```bash
# Build images and start services in the background
cd lucytech-url-analyzer/
make run
```

### Service Endpoints

Once the application is running, you can access the services at:

- **Frontend:** [http://localhost:3000](http://localhost:3000)
- **Backend:** [http://localhost:8080](http://localhost:8080)

---

## 2. Makefile Commands

| Target | Executed Command | Description |
| :--- | :--- | :--- |
| `make run` | `docker compose up --build -d` | Builds, creates, and starts containers in detached mode, then prints access URLs. |
| `make stop` | `docker compose down` | Stops and removes running containers and associated networks. |
| `make build` | `docker compose build` | Builds or rebuilds service images defined in `docker-compose.yml`. |
| `make logs` | `docker compose logs -f` | Streams live logs from all running containers. |
| `make ps` | `docker compose ps` | Lists the status and port mappings of running containers. |

---

## 3. Alternative Execution Methods

### Running Docker Compose Manually

If `make` is unavailable, execute the Docker Compose commands directly in your terminal:

```bash
# Start containers in background
docker compose up --build -d

# View live output logs
docker compose logs -f

# Check container status
docker compose ps

# Stop containers
docker compose down
```