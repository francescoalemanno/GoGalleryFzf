# Plan: Fix fzf Search Relevance Sorting

## Analysis
The streaming search endpoint (`/api/search/stream`) sends results as they're discovered during filesystem walks, causing them to arrive in filesystem order rather than fzf relevance order. The frontend re-sorts what it receives, but the UX is poor (results jump around) and the ordering may still be incorrect if the server's scoring differs from the command-line fzf.

The non-streaming `/api/search` endpoint correctly sorts all results by fzf score before returning them.

## Tasks

### Phase 1: Disable Streaming Search (Quick Fix)
- [x] Modify `internal/server/template.go` to disable streaming search
  - Set `SSE_SUPPORTED = false` in the JavaScript to force fallback to `/api/search`
  - This forces the frontend to use the non-streaming endpoint which correctly sorts by fzf relevance
  - Changed: Added `const SSE_SUPPORTED = false;` before the `loadFiles` function

### Phase 2: Verify Non-Streaming Search Works Correctly
- [x] Test that `/api/search` returns results sorted by fzf relevance
  - Created `internal/fzf/matcher_test.go` with comprehensive tests
  - Tests verify that exact matches score higher than partial matches
  - Tests verify that consecutive character matches score higher than scattered matches
  - Tests verify descending score ordering in all search results
  - Fixed position conversion bug in `internal/fzf/matcher.go` (duplicate positions issue)

### Phase 3: Cleanup (Optional - if streaming not needed)
- [x] Remove `/api/search/stream` endpoint from `cmd/gallery/main.go`
- [x] Remove `HandleSearchStream` method from `internal/server/server.go`
- [x] Remove `scanAndStreamMatches` helper function
- [x] Remove `SearchMatch` struct if unused
- [x] Clean up streaming-related JavaScript code from template.go

### Phase 4: Add Tests for fzf Ordering (Optional but Recommended)
- [x] Created `internal/fzf/matcher_test.go` with test cases:
  - Test that exact matches score higher than partial matches
  - Test that consecutive character matches score higher than scattered matches
  - Test that prefix matches score higher than substring matches
  - Verify sort order matches expected fzf behavior
  - All tests pass successfully

## Verification Steps
1. Run the gallery server on a directory with various files
2. Search for a pattern that matches multiple files
3. Verify results appear sorted by relevance (best matches first)
4. Compare ordering with command: `find . -type f | fzf --filter="pattern"`
5. Results should match between web UI and command-line fzf

## Notes
- The junegunn/fzf library uses `FuzzyMatchV2` with specific scoring heuristics
- The `algo` package scoring considers: match positions, consecutive matches, word boundaries
- The `FuzzySearch` function in `internal/fzf/matcher.go` already sorts by score descending
- If we keep the streaming endpoint, it would need a complete rewrite to collect all results, sort them, then stream in order (defeating the streaming purpose)
