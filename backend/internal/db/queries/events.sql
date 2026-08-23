-- name: CreateEvent :one
INSERT INTO events (
    organiser_id,
    venue_id,
    title,
    description,
    event_type,
    poster_url,
    start_time,
    end_time,
    hold_ttl_seconds,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: UpdateEvent :one
UPDATE events
SET title = $2,
    description = $3,
    event_type = $4,
    poster_url = $5,
    start_time = $6,
    end_time = $7,
    hold_ttl_seconds = $8,
    status = $9,
    updated_at = NOW()
WHERE id = $1 AND organiser_id = $10
RETURNING *;

-- name: PublishEvent :one
UPDATE events
SET status = 'PUBLISHED',
    updated_at = NOW()
WHERE id = $1 AND organiser_id = $2
RETURNING *;

-- name: CancelEvent :one
UPDATE events
SET status = 'CANCELLED',
    updated_at = NOW()
WHERE id = $1 AND organiser_id = $2
RETURNING *;

-- name: GetEventByID :one
SELECT 
    e.*,
    v.name AS venue_name,
    v.address AS venue_address,
    v.city AS venue_city,
    v.state AS venue_state,
    v.country AS venue_country,
    v.total_capacity AS venue_capacity,
    u.full_name AS organiser_name,
    u.email AS organiser_email
FROM events e
JOIN venues v ON e.venue_id = v.id
JOIN users u ON e.organiser_id = u.id
WHERE e.id = $1 LIMIT 1;

-- name: ListPublishedEvents :many
SELECT 
    e.*,
    v.name AS venue_name,
    v.city AS venue_city,
    v.total_capacity AS venue_capacity
FROM events e
JOIN venues v ON e.venue_id = v.id
WHERE e.status = 'PUBLISHED'
  AND (sqlc.narg('event_type')::event_type IS NULL OR e.event_type = sqlc.narg('event_type'))
  AND (sqlc.narg('search')::text IS NULL OR e.title ILIKE '%' || sqlc.narg('search') || '%' OR v.city ILIKE '%' || sqlc.narg('search') || '%')
  AND (sqlc.narg('from_date')::timestamptz IS NULL OR e.start_time >= sqlc.narg('from_date'))
  AND (sqlc.narg('to_date')::timestamptz IS NULL OR e.start_time <= sqlc.narg('to_date'))
ORDER BY e.start_time ASC
LIMIT $1 OFFSET $2;

-- name: ListOrganiserEvents :many
SELECT 
    e.*,
    v.name AS venue_name,
    v.city AS venue_city,
    v.total_capacity AS venue_capacity
FROM events e
JOIN venues v ON e.venue_id = v.id
WHERE e.organiser_id = $1
ORDER BY e.created_at DESC;

-- name: SetEventPricing :one
INSERT INTO event_pricing (
    event_id,
    seat_category_id,
    price,
    currency
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (event_id, seat_category_id)
DO UPDATE SET 
    price = EXCLUDED.price,
    currency = EXCLUDED.currency
RETURNING *;

-- name: GetEventPricingByEventID :many
SELECT 
    ep.*,
    sc.name AS category_name,
    sc.description AS category_description,
    sc.color_code AS category_color
FROM event_pricing ep
JOIN seat_categories sc ON ep.seat_category_id = sc.id
WHERE ep.event_id = $1
ORDER BY ep.price DESC;

-- name: DeleteEventPricingByEventID :exec
DELETE FROM event_pricing
WHERE event_id = $1;
