# Camera Detection Project (LLM Overview)

## What This Repository Is

This repo is a self-hosted video surveillance system for RTSP cameras (tested with Tapo), with object detection using YOLOv8.

It is built as a multi-service stack:

- Go services for camera capture, API, and background workers
- A Python (Flask) YOLO detection microservice (Ultralytics)
- PostgreSQL for primary metadata storage (recordings/frames/events/detections)
- RabbitMQ for asynchronous frame processing (optional but recommended)
- Redis for API caching
- ClickHouse for analytics/time-series queries (optional)
- Elasticsearch for search over events (optional)
- Web UIs: a Next.js viewer and a Svelte dashboard

The system stores video segments and extracted frames on disk under `output/`.

## Core Data Flow

1. **Camera capture (Go, FFmpeg)**
   - Connects to the RTSP stream.
   - Writes segmented MP4 recordings (e.g. 60s segments).
   - Extracts JPEG frames every N seconds (configurable).
2. **Frame processing**
   - If RabbitMQ is enabled: the camera service publishes frame jobs to `frames.to.detect`.
   - `detection-worker` consumes jobs and calls the YOLO detection service.
   - The worker updates frame status and (optionally) logs detections into ClickHouse.
3. **API serving**
   - `api-service` serves REST endpoints for recordings, frames, events, stats, search, and analytics.
   - Optionally uses Redis caching.
   - Can expose Swagger docs.
4. **UI**
   - Next.js app provides a public-style viewer for events/frames.
   - Svelte app provides a dashboard with status/stats.

## Main Services (by folder/entrypoint)

- Camera capture service (Go): `cmd/server/main.go`
  - Uses `internal/camera/ffmpeg_client.go`.
  - Writes files into `output/`.
  - Writes metadata into PostgreSQL.
  - Can publish frame jobs to RabbitMQ.
- API server (Go): `cmd/api_service/main.go`
  - Routes in `internal/api/handlers.go`.
  - Reads from PostgreSQL, optionally Redis cache.
  - Optional: ClickHouse analytics endpoints, Elasticsearch search endpoints.
- Detection worker (Go): `cmd/detection_worker/main.go`
  - Consumes RabbitMQ frame messages.
  - Calls the Python detection service.
  - Updates PostgreSQL (and ClickHouse if configured).
- YOLO detection service (Python): `detection_service/main.py`
  - Flask app with `/health` and `/detect`.
  - Uses Ultralytics YOLO (e.g. `yolov8n.pt`).

## Storage and Schemas

- PostgreSQL schema/migrations: `migrations/`
  - Tables include `cameras`, `recordings`, `frames`, `detections`, `events`, `system_stats`.
- ClickHouse schema: `clickhouse/init/01_create_tables.sql`
  - Stores detailed detections and system events/metrics with retention and materialized views.

## Frontends

- Next.js UI: `surveillance-next/`
  - Talks to the API at `/api/*` (server-side can use `API_URL`).
  - Pages include events and frames browsing.
- Svelte UI: `surveillance-ui/`
  - Dev-focused dashboard; includes charts and status cards.

## Configuration (Environment Variables)

Config is loaded from environment and optionally from `.env` (see `internal/config/config.go`).

Common camera settings:

- `RTSP_URL` (example: `rtsp://192.168.1.100:554/stream1`)
- `CAMERA_USERNAME`, `CAMERA_PASSWORD`
- `FRAME_RATE` (extract frame every N seconds)
- `SAVE_FRAMES` (`true/false`)
- `OUTPUT_DIR` (default `./output`)
- `FFMPEG_PATH` (default `ffmpeg`)
- `DETECTION_ENABLED` (`true/false`)

Stack settings:

