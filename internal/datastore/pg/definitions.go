package pg

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

const (
	albumsTable = "albums"
	mediaTable  = "media"
	// Albums table columns
	albumID          = "id"
	albumCreatedAt   = "created_at"
	albumPath        = "path"
	albumDescription = "description"
	albumParentID    = "parent_id"
	albumThumbnailID = "thumbnail_id"

	// Albums table columns for join scenarios
	albumChildID          = "child.id as child_id"
	albumChildCreatedAt   = "child.created_at as child_created_at"
	albumChildPath        = "child.path as child_path"
	albumChildDescription = "child.description as child_description"
	albumChildThumbnailID = "child.thumbnail_id as child_thumbnail_id"

	// Media table columns
	mediaID         = "id"
	mediaCreatedAt  = "created_at"
	mediaCapturedAt = "captured_at"
	mediaAlbumID    = "album_id"
	mediaFileName   = "file_name"
	mediaThumbnail  = "thumbnail"
	mediaExif       = "exif"
	mediaMediaType  = "media_type"
	mediaHash       = "hash"

	// Album table columns for media select join scenarios
	albumMediaCreatedAt   = "albums.created_at as album_created_at"
	albumMediaPath        = "albums.path as album_path"
	albumMediaDescription = "albums.description as album_description"
	albumMediaThumbnailID = "albums.thumbnail_id as album_thumbnail_id"

	// Media table columns for join scenarios
	mediaIDJoin         = "media.id as media_id"
	mediaCapturedAtJoin = "media.captured_at as media_captured_at"
	mediaAlbumIDJoin    = "media.album_id as media_album_id"
	mediaFileNameJoin   = "media.file_name as media_file_name"
	mediaThumbnailJoin  = "media.thumbnail as media_thumbnail"
	mediaExifJoin       = "media.exif as media_exif"
	mediaMediaTypeJoin  = "media.media_type as media_media_type"

	// Definition for zed token and lock
	lockKey        = "zed_token_lock_key"
	zedTable       = "zed_token"
	globalAdvisory = "pg_advisory_xact_lock"
	sharedAdvisory = "pg_advisory_xact_lock_shared"
	globalLockStmt = "SELECT pg_advisory_xact_lock(%d);"
	sharedLockStmt = "SELECT pg_advisory_xact_lock_shared(%d);"
	writeStmt      = "INSERT INTO zed_token (id, token) VALUES (1, $1) ON CONFLICT (id) DO UPDATE SET token = excluded.token;"
)

var (
	psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	listAlbumStmt = psql.Select(
		preffix(albumsTable, albumID),
		preffix(albumsTable, albumCreatedAt),
		preffix(albumsTable, albumPath),
		preffix(albumsTable, albumDescription),
		preffix(albumsTable, albumParentID),
		preffix(albumsTable, albumThumbnailID),
		albumChildID,
		albumChildCreatedAt,
		albumChildPath,
		albumChildDescription,
		albumChildThumbnailID,
		mediaIDJoin,
		mediaCapturedAtJoin,
		mediaAlbumIDJoin,
		mediaFileNameJoin,
		mediaThumbnailJoin,
		mediaExifJoin,
		mediaMediaTypeJoin,
	).
		From(albumsTable).
		LeftJoin("albums as child on child.parent_id = albums.id").
		LeftJoin("media as media on media.album_id = albums.id")

	listAlbumsStmt = psql.Select(
		preffix(albumsTable, albumID),
		preffix(albumsTable, albumCreatedAt),
		preffix(albumsTable, albumPath),
		preffix(albumsTable, albumDescription),
		preffix(albumsTable, albumParentID),
		preffix(albumsTable, albumThumbnailID),
		"count_media_in_album_with_children(albums.id) as media_count",
		albumChildID,
		albumChildCreatedAt,
		albumChildPath,
		albumChildDescription,
		albumChildThumbnailID,
	).
		From(albumsTable).
		LeftJoin("albums as child on child.parent_id = albums.id")

	listMediaStmt = psql.Select(
		preffix(mediaTable, mediaID),
		preffix(mediaTable, mediaCreatedAt),
		preffix(mediaTable, mediaCapturedAt),
		preffix(mediaTable, mediaAlbumID),
		preffix(mediaTable, mediaFileName),
		preffix(mediaTable, mediaThumbnail),
		preffix(mediaTable, mediaHash),
		preffix(mediaTable, mediaExif),
		preffix(mediaTable, mediaMediaType),
		albumMediaCreatedAt,
		albumMediaPath,
		albumMediaDescription,
		albumMediaThumbnailID,
	).
		From(mediaTable).
		InnerJoin("albums on albums.id = media.album_id")

	statAlbumMediaStmt = `select (select count(*) from albums) as total_albums, (select count(*) from media) as total_media;`

	statYearsStmt = `SELECT DISTINCT EXTRACT(YEAR FROM captured_at)::INTEGER AS year FROM media WHERE captured_at IS NOT NULL ORDER BY year DESC;`

	tokenWriteStmt  = psql.Insert(zedTable).Columns("id", "token")
	selectTokenStmt = psql.Select("token").From(zedTable).Limit(1)
)

func preffix(preffix, columnName string) string {
	return fmt.Sprintf("%s.%s", preffix, columnName)
}
