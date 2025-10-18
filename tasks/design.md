---
title: Feature design decisions
date: 2025-10-18
lastmod: 2025-10-18
status: active
---

This document captures key design decisions for htmltest features.

## Status Code Conventions

htmltest uses status codes in the refcache to represent the outcome of checking
external links. These codes include both real HTTP status codes and
tool-specific codes.

### Status Code Ranges

- **Positive (> 0)**: official HTTP status codes
- **Zero (= 0)**: Unchecked/undiscovered (neutral state)
- **Negative (< 0)**: Tool-specific error states, such as timeout, network
  error, DNS failure, etc.

### Rationale

1. **Clear distinction**: Three non-overlapping ranges for three distinct
   meanings
2. **Simple checks**: No bitwise operations needed, straightforward comparisons
3. **JSON readability**: Cache file remains human-readable (e.g., `-10` clearly
   indicates timeout)
4. **Future-proof**: -10 spacing provides 9 intermediate values per category
   (e.g., -11, -12, -13 for timeout variations)
5. **Industry alignment**: Negative codes for tool errors aligns with cURL and
   other tools
6. **Semantic clarity**: 0 as neutral/"not attempted" is intuitive
