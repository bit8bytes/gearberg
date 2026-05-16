---
name: Go interface naming convention
description: Single-method interfaces use -er suffix; multi-method interfaces use descriptive nouns
type: feedback
---

The `-er` suffix (e.g., `Renderer`, `Writer`, `Handler`) is only appropriate for single-method interfaces in Go, following the `io.Writer` / `http.Handler` convention.

**Why:** Go convention — the `-er` suffix implies a single capability. Multi-method interfaces describe a role or concept, not a single action.

**How to apply:** For interfaces with 2+ methods, use a descriptive noun instead (e.g., `View`, `Store`, `Service`, `Client`). Never use `-er` for multi-method interfaces.
