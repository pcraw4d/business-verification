### Task 4.1: Design Classification Data Models — Completion Summary

Status: COMPLETED

What’s implemented

- Business-facing request/response models in `internal/classification/service.go`:
  - `ClassificationRequest`, `ClassificationResponse`, `IndustryClassification`, batch request/response.
- Industry code data schema and loader in `internal/classification/data_loader.go`:
  - `IndustryCodeData` with NAICS/MCC/SIC maps; CSV loaders and keyword search helpers; getters for names/descriptions.
- Persistence schema in DB for results in `internal/database/models.go` and migration `001_initial_schema.sql`:
  - `BusinessClassification` with `industry_code`, `industry_name`, `confidence_score`, `classification_method`, `raw_data`.
- NAICS/MCC/SIC mapping usage:
  - Keyword search pipelines and code-to-name resolution when `industryData` is present.
- Business type categorization:
  - `classifyByBusinessType` mapping common types (e.g., llc, corporation) to NAICS with confidence.
- Confidence scoring models:
  - Per-method scores and `calculateOverallConfidence`; primary selection via `determinePrimaryClassification`.

Key references

- `internal/classification/service.go`: models, multi-method classification, confidence, primary selection, DB store.
- `internal/classification/data_loader.go`: NAICS/MCC/SIC loaders & search utilities.
- `internal/database/models.go` + `internal/database/migrations/001_initial_schema.sql`: storage schema.
- Tests: `internal/classification/service_test.go`, `internal/classification/data_loader_test.go` validate scoring and loaders.

Engineer guide

- Loading datasets: call `classification.LoadIndustryCodes(<data_dir>)`, then use `NewClassificationServiceWithData(...)` to inject into service.
- Classify a business: use `ClassifyBusiness(ctx, *ClassificationRequest)`; batch with `ClassifyBusinessesBatch`.
- Extend schemas: add fields to `IndustryClassification` if needed; align DB via a new migration and update `storeClassification` accordingly.
- Adjust scoring: tune per-method confidences and `determinePrimaryClassification`/`calculateOverallConfidence` logic.
- Add code systems: extend `IndustryCodeData` and provide loader + search helpers; wire into `classifyByKeywords` similarly to NAICS/MCC/SIC.
