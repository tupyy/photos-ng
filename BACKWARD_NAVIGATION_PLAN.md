# Backward Navigation Implementation Plan

## Goal
Implement backward cursor pagination to enable jumping to different years (e.g., 2000) and navigating forward/backward through time in the timeline view.

## Use Cases
1. **Jump to Year**: User clicks "2000" → load photos from 2000
2. **Navigate Forward**: From 2000 → move towards 2001, 2002, etc.
3. **Navigate Backward**: From 2020 → move towards 2019, 2018, etc.
4. **Bidirectional Scrolling**: Infinite scroll in both directions

## Implementation Steps

### Phase 1: Backend - SQL & Query Layer

#### 1.1 Update FilterByCursor Function
- **File**: `internal/datastore/pg/query_options.go`
- **Changes**:
  - Add `direction` parameter to `FilterByCursor(capturedAt, id, direction)`
  - Implement backward logic: `WHERE captured_at > cursor AND ORDER BY ASC`
  - Keep forward logic unchanged

```go
func FilterByCursor(capturedAt time.Time, id string, direction string) QueryOption {
    // if direction == "backward": use > operators + ASC order
    // if direction == "forward": use < operators + DESC order  
}
```

#### 1.2 Update MediaOptions
- **File**: `internal/services/options.go`
- **Changes**:
  - Add `Direction string` field to `MediaOptions`
  - Update `QueriesFn()` to handle direction-based sorting
  - Pass direction to `FilterByCursor`

#### 1.3 Update Query Logic
- **File**: `internal/services/options.go`
- **Changes**:
  - Conditional sorting based on direction:
    - Forward: `ORDER BY captured_at DESC, id DESC`
    - Backward: `ORDER BY captured_at ASC, id ASC`

#### 1.4 Result Reversal in Service
- **File**: `internal/services/media.go`
- **Changes**:
  - Add logic in `GetMedia()` to reverse results when `direction == "backward"`
  - Ensures final result maintains chronological order (newest first)

#### 1.5 Update HTTP Handler
- **File**: `internal/handlers/v1/http/media.go`
- **Changes**:
  - Parse `direction` query parameter 
  - Pass to `MediaOptions`

#### 1.6 Update API Spec
- **File**: `api/v1/http/openapi.yaml`
- **Changes**:
  - Add `direction` parameter to `/media` endpoint
  - Enum values: `["forward", "backward"]`
  - Default: `"forward"`

### Phase 2: Backend - Year Jump Functionality

#### 2.1 Create Date-Based Query Helper
- **File**: `internal/services/media.go`
- **Changes**:
  - Add `GetMediaFromYear(year int, limit int)` method
  - Constructs date range: `startDate = "${year}-01-01"`, `endDate = "${year}-12-31"`
  - Returns first page of photos from that year

#### 2.2 Update HTTP Handler for Year Jump
- **File**: `internal/handlers/v1/http/media.go`
- **Changes**:
  - Add logic to handle `year` parameter
  - If `year` provided: ignore cursor, fetch from year start
  - Generate proper cursor from results

### Phase 3: Frontend - API Layer

#### 3.1 Regenerate API Types
- Run `npm run generate:api` after API spec changes
- Verify `direction` parameter in generated types

#### 3.2 Update MediaFilters Interface
- **File**: `ui/src/shared/reducers/mediaSlice.ts`
- **Changes**:
  - Add `direction?: "forward" | "backward"` to `MediaFilters`
  - Add `year?: number` for year jumping

#### 3.3 Update fetchMedia Thunk
- **File**: `ui/src/shared/reducers/mediaSlice.ts`
- **Changes**:
  - Pass `direction` and `year` parameters to API call
  - Handle backward vs forward response logic

### Phase 4: Frontend - State Management

#### 4.1 Update MediaState
- **File**: `ui/src/shared/reducers/mediaSlice.ts`
- **Changes**:
  - Add `prevCursor?: string | null` for backward navigation
  - Add `hasPrevious: boolean` for UI state
  - Add `currentYear?: number` to track active year

