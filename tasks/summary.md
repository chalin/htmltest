---
title: Summary of Changes on dev/main Branch
date: 2025-10-18
lastmod: 2025-10-18
status: active
---

# Summary of Changes on dev/main Branch

This document summarizes the changes made on the `dev/main` branch relative to
`master`.

## Features

So far all features are related to external link checking, and whether such
links are cached and/or retried on subsequent encounters.

### 1. RetryCachedErrors

This Introduced the `RetryCachedErrors` configuration option. When set to false,
This feature allows htmltest to reuse cached error status codes (e.g., 404)
without retrying them on subsequent runs.

- **Config parameter**: `RetryCachedErrors` (boolean, default: `true`)
- **Behavior**: When set to `false`, cached errors are reused instead of
  retrying
- **Backward compatibility**: Default `true` maintains existing behavior

### 2. Timeout Caching (#2)

**Status**: Completed

Extended caching behavior to save timeout errors (HTTP 408 status) to the
refcache when `RetryCachedErrors: false`.

- Timeouts are cached as status code 408 (`http.StatusRequestTimeout`)
- Prevents repeated timeout attempts on unreachable URLs
- Works in conjunction with `RetryCachedErrors` option

### 3. URL Encoding in Cache (#4)

**Status**: Completed

Fixed URL escaping in the JSON refcache file.

- Disabled HTML escaping when encoding URLs to JSON
- URLs are now stored in their unescaped form in `refcache.json`
- Improves readability and prevents double-escaping issues

## Infrastructure & Tooling

### Version ID Suffix

Added repository-specific suffix to version IDs for better build tracking.

### CI/CD Improvements

- Added `workflow_dispatch` trigger to GitHub Actions CI workflow
- Allows manual workflow runs from GitHub UI

### Project Structure

- Added `Makefile` for build automation
- Added `CONTRIBUTING.md` with development guidelines

## Planned Features

Most features are documented under the `tasks/` directory.

### CacheUncheckedExternal (In Progress)

See `tasks/cache-unchecked-external-links.md` for details. This feature will
allow caching external links without checking them when `CheckExternal: false`.
