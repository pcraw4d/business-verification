# Legacy Merchant Detail Pages

This directory contains the legacy merchant detail pages that were consolidated into a single `merchant-details.html` page.

## Archived Files

- `merchant-detail.html` - Original 5-tab version (Overview, Contact, Financial, Risk Assessment, Compliance)
- `merchant-details-new.html` - 4-tab version (Merchant Details, Business Analytics, Risk Assessment, Risk Indicators)
- `merchant-details-old.html` - Legacy 4-tab version
- `merchant-detail-static.html` - Static version from cmd/frontend-service/static
- `merchant-details-new-static.html` - Static version from cmd/frontend-service/static
- `merchant-details-old-static.html` - Static version from cmd/frontend-service/static

## Consolidation Date

2025-01-27

## Status

All features from these pages have been consolidated into `services/frontend/public/merchant-details.html` with 8 tabs:
1. Merchant Details
2. Business Analytics
3. Risk Assessment
4. Risk Indicators
5. Overview
6. Contact
7. Financial
8. Compliance

## Redirects

All old URLs are automatically redirected to `/merchant-details` via handlers in `cmd/frontend-service/main.go`:
- `/merchant-detail` → `/merchant-details`
- `/merchant-details-new` → `/merchant-details`
- `/merchant-details-old` → `/merchant-details`

Query parameters are preserved during redirect.

