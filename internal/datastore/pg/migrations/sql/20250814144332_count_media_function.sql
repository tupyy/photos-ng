-- +goose Up
-- +goose StatementBegin
CREATE IF NOT EXISTS FUNCTION count_media_in_album_with_children(album_id_param VARCHAR(255))
RETURNS INTEGER AS $$
DECLARE
    total_count INTEGER := 0;
BEGIN
    -- Use a recursive CTE to find all child albums and count their media
    WITH RECURSIVE album_tree AS (
        -- Base case: the specified album
        SELECT id, path
        FROM albums 
        WHERE id = album_id_param
        
        UNION ALL
        
        -- Recursive case: find all children
        SELECT a.id, a.path
        FROM albums a
        INNER JOIN album_tree at ON a.parent_id = at.id
    )
    SELECT COUNT(m.id) INTO total_count
    FROM album_tree at
    LEFT JOIN media m ON m.album_id = at.id;
    
    RETURN total_count;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION count_media_in_album_with_children;
-- +goose StatementEnd
