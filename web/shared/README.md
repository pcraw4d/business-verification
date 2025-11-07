# Shared Component Library

This directory contains reusable components, services, and utilities that eliminate code duplication across the KYB Platform.

## Structure

```
shared/
├── data-services/      # Unified data access layer
├── visualizations/    # Reusable chart and visualization components
├── components/         # Reusable UI components (export, alerts, etc.)
├── navigation/         # Cross-page/tab navigation utilities
├── events/            # Event system for component communication
├── types/             # TypeScript type definitions
└── utilities/         # Helper utilities (formatters, validators, etc.)
```

## Usage

### Data Services

#### Risk Data Service
```javascript
import { getRiskDataService } from '../shared/data-services/risk-data-service.js';

const riskService = getRiskDataService();
const riskData = await riskService.loadRiskData(merchantId, {
    includeHistory: true,
    includePredictions: true,
    includeBenchmarks: true
});
```

#### Merchant Data Service
```javascript
import { getMerchantDataService } from '../shared/data-services/merchant-data-service.js';

const merchantService = getMerchantDataService();
const merchantData = await merchantService.loadMerchantData(merchantId, {
    includeClassification: true,
    includeRisk: true
});
```

### Event Bus

```javascript
import { getEventBus } from '../shared/events/event-bus.js';

const eventBus = getEventBus();

// Subscribe to events
eventBus.on('risk-data-loaded', (data) => {
    console.log('Risk data loaded:', data);
});

// Emit events
eventBus.emit('risk-data-loaded', { merchantId, riskData });
```

## Architecture Principles

1. **Separation of Concerns**: Data services, visualizations, and UI components are separate
2. **Dependency Injection**: Components accept dependencies for testability
3. **Event-Driven**: Components communicate via events, not direct coupling
4. **Progressive Enhancement**: Components work standalone or together
5. **Type Safety**: TypeScript definitions for all components

## Module System

The shared library uses ES6 modules. For compatibility:

- Use `import`/`export` syntax
- Components check for global fallbacks (APIConfig, EventBus, etc.)
- Works with both module and global scope

## Dependencies

- **APIConfig**: Must be available globally or injected
- **EventBus**: Available via `getEventBus()` or injected
- **Chart.js**: For visualization components
- **D3.js**: For advanced visualizations (optional)

## Migration Guide

When migrating existing code to use shared components:

1. Replace direct API calls with shared data services
2. Replace inline chart creation with shared visualization components
3. Replace duplicate export logic with SharedExportService
4. Use EventBus for component communication instead of direct callbacks

## Testing

All shared components should be unit tested. See individual component files for test examples.

## Contributing

When adding new shared components:

1. Follow the existing structure and patterns
2. Add TypeScript definitions in `types/`
3. Document usage in component JSDoc
4. Add to this README
5. Write unit tests

