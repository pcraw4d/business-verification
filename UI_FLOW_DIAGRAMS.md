# KYB Platform UI Flow Diagrams

## 1. Primary User Journey: Add Merchant Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                        USER JOURNEY                            │
└─────────────────────────────────────────────────────────────────┘

START: index.html (Landing Page)
  │
  │ [Auto-redirect after 3s OR User clicks "Enter"]
  ▼
┌─────────────────────┐
│ merchant-portfolio  │  ← Portfolio view of all merchants
│     .html          │
└──────────┬──────────┘
           │
           │ [User clicks "Add Merchant" button]
           ▼
┌─────────────────────┐
│   add-merchant.html │  ← Form to add new merchant
│                     │
│  [User fills form]  │
│  [Clicks "Verify    │
│   Merchant"]        │
└──────────┬──────────┘
           │
           │ ISSUE: Redirect failing here ⚠️
           │
           │ [processMerchantVerification()]
           │   1. Store data in sessionStorage
           │   2. Call APIs (Business Intelligence, Risk, Indicators)
           │   3. Store API results
           │   4. Redirect to merchant-details
           ▼
┌─────────────────────┐
│ merchant-details.html│  ← Should display merchant info
│                     │
│  [Load from         │
│   sessionStorage]   │
│  [Display data]     │
└─────────────────────┘
```

## 2. Navigation Structure

```
┌─────────────────────────────────────────────────────────────┐
│                    NAVIGATION HUB                            │
│              (dashboard-hub.html)                            │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│   Platform    │    │  Compliance   │    │   Merchant   │
│   Section     │    │   Section     │    │  Management  │
└───────┬───────┘    └───────┬───────┘    └───────┬───────┘
        │                    │                     │
        ▼                    ▼                     ▼
  ┌─────────┐          ┌─────────┐          ┌─────────┐
  │  Home   │          │ Status  │          │  Hub    │
  └─────────┘          └─────────┘          └─────────┘
  ┌─────────┐          ┌─────────┐          ┌─────────┐
  │  Hub    │          │  Gap    │          │Portfolio│
  └─────────┘          └─────────┘          └─────────┘
                              │                     │
                              ▼                     ▼
                        ┌─────────┐          ┌─────────┐
                        │Progress │          │ Details │
                        └─────────┘          │
                                                     ▼
                                              ┌───────────────┐
                                              │ merchant-    │
                                              │ details.html │
                                              └───────────────┘
```

## 3. Data Flow Patterns

### Pattern A: SessionStorage Flow (Current - Problematic)
```
┌──────────────┐
│ add-merchant  │
│   .html      │
└──────┬───────┘
       │
       │ [Form Submit]
       │
       ▼
┌──────────────┐
│ Store Data   │
│ sessionStorage│
│ .setItem()   │
└──────┬───────┘
       │
       │ ⚠️ Timing Issue
       │
       ▼
┌──────────────┐
│  Redirect    │
│ window.location│
│  .href      │
└──────┬───────┘
       │
       │ [Page Load]
       │
       ▼
┌──────────────┐
│merchant-     │
│details.html  │
└──────┬───────┘
       │
       │ [Read Data]
       │
       ▼
┌──────────────┐
│ sessionStorage│
│ .getItem()   │
└──────┬───────┘
       │
       │ ❌ May be empty
       │
       ▼
┌──────────────┐
│ Display Data │
│ (or error)   │
└──────────────┘
```

### Pattern B: URL Parameter Flow (Alternative)
```
┌──────────────┐
│ add-merchant  │
│   .html      │
└──────┬───────┘
       │
       │ [Form Submit]
       │
       ▼
┌──────────────┐
│ Create       │
│ Merchant ID  │
└──────┬───────┘
       │
       │ [Store in DB/API]
       │
       ▼
┌──────────────┐
│  Redirect    │
│ with ID      │
│ ?merchantId= │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│merchant-     │
│details.html │
│?merchantId=  │
└──────┬───────┘
       │
       │ [Extract ID]
       │
       ▼
┌──────────────┐
│ Fetch from   │
│ API/DB       │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Display Data │
└──────────────┘
```

### Pattern C: Hybrid Flow (Recommended)
```
┌──────────────┐
│ add-merchant  │
│   .html      │
└──────┬───────┘
       │
       │ [Form Submit]
       │
       ▼
┌──────────────┐
│ Store Data   │
│ (Multiple)   │
│ 1. sessionStorage│
│ 2. Create ID │
│ 3. Store API │
└──────┬───────┘
       │
       │ [Validate All]
       │
       ▼
┌──────────────┐
│  Redirect    │
│ with ID +    │
│ sessionStorage│
└──────┬───────┘
       │
       ▼
┌──────────────┐
│merchant-     │
│details.html  │
│?merchantId=  │
└──────┬───────┘
       │
       │ [Try Multiple Sources]
       │
       ▼
┌──────────────┐
│ 1. sessionStorage│
│ 2. URL param │
│ 3. API fetch │
└──────┬───────┘
       │
       │ [Use first available]
       │
       ▼
┌──────────────┐
│ Display Data │
└──────────────┘
```

## 4. Page Relationships Map

```
┌─────────────────────────────────────────────────────────────┐
│                    PAGE RELATIONSHIPS                        │
└─────────────────────────────────────────────────────────────┘

Entry Points:
  index.html ──┐
               ├──► dashboard-hub.html
  register.html┘

Merchant Flow:
  add-merchant.html ──► merchant-details.html
                            │
                            ├──► merchant-detail.html (alternative)
                            ├──► merchant-details-new.html (new version)
                            └──► merchant-details-old.html (legacy)

Portfolio Flow:
  merchant-portfolio.html ──► merchant-detail.html?merchant={id}
  merchant-hub.html ───────► merchant-hub-integration.html
  merchant-bulk-operations.html ──► merchant-comparison.html

Dashboard Flow:
  dashboard-hub.html ──┬──► dashboard.html (Business Intelligence)
                       ├──► risk-dashboard.html
                       ├──► enhanced-risk-indicators.html
                       ├──► compliance-dashboard.html
                       ├──► market-analysis-dashboard.html
                       └──► competitive-analysis-dashboard.html

Compliance Flow:
  compliance-dashboard.html ──► compliance-gap-analysis.html
                                    │
                                    └──► compliance-progress-tracking.html

Admin Flow:
  admin-dashboard.html ──┬──► admin-models.html
                         └──► admin-queue.html
```

## 5. Current Issues Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    ISSUE FLOW DIAGRAM                        │
└─────────────────────────────────────────────────────────────┘

ISSUE 1: Broken Redirect
  add-merchant.html
    │
    │ [Form Submit]
    │
    ▼
  sessionStorage.setItem() ──► ⚠️ May not complete
    │
    │ [Immediate redirect]
    │
    ▼
  window.location.href ───────► ⚠️ Redirects before data saved
    │
    ▼
  merchant-details.html
    │
    │ [Tries to read data]
    │
    ▼
  sessionStorage.getItem() ──► ❌ Returns null/empty
    │
    ▼
  Page shows error/empty ────► ❌ User sees broken page

ISSUE 2: Multiple Detail Pages
  User clicks "View Details"
    │
    ├──► merchant-detail.html (old)
    ├──► merchant-details.html (current)
    ├──► merchant-details-new.html (new)
    └──► merchant-details-old.html (legacy)
    
  ❌ Confusion: Which page is correct?
  ❌ Inconsistent behavior
  ❌ Maintenance nightmare

ISSUE 3: No Loading Feedback
  User submits form
    │
    │ [No visual feedback]
    │
    ▼
  Page appears frozen ───────► ⚠️ User doesn't know what's happening
    │
    │ [User clicks again]
    │
    ▼
  Duplicate submission ──────► ❌ Creates duplicate data
```

## 6. Recommended Solution Flow

```
┌─────────────────────────────────────────────────────────────┐
│                  RECOMMENDED SOLUTION                        │
└─────────────────────────────────────────────────────────────┘

STEP 1: Enhanced Data Storage
  add-merchant.html
    │
    │ [Form Submit]
    │
    ▼
  ┌─────────────────────┐
  │ 1. Validate Form    │
  │ 2. Show Loading     │
  │ 3. Store Data       │
  │    - sessionStorage │
  │    - Create ID      │
  │    - Store in API   │
  └──────────┬──────────┘
             │
             │ [Wait for confirmation]
             │
             ▼
  ┌─────────────────────┐
  │ Validate All Stores │
  │ ✓ sessionStorage    │
  │ ✓ API Response      │
  │ ✓ ID Generated      │
  └──────────┬──────────┘
             │
             │ [All validated]
             │
             ▼

STEP 2: Smart Redirect
  ┌─────────────────────┐
  │ Build Redirect URL  │
  │ /merchant-details?  │
  │ merchantId={id}     │
  └──────────┬──────────┘
             │
             │ [Redirect with ID]
             │
             ▼

STEP 3: Reliable Data Loading
  merchant-details.html?merchantId={id}
    │
    │ [Page Load]
    │
    ▼
  ┌─────────────────────┐
  │ Try Data Sources    │
  │ (in order)          │
  │ 1. sessionStorage  │
  │ 2. URL Parameter    │
  │ 3. API Fetch        │
  └──────────┬──────────┘
             │
             │ [Use first available]
             │
             ▼
  ┌─────────────────────┐
  │ Display Data        │
  │ ✓ Success            │
  └─────────────────────┘
```

## 7. Navigation State Machine

```
┌─────────────────────────────────────────────────────────────┐
│              NAVIGATION STATE MACHINE                        │
└─────────────────────────────────────────────────────────────┘

States:
  [HOME] ──► [DASHBOARD_HUB] ──► [ADD_MERCHANT]
    │              │                      │
    │              │                      │ [Submit]
    │              │                      ▼
    │              │              [PROCESSING]
    │              │                      │
    │              │                      │ [Success]
    │              │                      ▼
    │              │              [MERCHANT_DETAILS]
    │              │                      │
    │              │                      │ [View Portfolio]
    │              │                      ▼
    │              │              [MERCHANT_PORTFOLIO]
    │              │                      │
    │              │                      │ [Select Merchant]
    │              │                      ▼
    │              │              [MERCHANT_DETAIL]
    │              │                      │
    │              │                      │ [View Risk]
    │              │                      ▼
    │              │              [RISK_DASHBOARD]
    │              │                      │
    │              │                      │ [View Compliance]
    │              │                      ▼
    │              │              [COMPLIANCE_DASHBOARD]
    │              │                      │
    │              │                      │ [Back]
    │              └──────────────────────┘
    │
    └──► [REGISTER] ──► [LOGIN] ──► [DASHBOARD_HUB]
```

## 8. Data Persistence Strategy

```
┌─────────────────────────────────────────────────────────────┐
│            DATA PERSISTENCE STRATEGY                         │
└─────────────────────────────────────────────────────────────┘

Tier 1: SessionStorage (Fast, Temporary)
  ┌──────────────┐
  │ sessionStorage│
  │ - merchantData│
  │ - apiResults │
  │ - formState  │
  └──────────────┘
       │
       │ [Primary for current session]
       │

Tier 2: URL Parameters (Fallback)
  ┌──────────────┐
  │ URL Params   │
  │ ?merchantId= │
  │ ?view=       │
  └──────────────┘
       │
       │ [Secondary if sessionStorage fails]
       │

Tier 3: API/Database (Permanent)
  ┌──────────────┐
  │ Backend API  │
  │ /api/merchants│
  │ /api/details │
  └──────────────┘
       │
       │ [Tertiary - always available]
       │

Tier 4: LocalStorage (User Preferences)
  ┌──────────────┐
  │ localStorage │
  │ - preferences│
  │ - settings   │
  │ - cache      │
  └──────────────┘
```

---

## Summary

These diagrams illustrate:
1. **Current Flow**: How pages currently connect
2. **Data Flow**: How data moves between pages
3. **Issues**: Where problems occur
4. **Solutions**: Recommended improvements
5. **State Management**: Navigation states and transitions

Use these diagrams to:
- Understand current architecture
- Identify improvement opportunities
- Plan implementation
- Communicate with team
- Track progress

