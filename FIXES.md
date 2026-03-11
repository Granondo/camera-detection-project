# Project Fixes & Improvements Plan

Issues found during full project audit (2026-03-11).
Check off items as they are completed.

---

## Critical

- [ ] **Redis healthcheck auth** — add `-a $REDIS_PASSWORD` to the healthcheck command in `docker-compose.yml:39` so dependent services (api-server, camera-detection) can actually start
- [ ] **CORS reflects arbitrary origins** — replace the wildcard origin reflection with an explicit allowlist in `internal/api/handlers.go:201`
- [ ] **HTTP response body leak in retry loop** — replace `defer resp.Body.Close()` inside the retry loop with explicit `resp.Body.Close()` calls in `internal/camera/ffmpeg_client.go:654`
- [ ] **ClickHouse materialized views wrong engine** — change `SummingMergeTree` to `AggregatingMergeTree` with `avgState`/`minState`/`maxState` combinators in `clickhouse/init/01_create_tables.sql`
- [ ] **Confidence threshold mismatch** — Python service reads `CONFIDENCE_THRESHOLD`, Go reads `DETECTION_CONFIDENCE_THRESHOLD`; align variable names across `docker-compose.yml`, `.env`, and `detection_service/main.py`

---

## High

- [ ] **RabbitMQ reconnection** — add connection/channel reconnect logic in `internal/queue/rabbitmq.go` so the consumer recovers from drops instead of silently dying
- [ ] **RabbitMQ infinite retry loop** — the worker checks `x-retry-count` header but never sets/increments it; implement proper retry counting so poison messages eventually reach the DLQ (`internal/queue/rabbitmq.go:282`)
- [ ] **Goroutine leak on shutdown** — pass a context or stop channel to the camera ping and storage cleanup tickers in `cmd/server/main.go:144` so they exit on SIGINT/SIGTERM
- [ ] **Path traversal on file serve** — sanitize `file_path` values before passing to `http.ServeFile` in `internal/api/handlers.go` (video stream, download, image endpoints)
- [ ] **YOLO model lazy init** — load model at startup (not first request) in `detection_service/main.py:103`, and reflect actual readiness in the `/health` response
- [ ] **Chakra UI v2/v3 mismatch** — remove `@chakra-ui/next-js` (v2, incompatible with Chakra v3) and `@emotion/styled` from `surveillance-next/package.json`
- [ ] **`getDurationEnv` appends "s" unconditionally** — fix so it handles values already containing a suffix (e.g. `30s`) in `internal/config/config.go:207`
- [ ] **PyTorch double-install** — remove `torch`/`torchvision` from `requirements.txt` since `detection_service/Dockerfile` already installs them from the CPU-only index
- [ ] **Camera passwords in plaintext** — encrypt or hash passwords in the `cameras` table (`migrations/01_init_schema.sql`)
- [ ] **Infra ports exposed to host** — remove host port bindings for Postgres, Redis, RabbitMQ, ClickHouse, Elasticsearch from `docker-compose.yml` (only expose API, UIs, and management tools)
- [ ] **`camera-detection` missing RabbitMQ env vars** — add `RABBITMQ_HOST`, `RABBITMQ_PORT`, `RABBITMQ_USER`, `RABBITMQ_PASSWORD`, `RABBITMQ_ENABLED` to the `camera-detection` service in `docker-compose.yml`

---

## Medium

