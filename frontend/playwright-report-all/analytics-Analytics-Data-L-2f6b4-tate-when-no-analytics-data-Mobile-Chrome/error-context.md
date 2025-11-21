# Page snapshot

```yaml
- generic [ref=e1]:
  - generic [ref=e2]:
    - generic [ref=e4]:
      - generic [ref=e5]:
        - heading "Test Business" [level=1] [ref=e7]
        - paragraph [ref=e8]: "Status: active"
      - button "Enrich Data" [ref=e10]:
        - img
        - text: Enrich Data
    - generic [ref=e11]:
      - tablist [ref=e12]:
        - tab "Overview" [ref=e13]
        - tab "Business Analytics" [active] [selected] [ref=e14]
        - tab "Risk Assessment" [ref=e15]
        - tab "Risk Indicators" [ref=e16]
      - tabpanel "Business Analytics" [ref=e17]:
        - generic [ref=e19]:
          - generic [ref=e21]:
            - generic [ref=e22]:
              - generic [ref=e23]: Website Analysis
              - generic [ref=e24]: Website performance and security
            - generic [ref=e25]: From Website Analysis API
          - generic [ref=e26]:
            - generic [ref=e27]:
              - paragraph [ref=e28]: Website URL
              - paragraph [ref=e29]: N/A
            - generic [ref=e30]:
              - paragraph [ref=e31]: Performance Score
              - paragraph [ref=e32]: N/A
            - generic [ref=e33]:
              - paragraph [ref=e34]: Accessibility Score
              - paragraph [ref=e35]: N/A
  - region "Notifications alt+T"
  - alert [ref=e36]
```