- PostgreSQL: `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_NAME`
- Redis: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`
- RabbitMQ: `RABBITMQ_HOST`, `RABBITMQ_PORT`, `RABBITMQ_USER`, `RABBITMQ_PASSWORD`, `RABBITMQ_ENABLED`
- Detection service: `DETECTION_SERVICE_URL`, `DETECTION_SERVICE_TIMEOUT`, `DETECTION_CONFIDENCE_THRESHOLD`
- ClickHouse: `CLICKHOUSE_HOST`, `CLICKHOUSE_PORT`, `CLICKHOUSE_USER`, `CLICKHOUSE_PASSWORD`, `CLICKHOUSE_DATABASE`
- Elasticsearch: `ELASTICSEARCH_ENABLED`, `ELASTICSEARCH_URL`, `ELASTICSEARCH_INDEX`

## Running It

Recommended: Docker Compose.

- Full stack: `make start-full`
- Stop: `make stop-full`
- API health: `make api-status`

Local (non-Docker) development:

- Camera service: `make run-dev`
- API service: `make api-dev`

## Ports (Docker Compose defaults)

These are defined in `docker-compose.yml`:

- API server: `8080`
- Next.js UI: `3000`
- Svelte UI dev server: `5173`
- PostgreSQL: `5432`
- Redis: `6379`
- RabbitMQ: `5672` (AMQP), `15672` (UI)
- ClickHouse: `8123` (HTTP), `9000` (native), Tabix UI: `8083`
- Elasticsearch: `9200`, Kibana: `5601`
- Prometheus: `9090`, Grafana: `3001`

## Where Things Live On Disk

- Output media: `output/` (recordings `.mp4`, frames `.jpg`)
- Logs: `logs/`
- DB migrations: `migrations/`
- API Swagger/OpenAPI specs: `docs/`

## Go Module & Internal Packages

Go module: `camera-detection-project` (Go 1.24)

| Package | File(s) | Purpose |
|---------|---------|---------|
| `internal/config` | `config.go` | Loads env vars (and `.env`), defines `Config` struct |
| `internal/storage` | `database.go`, `models.go`, `repositories.go`, `service.go` | PostgreSQL connection, domain models, CRUD repositories |
| `internal/camera` | `ffmpeg_client.go`, `utils.go` | FFmpeg-based RTSP capture and frame extraction |
| `internal/api` | `handlers.go` | HTTP server with REST handlers, CORS middleware, Swagger |
| `internal/queue` | `rabbitmq.go`, `messages.go` | RabbitMQ producer/consumer, message types |
| `internal/cache` | `redis.go`, `service.go` | Redis cache client and caching service |
| `internal/analytics` | `clickhouse.go`, `migrations.go`, `metrics.go` | ClickHouse client, schema migrations, metrics ingestion |
| `internal/search` | `elasticsearch.go` | Elasticsearch indexing and querying |

Key dependencies: `lib/pq` (Postgres), `go-redis/v9`, `amqp091-go` (RabbitMQ), `clickhouse-go/v2`, `elastic/go-elasticsearch/v8`, `swaggo/swag` (Swagger).

## API Endpoints

All served by `api-service` on port `8080`:

- `GET /api/health` — health check
- `GET /api/status` — system status
- `GET /api/camera` — camera info
- `GET /api/recordings` — list recordings (paginated)
- `GET /api/recordings/{id}` — single recording
- `GET /api/video/{id}` — stream video
- `GET /api/download/{id}` — download video
- `GET /api/frames` — list frames (`?has_detection=true`, paginated)
- `GET /api/frames/{id}` — single frame
- `GET /api/image/{path}` — serve frame image
- `GET /api/events` — list events (paginated)
- `GET /api/events/{id}` — single event
- `GET /api/search` — search (Elasticsearch)
- `GET /api/stats` — aggregate stats
- `GET /api/stats/daily` — daily stats
- `POST /api/cache/clear` — clear Redis cache
- `GET /api/analytics/summary` — ClickHouse analytics summary
- `GET /api/analytics/detections/hourly` — hourly detection counts
- `GET /api/analytics/detections/count` — total detection count
- `GET /api/analytics/top-objects` — top detected object classes
- `GET /files/*` — static file server for `output/`

## Dockerfiles

- `Dockerfile` — camera capture service (Go)
- `Dockerfile.api` — API server (Go)
- `cmd/detection_worker/Dockerfile` — detection worker (Go)
- `detection_service/Dockerfile` — YOLO detection service (Python)
- `surveillance-next/Dockerfile` — Next.js UI
- `surveillance-ui/Dockerfile.dev` — Svelte UI (dev mode)

## Monitoring (Prometheus & Grafana)

- Prometheus config: `prometheus/prometheus.yml`
- Grafana datasources: `grafana/provisioning/datasources/datasource.yml`
- Grafana dashboards: `grafana/dashboards/infrastructure.json`
- Dashboard provisioning: `grafana/provisioning/dashboards/dashboard.yml`
- Prometheus UI: `http://localhost:9090`, Grafana UI: `http://localhost:3001`

## Testing & Code Quality

- `make test` — run all Go tests (`go test ./...`)
- `make check` — fmt + vet + test
- `make swagger` — regenerate Swagger docs (`swag init`)
- `make ts-types` — generate TypeScript types from OpenAPI spec (Svelte UI)
- Config has a unit test: `internal/config/config_test.go`

## Notes for LLMs / Contributors

- The camera capture path is FFmpeg-first (chosen for macOS compatibility and stability).
- The async detection pipeline is RabbitMQ + `detection-worker`; disabling RabbitMQ changes behavior toward synchronous/limited processing.
- ClickHouse and Elasticsearch are optional; the system is designed to run without them (reduced features).
- The Go module name is `camera-detection-project` — use this for all import paths (e.g. `camera-detection-project/internal/storage`).
- Swagger docs are generated from annotations in `cmd/api_service/main.go` and `internal/api/handlers.go`.
- RabbitMQ queue name for frame processing: `frames.to.detect` (with DLQ: `frames.to.detect.dlq`).
- Domain models are in `internal/storage/models.go` — Camera, Recording, Frame, Detection, Event, SystemStats.

