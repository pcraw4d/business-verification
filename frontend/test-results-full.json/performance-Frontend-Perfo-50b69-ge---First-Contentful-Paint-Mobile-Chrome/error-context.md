# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - main [ref=e2]:
    - link "Skip to main content" [ref=e3] [cursor=pointer]:
      - /url: "#merchant-content"
    - generic [ref=e5]:
      - generic [ref=e6]:
        - heading "Test Business" [level=1] [ref=e8]
        - paragraph [ref=e9]: "Technology • Status: active"
      - button "Enrich merchant data from third-party vendors (Press E)" [ref=e11]:
        - img
        - text: Enrich Data
    - region "Merchant details" [ref=e12]:
      - generic [ref=e13]:
        - tablist [ref=e14]:
          - tab "Overview tab" [selected] [ref=e15]: Overview
          - tab "Business Analytics tab" [ref=e16]: Business Analytics
          - tab "Risk Assessment tab" [ref=e17]: Risk Assessment
          - tab "Risk Indicators tab" [ref=e18]: Risk Indicators
        - tabpanel "Overview tab" [ref=e19]
  - region "Notifications alt+T"
  - button "Open Next.js Dev Tools" [ref=e26] [cursor=pointer]:
    - img [ref=e27]
  - alert [ref=e30]
```