# Task 4: Business Classification Engine — Completion Summary

## Executive Summary

Task 4 delivers a production-ready classification engine that maps businesses to industry codes accurately and quickly. It combines multiple signals (keywords, type, industry text, name patterns, fuzzy matching), enriches results with crosswalks, and exposes simple APIs with strong observability and performance.

- What we did: Built a hybrid classification pipeline, added NAICS↔MCC/SIC mapping, confidence scoring, history, and health endpoints; optimized with caching, batching, database indexing, and HTTP pooling.
- Why it matters: Higher accuracy and faster response times enable better onboarding decisions and user experiences.
- Success metrics: Accuracy target (>95% on tests), latency target (<500ms single request), batch scaling (1000+), and healthy data source checks.

## How to Validate Success (Checklist)

- Single classify: POST /v1/classify returns a code, name, and confidence quickly (<500ms local).
- Batch classify: POST /v1/classify/batch processes hundreds+ entries with stable throughput.
- Confidence report: POST /v1/classify/confidence-report returns meaningful summary.
- History: GET /v1/classify/history/{business_id} returns prior results when available.
- Datasource health: GET /v1/datasources/health shows green; errors are surfaced with latency.
- Performance: Observe improved latency after repeated requests (cache hits); metrics include classification durations.
- Mapping: NAICS↔MCC/SIC fields present when applicable.

## PM Briefing

- Elevator pitch: Accurate, fast business classification with simple APIs and clear confidence.
- Business impact: Better onboarding decisions, fewer manual reviews, and improved reporting.
- KPIs to watch: Classification accuracy, P95 latency, cache hit rate, error rate for external sources.
- Stakeholder impact: Sales/Onboarding get faster, more accurate classifications; Analytics gains consistent codes.
- Rollout: Safe to expose to early adopters; publish request/response examples and confidence guidance.
- Risks & mitigations: Ambiguous names—mitigated by fuzzy matching and history; external source issues—mitigated by health checks and caching.
- Known limitations: Confidence weights are conservative; can be tuned with real-world feedback.
- Next decisions for PM: Approve accuracy thresholds for GA; prioritize new data sources or verticals for enrichment.
- Demo script: Single and batch classify, inspect confidence, view health, and show metrics panel.

## Overview

Task 4 implemented an end-to-end, production-ready business classification engine with clean architecture, observability, and performance optimizations. It includes:

- Hybrid classification pipeline (keywords, business type, industry, name patterns, fuzzy matching)
- NAICS↔MCC/SIC mapping and crosswalk enrichment
- Confidence scoring with method weighting and agreement boosts
- Classification history storage and retrieval
- API endpoints for single, batch, history, and confidence reporting
- Result caching, request batching, data source abstraction, validation/cleaning
- Health checks for enrichment sources, metrics, and slow-path alerting

## Primary Files & Responsibilities

- `internal/classification/service.go`: Core pipeline, cache, batch, history, scoring
- `internal/classification/normalize.go`: Text normalization/tokenization
- `internal/classification/fuzzy.go`: Levenshtein similarity and helpers
- `internal/classification/mapping.go`: Industry text→NAICS, NAICS→MCC/SIC crosswalk
- `internal/classification/data_loader.go`: Industry datasets + search (keyword/fuzzy)
- `internal/datasource/`: Data source abstraction and DB-backed enricher
- `cmd/api/main.go`: Routes and handlers for classify endpoints and datasource health
- `internal/observability/metrics.go`: Metrics including classification duration
- `internal/config/config.go`: New cache + HTTP pooling + performance thresholds
- `internal/database/migrations/003_performance_indexes.sql`: Trigram and composite indexes

## Endpoints

- POST `/v1/classify`: Single classification
- POST `/v1/classify/batch`: Batch classification with in-batch dedup + bounded concurrency
- GET `/v1/classify/history/{business_id}`: Paginated history
- POST `/v1/classify/confidence-report`: Summarizes confidence across results
- GET `/v1/datasources/health`: Health summary for enrichment sources

## Classification Pipeline (high level)

1) Validate input → optional enrichment (DB source) → sanitize
2) Methods (aggregated):
   - keyword_based(_naics|_mcc|_sic)
   - business_type_based
   - industry_based + industry_text_mapping
   - name_pattern_based
   - fuzzy_* across NAICS/MCC/SIC
   - history_* DB fallback
3) Crosswalk enrichment from primary NAICS → MCC/SIC
4) Confidence post-processing: method weights, agreement boosts, dedup by code
5) Persist primary result (if DB available) → Cache result

## Observability & Performance

