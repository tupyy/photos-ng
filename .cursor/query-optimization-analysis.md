# Query Optimization Analysis for Photos NG

## Current State Analysis

The Photos NG codebase uses:
- **Squirrel query builder** for dynamic queries
- **PostgreSQL** as the main database
- **Basic pagination** with limit/offset
- **Simple indexing** strategy

## Query Optimization Areas

### 1. **Indexing Strategy**
```sql
-- Current: Likely basic indexes
-- Optimized: Strategic composite indexes for actual use cases

-- Media queries (core access patterns from Photos NG)
CREATE INDEX idx_media_album_captured ON media(album_id, captured_at DESC);
CREATE INDEX idx_media_type_date ON media(type, captured_at DESC);
CREATE INDEX idx_media_captured_desc ON media(captured_at DESC); -- Timeline view

-- Album queries
CREATE INDEX idx_album_path ON albums(path); -- File system sync
CREATE INDEX idx_album_parent ON albums(parent_id); -- Hierarchical albums

-- Partial indexes for filtering by media type
CREATE INDEX idx_media_photos ON media(captured_at DESC) WHERE type = 'photo';
CREATE INDEX idx_media_videos ON media(captured_at DESC) WHERE type = 'video';

-- Stats queries optimization
CREATE INDEX idx_media_year ON media(EXTRACT(YEAR FROM captured_at));
```

### 2. **Pagination Improvements**
```go
// Current: OFFSET-based pagination (slow for large datasets)
SELECT * FROM media ORDER BY captured_at LIMIT 20 OFFSET 1000;

// Problem: OFFSET becomes slower as offset increases
// For offset 10000, database still needs to scan and skip 10000 rows

// Optimized: Cursor-based pagination
type PaginationCursor struct {
    CapturedAt time.Time `json:"captured_at"`
    ID         string    `json:"id"` // tiebreaker for same timestamps
}

func GetMediaWithCursor(cursor *PaginationCursor, limit int) ([]Media, *PaginationCursor, error) {
    query := `
        SELECT id, filename, captured_at, thumbnail, album_id 
        FROM media 
        WHERE ($1::timestamp IS NULL OR captured_at < $1 OR (captured_at = $1 AND id < $2))
        ORDER BY captured_at DESC, id DESC 
        LIMIT $3`
    
    // Execute query...
    // Return next cursor from last item
}

// Benefits:
// - Consistent performance regardless of page depth
// - Real-time data changes don't affect pagination
// - Better for infinite scroll UIs
```

### 3. **N+1 Query Problems**
```go
// Current: Potential N+1 when loading album media
func (h *Handler) GetAlbumsWithMedia() ([]AlbumWithMedia, error) {
    albums, err := h.albumSrv.ListAlbums()
    if err != nil {
        return nil, err
    }
    
    for _, album := range albums {
        // This creates N additional queries!
        media, err := h.mediaSrv.GetMediaForAlbum(album.ID)
        if err != nil {
            return nil, err
        }
        album.Media = media
    }
    return albums, nil
}

// Optimized: Batch loading with single query
func (h *Handler) GetAlbumsWithMediaOptimized() ([]AlbumWithMedia, error) {
    // Single query with JOIN or batch loading
    query := `
        SELECT 
            a.id as album_id, a.name, a.path,
            m.id as media_id, m.filename, m.thumbnail
        FROM albums a
        LEFT JOIN media m ON a.id = m.album_id
        ORDER BY a.name, m.captured_at DESC`
    
    // Group results in application code
    // Or use separate batch query for media
}

// Alternative: Use data loader pattern
type MediaLoader struct {
    batch []string
    cache map[string][]Media
}

func (l *MediaLoader) LoadByAlbumID(albumID string) []Media {
    // Batches requests and executes single query
}
```

