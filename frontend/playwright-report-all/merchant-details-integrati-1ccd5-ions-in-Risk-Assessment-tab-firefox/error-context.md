# Page snapshot

```yaml
- generic [ref=e1]:
  - generic [ref=e2]:
    - generic [ref=e4]:
      - generic [ref=e5]:
        - heading "Test Business Inc" [level=1] [ref=e7]
        - paragraph [ref=e8]: "Technology • Status: active"
      - button "Enrich Data" [ref=e10]:
        - img
        - text: Enrich Data
    - generic [ref=e11]:
      - tablist [ref=e12]:
        - tab "Overview" [ref=e13]
        - tab "Business Analytics" [ref=e14]
        - tab "Risk Assessment" [active] [selected] [ref=e15]
        - tab "Risk Indicators" [ref=e16]
      - tabpanel "Risk Assessment" [ref=e17]:
        - generic [ref=e18]:
          - generic [ref=e21]:
            - img
            - text: Connecting
          - generic [ref=e23]:
            - img [ref=e24]
            - heading "No Risk Assessment" [level=2] [ref=e29]
            - generic [ref=e30]: No risk assessment has been performed for this merchant yet.
            - button "Start Assessment" [ref=e31]
  - region "Notifications alt+T"
  - alert [ref=e32]
```