#### 4.2 Update Cursor Computation
- **File**: `ui/src/shared/reducers/mediaSlice.ts`
- **Changes**:
  - Compute both `nextCursor` (from last item) and `prevCursor` (from first item)
  - Update logic in `fetchMedia.fulfilled` reducer

#### 4.3 Update Response Handling
- **File**: `ui/src/shared/reducers/mediaSlice.ts`
- **Changes**:
  - Forward: append to end of media array
  - Backward: prepend to beginning of media array
  - Year jump: replace entire media array

#### 4.4 Add Navigation Actions
- **File**: `ui/src/shared/reducers/mediaSlice.ts`
- **Changes**:
  - Add `loadPreviousPage` action
  - Add `jumpToYear` action
  - Update `loadNextPage` to be explicit about direction

### Phase 5: Frontend - Timeline Component

#### 5.1 Update Timeline State
- **File**: `ui/src/pages/timeline/index.tsx`
- **Changes**:
  - Access `prevCursor`, `hasPrevious` from state
  - Add `loadPrevious` callback

#### 5.2 Implement Year Navigation
- **File**: `ui/src/pages/timeline/index.tsx`
- **Changes**:
  - Update `handleYearSelect` to use `jumpToYear` action
  - Remove date filter workarounds, use proper year jumping

#### 5.3 Add Bidirectional Scroll
- **File**: `ui/src/pages/timeline/components/TimelineMediaGallery.tsx`
- **Changes**:
  - Add "Load Previous" sentinel at top of list
  - Add intersection observer for top sentinel
  - Call `loadPreviousPage` when top sentinel enters viewport

#### 5.4 Update Year Navigation Component
- **File**: `ui/src/pages/timeline/components/YearNavigation.tsx`
- **Changes**:
  - Highlight currently loaded year range
  - Handle year selection with jump functionality

### Phase 6: UI/UX Improvements

#### 6.1 Loading States
- Add separate loading indicators for:
  - Loading next page (bottom)
  - Loading previous page (top)
  - Jumping to year (full overlay)

#### 6.2 Visual Feedback
- Show current year range in timeline header
- Animate transitions when jumping years
- Add "scroll to top" button when far down timeline

#### 6.3 Performance Optimizations
- Implement virtual scrolling for very large timelines
- Limit total items in memory (remove old items when limit reached)
- Preload adjacent years when nearing boundaries

## Testing Strategy

### 6.4 Backend Testing
- Unit tests for backward cursor SQL generation
- Integration tests for year jumping
- Performance tests with large datasets

### 6.5 Frontend Testing  
- Test bidirectional scrolling
- Test year jumping maintains proper state
- Test edge cases (empty years, single items)

### 6.6 E2E Testing
- Test complete user journey: 2025 → 2000 → forward navigation
- Test performance with 10k+ photos across multiple years
- Test browser back/forward button compatibility

## Database Considerations

### 6.7 Index Optimization
- Ensure `idx_media_cursor_pagination (captured_at DESC, id DESC)` supports both directions
- Consider adding ASC index if needed: `idx_media_cursor_pagination_asc (captured_at ASC, id ASC)`

### 6.8 Query Performance
- Monitor query performance for backward navigation
- Optimize for year boundary queries

## API Compatibility

### 6.9 Backward Compatibility
- Default `direction=forward` maintains existing behavior
- Existing clients continue working unchanged
- Add API versioning if needed

## Deployment Strategy

### 6.10 Rollout Plan
1. Deploy backend changes (backward compatible)
2. Deploy frontend with feature flag
3. Enable bidirectional navigation gradually
4. Monitor performance and user feedback

## Success Criteria

- ✅ Can jump to any year (2000-2025) in <2 seconds
- ✅ Smooth bidirectional scrolling with no duplicate items
- ✅ Consistent performance regardless of timeline position
- ✅ Maintains chronological order in all navigation modes
- ✅ UI clearly indicates current year and navigation state