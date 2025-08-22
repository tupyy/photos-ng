// Package pg provides PostgreSQL datastore implementation with query filtering capabilities.
//
// This file contains query filter functions that can be applied to SQL SELECT statements
// to add WHERE clauses, LIMIT/OFFSET clauses, and other query modifications. These filters
// follow the functional composition pattern, allowing multiple filters to be chained together.
package pg

import (
	"time"

	sq "github.com/Masterminds/squirrel"
)

// FilterByAlbumId creates a filter that adds a WHERE clause to match a specific ID.
// If the provided ID is empty, the filter is a no-op and returns the original query unchanged.
// This filter specifically targets the "id" column.
//
// Parameters:
//   - id: The ID to filter by. If empty, no filtering is applied.
//
// Returns: A QueryOption function that can be applied to a SelectBuilder.
func FilterByAlbumId(id string) QueryOption {
	return FilterByColumnName("albums.id", id)
}

func FilterByMediaId(id string) QueryOption {
	return FilterByColumnName("media.id", id)
}

func FilterAlbumByParentId(parentId string) QueryOption {
	return func(orig sq.SelectBuilder) sq.SelectBuilder {
		if parentId == "" {
			return orig.Where(sq.Eq{"albums.parent_id": nil})
		}
		return orig.Where(sq.Eq{"albums.parent_id": parentId})
	}
}

func FilterAlbumWithParents(hasParent bool) QueryOption {
	return func(orig sq.SelectBuilder) sq.SelectBuilder {
		if !hasParent {
			return orig.Where("albums.parent_id IS NULL")
		}
		return orig
	}
}

// Limit creates a filter that adds a LIMIT clause to restrict the number of results.
// If the limit is 0 or negative, no LIMIT clause is added to the query.
//
// Parameters:
//   - limit: The maximum number of rows to return. If 0 or negative, no limit is applied.
//
// Returns: A QueryOption function that can be applied to a SelectBuilder.
func Limit(limit int) QueryOption {
	return func(orig sq.SelectBuilder) sq.SelectBuilder {
		if limit <= 0 {
			return orig
		}
		return orig.Limit(uint64(limit))
	}
}

// Offset creates a filter that adds an OFFSET clause to skip a number of rows.
// This is commonly used in conjunction with LimitQueryFilter for pagination.
// If the offset is negative, no OFFSET clause is added to the query.
//
// Parameters:
//   - offset: The number of rows to skip from the beginning of the result set.
//
// Returns: A QueryOption function that can be applied to a SelectBuilder.
func Offset(offset int) QueryOption {
	return func(orig sq.SelectBuilder) sq.SelectBuilder {
		if offset < 0 {
			return orig
		}
		return orig.Offset(uint64(offset))
	}
}

// SortByColumn creates a filter that adds an ORDER BY clause to sort the results.
// Supports both ascending and descending order. If the column name is empty,
// the filter is a no-op and returns the original query unchanged.
//
// Parameters:
//   - column: The column name to sort by. If empty, no sorting is applied.
//   - descending: If true, sorts in descending order; if false, sorts in ascending order.
//
// Returns: A QueryOption function that can be applied to a SelectBuilder.
//
// Example:
//   - SortByColumn("created_at", true) generates "ORDER BY created_at DESC"
//   - SortByColumn("name", false) generates "ORDER BY name ASC"
func SortByColumn(column string, descending bool) QueryOption {
	return func(orig sq.SelectBuilder) sq.SelectBuilder {
		if column == "" {
			return orig
		}
		if descending {
			return orig.OrderBy(column + " DESC")
		}
		return orig.OrderBy(column + " ASC")
	}
}

// SortBy creates a filter that adds an ORDER BY clause with multiple sort criteria.
// This allows for complex sorting with multiple columns and different sort directions.
// If the sortBy slice is empty, the filter is a no-op and returns the original query unchanged.
//
// Parameters:
//   - sortBy: A slice of sort criteria strings. Each string should be in the format
//     "column_name" for ASC or "column_name DESC" for DESC.
//
// Returns: A QueryOption function that can be applied to a SelectBuilder.
//
// Example:
//   - SortBy([]string{"created_at DESC", "name ASC"})
//   - SortBy([]string{"priority", "updated_at DESC"})
func SortBy(sortBy []string) QueryOption {
	return func(orig sq.SelectBuilder) sq.SelectBuilder {
		if len(sortBy) == 0 {
			return orig
		}
		return orig.OrderBy(sortBy...)
	}
}

// FilterByColumnName creates a filter that adds a WHERE clause to match a specific value in any column.
// If the provided value is empty or column name is empty, the filter is a no-op and returns the original query unchanged.
//
// Parameters:
//   - columnName: The name of the column to filter on (e.g., "album_id", "type", "name", etc.)
//   - value: The value to filter by. If empty, no filtering is applied.
//
// Returns: A QueryOption function that can be applied to a SelectBuilder.
//
// Example:
//   - FilterByColumnName("album_id", "123")
//   - FilterByColumnName("type", "photo")
func FilterByColumnName(columnName string, value string) QueryOption {
	return func(orig sq.SelectBuilder) sq.SelectBuilder {
		if columnName == "" || value == "" {
			return orig
		}
		return orig.Where(sq.Eq{columnName: value})
	}
}

func FilterByMediaDate(start, end *time.Time) func(orig sq.SelectBuilder) sq.SelectBuilder {
	return func(orig sq.SelectBuilder) sq.SelectBuilder {
		if start != nil && end != nil {
			return orig.Where(
				sq.And{
					sq.GtOrEq{"captured_at": start},
					sq.LtOrEq{"captured_at": end},
				},
			)
		}

		if start != nil {
			return orig.Where(sq.GtOrEq{"captured_at": start})
		}

		if end != nil {
			return orig.Where(sq.LtOrEq{"captured_at": end})
		}

		return orig
	}
}

// FilterByCursor creates a filter for cursor-based pagination using captured_at and id.
// This implements the WHERE condition: captured_at < cursor_time OR (captured_at = cursor_time AND id < cursor_id)
// which allows for consistent pagination regardless of page depth.
//
// Parameters:
//   - capturedAt: The captured_at timestamp from the cursor
//   - id: The id from the cursor for tie-breaking when timestamps are equal
//
// Returns: A QueryOption function that can be applied to a SelectBuilder.
func FilterByCursor(capturedAt time.Time, id string) QueryOption {
	return func(orig sq.SelectBuilder) sq.SelectBuilder {
		return orig.Where(
			sq.Or{
				sq.Lt{"media.captured_at": capturedAt},
				sq.And{
					sq.Eq{"media.captured_at": capturedAt},
					sq.Lt{"media.id": id},
				},
			},
		)
	}
}
