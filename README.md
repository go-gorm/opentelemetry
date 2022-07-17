# opentelemetry

Feature
---
- Tracing 
  - support tracing gorm by Hook `Create` `Query` `Delete` `Update` `Row` `Raw` 
- Metrics 
  - Collect DB Status 
- Logrus
  - Use logrus replace gorm default logger
  - Use hook to report span message
- Provider
  - Out-of-the-box default opentelemetry provider
  - Support setting via environment variables
---

