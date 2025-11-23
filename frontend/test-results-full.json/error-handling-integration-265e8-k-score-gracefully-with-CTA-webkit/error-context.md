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
          - generic [ref=e12]: Network request failed. Please check your connection.
          - generic [ref=e13]: "Error Code: UNKNOWN_ERROR"
  - generic [ref=e18] [cursor=pointer]:
    - button "Open Next.js Dev Tools" [ref=e19]:
      - img [ref=e20]
    - generic [ref=e25]:
      - button "Open issues overlay" [ref=e26]:
        - generic [ref=e27]:
          - generic [ref=e28]: "1"
          - generic [ref=e29]: "2"
        - generic [ref=e30]:
          - text: Issue
          - generic [ref=e31]: s
      - button "Collapse issues badge" [ref=e32]:
        - img [ref=e33]
  - alert [ref=e35]
```