- Metrics: `http_requests_*`, `classification_duration_seconds`, DB + external call metrics
- Slow-path logs: configurable thresholds via `SLOW_REQUEST_THRESHOLD`, `SLOW_CLASSIFICATION_THRESHOLD`
- Health: `/v1/datasources/health` probes sources and returns latency/health
- Performance indexes (migration 003): trigram GIN on business text fields, composite on classifications
- Batch efficiency: in-batch dedup + worker pool (default 8 workers)
- Caching: TTL-based in-memory cache with janitor and cap

## Configuration (env)

- Classification cache: `CLASSIFICATION_CACHE_ENABLED`, `CLASSIFICATION_CACHE_TTL`, `CLASSIFICATION_CACHE_MAX_ENTRIES`, `CLASSIFICATION_CACHE_JANITOR_INTERVAL`
- External HTTP pooling: `EXT_HTTP_MAX_IDLE_CONNS`, `EXT_HTTP_MAX_IDLE_CONNS_PER_HOST`, `EXT_HTTP_IDLE_CONN_TIMEOUT`, `EXT_HTTP_TLS_HANDSHAKE_TIMEOUT`, `EXT_HTTP_EXPECT_CONTINUE_TIMEOUT`, `EXT_HTTP_REQUEST_TIMEOUT`
- Observability thresholds: `SLOW_REQUEST_THRESHOLD`, `SLOW_CLASSIFICATION_THRESHOLD`

## Running & Testing

- Run API: `go run cmd/api/main.go`
- Unit tests: `go test ./...`
- Quick curls:
  - Classify single:

    ```sh
    curl -s localhost:8080/v1/classify -H 'Content-Type: application/json' \
      -d '{"business_name":"Acme Software Solutions","business_type":"technology"}'
    ```

  - Batch classify:

    ```sh
    curl -s localhost:8080/v1/classify/batch -H 'Content-Type: application/json' \
      -d '{"businesses":[{"business_name":"XYZ Corp"},{"business_name":"Law Office Associates","industry":"legal"}]}'
    ```

  - Datasource health:

    ```sh
    curl -s localhost:8080/v1/datasources/health | jq
    ```

## Developer Guide: Extending Classification

- Add a method: implement `classifyByXxx` in `service.go`, return `[]IndustryClassification`, add method weight in `methodWeightFor`.
- Tune fuzzy thresholds: adjust constants in `classifyByFuzzy` or expose via config.
- Add a data source: implement `DataSource` in `internal/datasource`, register it in `initEnrichment`, optionally use pooled HTTP client via aggregator.
- Update mappings/crosswalk: edit `mapping.go`; consider adding tests.

## Known Notes

- Default primary ID is time-based; switch to UUIDs when integrating ID provider.
- Confidence scoring is conservative by design; revisit weights after empirical evaluation.

## Acceptance

- All Task 4 subtasks (4.1–4.5) completed and tested.

## Non-Technical Summary of Completed Subtasks

### 4.1 Design Classification Data Models

- What we did: Defined clear data structures for businesses and industry codes so the system knows how to store and talk about a business and its classification.
- Why it matters: A consistent structure avoids confusion and enables accurate, repeatable results and reporting.
- Success metrics: Schema stability (no breaking changes), data correctness in tests, and ability to export results without manual cleanup.

### 4.2 Implement Core Classification Logic

- What we did: Built a “hybrid” engine that looks at a business name, type, description, and keywords, and also uses fuzzy matching to find the best industry codes.
- Why it matters: Using multiple signals improves accuracy and reduces “unclassified” outcomes.
- Success metrics: Target accuracy >95% on test datasets; average response time under 500ms for single requests; low rate of default classifications.

### 4.3 Build Classification API Endpoints

- What we did: Added simple, secure endpoints to classify single businesses, run batch jobs, fetch history, and summarize confidence.
- Why it matters: Clear APIs make integration easy for partners and internal tools.
- Success metrics: Stable 2xx/4xx/5xx distribution; predictable errors for invalid requests; batch throughput scales to 1,000+ items.

### 4.4 Integrate External Data Sources

- What we did: Created a plug-in layer to pull helpful business info (e.g., cleaned names, industry hints) from our database and future providers; added validation and health checks.
- Why it matters: Better input data improves results; health checks ensure reliability.
- Success metrics: Enrichment success rate; health endpoint shows green status; reduced manual data fixes.

### 4.5 Performance Optimization

- What we did: Added caching, batching, database indexes, and pooled HTTP connections; instrumented performance metrics and slow-path alerts.
- Why it matters: Faster responses, lower costs, and more reliable service under load.
- Success metrics: P95 latency under 500ms for single classify; batch completes within expected SLO; cache hit rate improves with repeated traffic; alerts trigger only on true slow paths.
