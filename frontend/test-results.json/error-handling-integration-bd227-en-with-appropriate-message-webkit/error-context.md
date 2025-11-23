# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - main [ref=e2]:
    - link "Skip to main content" [ref=e3]:
      - /url: "#merchant-content"
    - generic [ref=e5]:
      - generic [ref=e6]:
        - heading "Test Business" [level=1] [ref=e8]
        - paragraph [ref=e10]: "Technology • Status: active"
      - button "Enrich merchant data from third-party vendors (Press E)" [ref=e12]:
        - img
        - text: Enrich Data
    - region "Merchant details" [ref=e13]:
      - generic [ref=e14]:
        - tablist [ref=e15]:
          - tab "Overview tab" [selected] [ref=e16]: Overview
          - tab "Business Analytics tab" [ref=e17]: Business Analytics
          - tab "Risk Assessment tab" [ref=e18]: Risk Assessment
          - tab "Risk Indicators tab" [ref=e19]: Risk Indicators
        - tabpanel "Overview tab" [ref=e20]
  - region "Notifications alt+T"
  - button "Open Next.js Dev Tools" [ref=e27] [cursor=pointer]:
    - img [ref=e28]
  - alert [ref=e33]
```