### 4. **Aggregation Optimization**
```sql
-- Current: Multiple queries for stats
SELECT COUNT(*) FROM media;
SELECT COUNT(*) FROM albums;
SELECT DISTINCT EXTRACT(YEAR FROM captured_at) FROM media ORDER BY 1 DESC;

-- Problem: Multiple round trips, multiple table scans

-- Optimized: Single query with CTEs and window functions
WITH media_stats AS (
  SELECT 
    COUNT(*) as media_count,
    COUNT(DISTINCT EXTRACT(YEAR FROM captured_at)) as year_count,
    array_agg(DISTINCT EXTRACT(YEAR FROM captured_at) ORDER BY EXTRACT(YEAR FROM captured_at) DESC) as years,
    MIN(captured_at) as earliest_date,
    MAX(captured_at) as latest_date
  FROM media
),
album_stats AS (
  SELECT COUNT(*) as album_count FROM albums
),
storage_stats AS (
  SELECT 
    SUM(file_size) as total_size,
    AVG(file_size) as avg_size
  FROM media 
  WHERE file_size IS NOT NULL
)
SELECT 
  m.media_count,
  a.album_count,
  m.years,
  m.earliest_date,
  m.latest_date,
  s.total_size,
  s.avg_size
FROM media_stats m, album_stats a, storage_stats s;

-- Benefits: Single query, comprehensive stats, better performance
```

### 5. **Query Plan Analysis & Monitoring**
```sql
-- Analyze slow queries with detailed execution plans
EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) 
SELECT * FROM media 
WHERE album_id = $1 
ORDER BY captured_at DESC 
LIMIT 20;

-- Look for performance issues:
-- 1. Sequential scans instead of index usage
-- 2. High cost operations (>1000 cost units)
-- 3. Memory spills to disk (temp files)
-- 4. Nested loop joins with large datasets

-- Enable query logging in PostgreSQL
ALTER SYSTEM SET log_min_duration_statement = 100; -- Log queries >100ms
ALTER SYSTEM SET log_statement = 'all'; -- Log all statements (dev only)
ALTER SYSTEM SET log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h ';

-- Query performance monitoring queries
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    stddev_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements 
ORDER BY total_time DESC 
LIMIT 10;
```

### 6. **Connection Pool Optimization**
```go
// Current: Basic connection settings in config.go
type Database struct {
    URI                string `debugmap:"visible"`
    SSL                bool   `debugmap:"visible"`
    MaxOpenConnections int    `debugmap:"visible" default:"10"`
    Debug              bool   `debugmap:"visible" default:"false"`
}

// Optimized: Comprehensive connection tuning
type Database struct {
    URI                    string        `debugmap:"visible"`
    SSL                    bool          `debugmap:"visible"`
    MaxOpenConnections     int           `debugmap:"visible" default:"25"`
    MaxIdleConnections     int           `debugmap:"visible" default:"5"`
    ConnMaxLifetime        time.Duration `debugmap:"visible" default:"5m"`
    ConnMaxIdleTime        time.Duration `debugmap:"visible" default:"2m"`
    QueryTimeout           time.Duration `debugmap:"visible" default:"30s"`
    SlowQueryThreshold     time.Duration `debugmap:"visible" default:"100ms"`
    EnableQueryLogging     bool          `debugmap:"visible" default:"false"`
    Debug                  bool          `debugmap:"visible" default:"false"`
}

// Connection pool monitoring
func (db *Database) GetPoolStats() sql.DBStats {
    return db.conn.Stats()
}

// Ideal ratios:
// - MaxOpen should be ~2-4x number of CPU cores
// - MaxIdle should be ~25% of MaxOpen
// - ConnMaxLifetime prevents connection leaks
// - ConnMaxIdleTime reduces idle resource usage
```

### 7. **Query-Specific Optimizations**

