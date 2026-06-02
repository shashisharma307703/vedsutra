-- name: UpsertClass :one
INSERT INTO class_levels (
    class_level_id,
    class_name,
    class_code,
    display_order,
    description,
    is_lower_primary,
    is_upper_primary,
    is_secondary,
    is_higher_secondary
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (class_level_id) 
DO UPDATE SET
    class_name = EXCLUDED.class_name,
    class_code = EXCLUDED.class_code,
    display_order = EXCLUDED.display_order,
    description = EXCLUDED.description,
    is_lower_primary = EXCLUDED.is_lower_primary,
    is_upper_primary = EXCLUDED.is_upper_primary,
    is_secondary = EXCLUDED.is_secondary,
    is_higher_secondary = EXCLUDED.is_higher_secondary
    -- created_at remains unchanged (no updated_at column)
RETURNING *;

-- name: GetClassByID :one
SELECT 
    class_level_id,
    class_name,
    class_code,
    display_order,
    description,
    is_lower_primary,
    is_upper_primary,
    is_secondary,
    is_higher_secondary,
    created_at
FROM class_levels 
WHERE class_level_id = $1
LIMIT 1;

-- name: ReplaceClass :one
UPDATE class_levels
SET
    class_name = $2,
    class_code = $3,
    display_order = $4,
    description = $5,
    is_lower_primary = $6,
    is_upper_primary = $7,
    is_secondary = $8,
    is_higher_secondary = $9
WHERE class_level_id = $1
RETURNING *;

-- name: DeleteClass :exec
DELETE FROM class_levels 
WHERE class_level_id = $1;