# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - alert [ref=e3]:
    - generic [ref=e4]: Merchant not found
  - region "Notifications alt+T":
    - list:
      - listitem [ref=e5]:
        - img [ref=e7]
        - generic [ref=e11]:
          - generic [ref=e12]: CORS policy blocked the request. Please check server configuration.
          - generic [ref=e13]: "Error Code: CORS_ERROR"
  - generic [ref=e18] [cursor=pointer]:
    - button "Open Next.js Dev Tools" [ref=e19]:
      - img [ref=e20]
    - generic [ref=e23]:
      - button "Open issues overlay" [ref=e24]:
        - generic [ref=e25]:
          - generic [ref=e26]: "1"
          - generic [ref=e27]: "2"
        - generic [ref=e28]:
          - text: Issue
          - generic [ref=e29]: s
      - button "Collapse issues badge" [ref=e30]:
        - img [ref=e31]
  - alert [ref=e33]
```