#### Media Gallery Loading
```go
// Current: Loading full media objects
type Media struct {
    ID          string    `json:"id"`
    Href        string    `json:"href"`
    AlbumHref   string    `json:"albumHref"`
    CapturedAt  string    `json:"capturedAt"`
    Type        string    `json:"type"`
    Filename    string    `json:"filename"`
    Thumbnail   string    `json:"thumbnail"`
    Content     string    `json:"content"`
    Exif        []ExifHeader `json:"exif"`
    FileSize    int64     `json:"fileSize,omitempty"`
    Dimensions  string    `json:"dimensions,omitempty"`
}

// Optimized: Separate light and full media types
type MediaThumbnail struct {
    ID         string `json:"id"`
    Thumbnail  string `json:"thumbnail"`
    Filename   string `json:"filename"`
    Type       string `json:"type"`
    CapturedAt string `json:"capturedAt"`
}

type MediaFull struct {
    MediaThumbnail
    Content    string       `json:"content"`
    Exif       []ExifHeader `json:"exif"`
    FileSize   int64        `json:"fileSize"`
    Dimensions string       `json:"dimensions"`
    AlbumHref  string       `json:"albumHref"`
}

// Gallery endpoint: Only load thumbnails
func (h *Handler) GetAlbumThumbnails(albumID string) ([]MediaThumbnail, error) {
    query := `
        SELECT id, filename, thumbnail, type, captured_at 
        FROM media 
        WHERE album_id = $1 
        ORDER BY captured_at DESC`
    // Much faster due to smaller data transfer
}

// Detail endpoint: Load full media when clicked
func (h *Handler) GetMediaDetail(mediaID string) (*MediaFull, error) {
    // Load complete media object with EXIF data
}
```

#### Date Range Filtering Optimization
```sql
-- Common query: Get media for specific year (timeline view)
SELECT id, filename, thumbnail, type, captured_at 
FROM media 
WHERE EXTRACT(YEAR FROM captured_at) = $1
ORDER BY captured_at DESC;

-- Optimized with date range instead of function
SELECT id, filename, thumbnail, type, captured_at 
FROM media 
WHERE captured_at >= $1::date 
  AND captured_at < ($1::date + INTERVAL '1 year')
ORDER BY captured_at DESC;

-- Month-based filtering for timeline navigation
SELECT id, filename, thumbnail, type, captured_at 
FROM media 
WHERE captured_at >= $1::date 
  AND captured_at < ($1::date + INTERVAL '1 month')
ORDER BY captured_at DESC;
```

### 8. **Monitoring & Profiling Implementation**
```go
// Query timing middleware
type QueryLogger struct {
    logger *zap.Logger
    slowQueryThreshold time.Duration
}

func (q *QueryLogger) LogQuery(ctx context.Context, query string, args []interface{}, duration time.Duration) {
    if duration > q.slowQueryThreshold {
        q.logger.Warn("Slow query detected",
            zap.String("query", query),
            zap.Any("args", args),
            zap.Duration("duration", duration),
            zap.String("caller", getCaller()),
        )
    }
}

// Database metrics collection
type DBMetrics struct {
    QueryDuration *prometheus.HistogramVec
    QueryCount    *prometheus.CounterVec
    PoolStats     *prometheus.GaugeVec
}

func (m *DBMetrics) RecordQuery(operation string, duration time.Duration) {
    m.QueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
    m.QueryCount.WithLabelValues(operation).Inc()
}

// Health check queries
func (db *Database) HealthCheck() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return db.conn.PingContext(ctx)
}
```

## Implementation Priority

### Phase 1: Immediate Improvements (1-2 weeks)
1. **Add strategic indexes** for common query patterns
2. **Implement query timing middleware** to identify bottlenecks
3. **Optimize connection pool settings** based on workload
4. **Add database health checks**

### Phase 2: Performance Enhancements (2-4 weeks)
1. **Implement cursor-based pagination** for large result sets
2. **Optimize media loading** with selective field loading
3. **Add query monitoring dashboard** with Prometheus metrics
4. **Implement N+1 query detection and fixes**

### Phase 3: Advanced Optimizations (1-2 months)
1. **Date range query optimization** for timeline features
2. **Query plan analysis automation** with alerting
3. **Database-level optimizations** (PostgreSQL tuning)
4. **Caching layer integration** (Redis for frequent queries)

## Expected Performance Improvements

- **Pagination**: 10x faster for deep pages (cursor vs offset)
- **Gallery loading**: 3-5x faster with selective field loading
- **Timeline queries**: 5-10x faster with proper date range indexes
- **Aggregations**: 2-3x faster with optimized queries
- **Connection efficiency**: 20-30% better resource utilization

## Monitoring Success Metrics

- **Response times**: Target <100ms for 95th percentile
- **Database connection utilization**: <80% of pool capacity
- **Query success rate**: >99.9%
- **Slow query count**: <5% of total queries
- **Cache hit ratio**: >90% for frequently accessed data

The goal is to maintain sub-100ms response times even with thousands of photos and hundreds of albums while providing a foundation for future scaling.