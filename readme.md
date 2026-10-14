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

## ⚙️ Features
* **Link Extraction:** Scrapes web pages to find all embedded URLs.
* **Link Categorization:** Differentiates between internal and external links (based on strict technical Origin standards, treating subdomains as external).
* **Accessibility Validation:** Pings discovered links to determine if they are alive or broken.

## 🧗‍♂️ Challenges & Approaches

During the development of the URL Analyzer, several technical constraints and edge cases were encountered. Here is how they were addressed:

* **Bypassing Modern Firewalls & Bot Detection**
  * *Challenge:* Simply spoofing browser headers is no longer sufficient to mimic legitimate, non-scripted HTTP calls. Modern firewalls use TLS layer cryptography and JA3 fingerprinting to detect and block scripts.
  * *Approach:* Explored and experimented with custom HTTP clients (such as [enetx/surf](https://github.com/enetx/surf)) to mimic browser behavior. While it offered some improvements, bypassing advanced TLS-layer verification remains an ongoing challenge for standard HTTP clients.
* **Defining Internal vs. External Links**
  * *Challenge:* The exact definition of an "internal" link can be ambiguous when dealing with complex domain architectures.
  * *Approach:* Adopted strict technical web standards (Origin rules). Subdomains are currently treated as external links, as browsers view them as completely different origins.
* **Handling Rate Limits & False Positives (HTTP 429)**
  * *Challenge:* Servers returning `429 Too Many Requests` were causing valid URLs to be marked as inaccessible. 
  * *Approach:* Tried with a small request jitter (delay) per hostname to reduce the likelihood of triggering rate limits. However, had not much success.
* **Handling StatusMethodNotAllowed (HTTP 405)**
     * *Approach:* this will be considerd as a accessible link as
       origin server recognizes the requested URL.

* **Network Anomalies**
  * *Challenge:* Transient client-side network issues (e.g., `ConnectionRefused`, DNS lookup failures) resulted in links being flagged as inaccessible, even if the target server was perfectly healthy.
 
## 🚀 Possible Improvements

To scale the application and increase the accuracy of link validation, the following architectural and code-level improvements are proposed for future iterations:

* **Headless Browser Integration:** Shift from standard HTTP fetching to a browser-based WebDriver architecture (e.g., Playwright or Puppeteer). This will solve current limitations with accurately scraping dynamic single-page applications (SPAs) and websites that rely on inline/infinite scrolling.however this approtch may need a different async architectural pattern.

* **Advanced Error Handling & Retry Mechanisms:** 
  * Differentiate actual broken links (e.g., `404 Not Found`) from temporary network failures (DNS drops, timeouts). 
  * Implement an exponential backoff and retry system for transient network errors.[`go-retryablehttp`](https://github.com/hashicorp/go-retryablehttp)

## 📚 Documentation & Testing

* **API Documentation:** Full API specifications and documentation can be found in the [`cmd/http/docs`](https://github.com/istore221/lucytech-url-analyzer/tree/main/url-analyzer-be/cmd/http/docs) directory.
* **Test Coverage:** Comprehensive test coverage reports are generated and available in [`coverage.html`](https://github.com/istore221/lucytech-url-analyzer/blob/main/url-analyzer-be/coverage.html).