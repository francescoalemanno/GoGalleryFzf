# Goal: Fix fzf Search Relevance Sorting

## Problem
The junegunn/fzf library is not being used correctly for search result ranking. Search results in the browser are not sorted by relevance as they appear when using fzf manually from the command line.

## Current Behavior
- The streaming search endpoint (`/api/search/stream`) sends results as they're discovered during filesystem walks
- High-score matches are streamed immediately in filesystem order (not fzf relevance order)
- Low-score matches are batched with only batch-level sorting
- Frontend re-sorts received results, but they arrive in incorrect order

## Expected Behavior
- Search results should be ranked by fzf relevance score (highest first)
- Results should appear in the same order as running `fzf` manually on the file list

## Scope
- Fix search result ordering in the gallery web interface
- Ensure both `/api/search` and `/api/search/stream` return results sorted by fzf relevance
- Maintain existing pagination and streaming behavior where possible

## Constraints
- Keep using the junegunn/fzf library (v0.70.0)
- Maintain backward compatibility with existing API responses
- Don't break existing thumbnail, rotate, or rename functionality