- [ ] **No authentication** — add at minimum a static API key or JWT middleware to `internal/api/handlers.go`
- [ ] **Elasticsearch unsafe type assertions** — add safe type assertions (`val, ok := x.(type)`) in `internal/search/elasticsearch.go:256,291` to avoid panics on unexpected responses
- [ ] **`getImageUrl()` wrong path** — Next.js frontend builds `/images/{path}` but backend serves at `/api/image/{id}`; fix `surveillance-next/src/lib/api.ts`
- [ ] **Event type filter mismatch** — unify the event type lists between `EventsToolbar.tsx` and `search/page.tsx` to use the same values that the backend actually emits
- [ ] **Range header parser** — handle malformed, multi-range, and suffix ranges properly in `internal/api/handlers.go:1264`
- [ ] **Swagger URL hardcoded** — make the Swagger doc URL configurable (env var or derived from request host) in `cmd/api_service/main.go:132`
- [ ] **Detection worker path panic** — validate that `imagePath` has the expected prefix before slicing in `cmd/detection_worker/main.go:141`
- [ ] **Missing DB indexes** — add indexes on `frames.recording_id` and `frames.processed` in a new migration
- [ ] **`colorScheme` → `colorPalette`** — fix Chakra v3 prop rename in `surveillance-next/src/app/events/[id]/page.tsx` and `NotFound.tsx`
- [ ] **Race condition on `currentRecordingID`** — apply consistent mutex locking in `internal/camera/ffmpeg_client.go`
- [ ] **`handleNewRecording` race with FFmpeg** — increase or make configurable the sleep before `FinishRecording`, or use file-close detection instead of a timer (`internal/camera/ffmpeg_client.go:498`)

---

## Low

- [ ] **`GetAllFrames` missing LIMIT** — add a configurable limit to prevent OOM on large datasets (`internal/storage/repositories.go:308`)
- [ ] **CPU metric is fake** — replace goroutine-count-based fake with actual CPU sampling in `internal/analytics/metrics.go:67`
- [ ] **Disk usage path hardcoded** — use configured `OutputDir` instead of `/app/output` in `internal/analytics/metrics.go:94`
- [ ] **Per-request context for Redis/ClickHouse** — pass request context into Redis and ClickHouse calls so they can be cancelled/timed out
- [ ] **Graceful API shutdown** — replace `server.Close()` with `server.Shutdown(ctx)` in `cmd/api_service/main.go:199`
- [ ] **`strings.Title` deprecated** — replace with `golang.org/x/text/cases` (`internal/camera/ffmpeg_client.go:772`)
- [ ] **ES `Refresh: "true"` per insert** — change to `"false"` or `"wait_for"` in `internal/search/elasticsearch.go:201`
- [ ] **Error details leaked in API responses** — return generic messages to clients, log detailed errors server-side only
- [ ] **`createOutputDir` shells to `mkdir`** — replace with `os.MkdirAll` in `internal/camera/utils.go:204`
- [ ] **`console.log` in production** — remove debug log in `surveillance-next/src/app/events/[id]/page.tsx:66`
- [ ] **Prometheus app scrape targets** — uncomment/configure api-server and detection-service scrape jobs in `prometheus/prometheus.yml`
- [ ] **Pin Docker image versions** — replace `alpine:latest` and `clickhouse:latest` with specific versions in Dockerfiles and `docker-compose.yml`
- [ ] **Remove unused dependencies** — `framer-motion` (Next.js), `lucide-svelte`, `chart.js`, `Counter.svelte`, `RecordingCard.svelte` (Svelte)
- [ ] **Hardcoded camera IDs** — fetch camera list from API instead of `const CAMERA_IDS = [1, 2, 3, 4]` in `surveillance-next/src/app/frames/FramesToolbar.tsx`
- [ ] **Svelte UI production Dockerfile** — create a production build Dockerfile for `surveillance-ui/` (currently only has `Dockerfile.dev`)
- [ ] **ClickHouse hardcoded DB name** — use configured database name instead of hardcoded `"surveillance"` in `internal/analytics/migrations.go:79`

---

## Notes

- The `.env` file should be in `.gitignore` — verify it is not tracked. Rotate the camera credentials if they were ever committed.
- `02_add_frame_id_to_events.sql` uses `ON DELETE CASCADE` for `frame_id` (events deleted when frame deleted) but `ON DELETE SET NULL` for `camera_id` — confirm this asymmetry is intentional.
- Grafana has no ClickHouse or Elasticsearch datasources provisioned — add them when those stores are actively used.
- The Svelte `api.d.ts` OpenAPI types use no `/api/` prefix but actual calls do — types are currently unused for validation so it doesn't break, but should be fixed if typed fetch wrappers are added later.
