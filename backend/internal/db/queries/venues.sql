-- name: CreateVenue :one
INSERT INTO venues (
    name,
    address,
    city,
    state,
    country,
    total_capacity,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: ListVenues :many
SELECT * FROM venues
ORDER BY name ASC;

-- name: GetVenueByID :one
SELECT * FROM venues
WHERE id = $1 LIMIT 1;

-- name: UpdateVenue :one
UPDATE venues
SET name = $2,
    address = $3,
    city = $4,
    state = $5,
    country = $6,
    total_capacity = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteVenue :exec
DELETE FROM venues
WHERE id = $1;

-- name: CreateSeatCategory :one
INSERT INTO seat_categories (
    name,
    description,
    color_code
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: ListSeatCategories :many
SELECT * FROM seat_categories
ORDER BY name ASC;

-- name: GetSeatCategoryByID :one
SELECT * FROM seat_categories
WHERE id = $1 LIMIT 1;

-- name: GetSeatCategoryByName :one
SELECT * FROM seat_categories
WHERE name = $1 LIMIT 1;

-- name: CreateSeat :one
INSERT INTO seats (
    venue_id,
    seat_category_id,
    row_label,
    seat_number,
    grid_row,
    grid_col,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetSeatsByVenueID :many
SELECT 
    s.*,
    sc.name AS category_name,
    sc.color_code AS category_color
FROM seats s
JOIN seat_categories sc ON s.seat_category_id = sc.id
WHERE s.venue_id = $1 AND s.is_active = TRUE
ORDER BY s.grid_row ASC, s.grid_col ASC;

-- name: DeleteSeatsByVenueID :exec
DELETE FROM seats
WHERE venue_id = $1;
