# Page snapshot

```yaml
- generic [ref=e1]:
  - main [ref=e2]:
    - link "Skip to main content" [ref=e3] [cursor=pointer]:
      - /url: "#merchant-content"
    - generic [ref=e5]:
      - generic [ref=e6]:
        - heading "Test Business" [level=1] [ref=e8]
        - paragraph [ref=e9]: "Status: active"
      - button "Enrich merchant data from third-party vendors (Press E)" [ref=e11]:
        - img
        - text: Enrich Data
    - region "Merchant details" [ref=e12]:
      - generic [ref=e13]:
        - tablist [ref=e14]:
          - tab "Overview tab" [ref=e15]: Overview
          - tab "Business Analytics tab" [active] [selected] [ref=e16]: Business Analytics
          - tab "Risk Assessment tab" [ref=e17]: Risk Assessment
          - tab "Risk Indicators tab" [ref=e18]: Risk Indicators
        - tabpanel "Business Analytics tab" [ref=e19]
  - region "Notifications alt+T"
  - generic [ref=e25] [cursor=pointer]:
    - button "Open Next.js Dev Tools" [ref=e26]:
      - img [ref=e27]
    - generic [ref=e30]:
      - button "Open issues overlay" [ref=e31]:
        - generic [ref=e32]:
          - generic [ref=e33]: "7"
          - generic [ref=e34]: "8"
        - generic [ref=e35]:
          - text: Issue
          - generic [ref=e36]: s
      - button "Collapse issues badge" [ref=e37]:
        - img [ref=e38]
  - alert [ref=e40]
```