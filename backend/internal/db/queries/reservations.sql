-- name: CreateSeatReservation :one
INSERT INTO seat_reservations (
    event_id,
    seat_id,
    user_id,
    status,
    hold_token,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: CheckSeatAvailable :one
SELECT s.id AS seat_id, s.venue_id, s.seat_category_id, ep.price, ep.currency
FROM seats s
JOIN events e ON e.venue_id = s.venue_id
JOIN event_pricing ep ON ep.event_id = e.id AND ep.seat_category_id = s.seat_category_id
WHERE e.id = $1 AND s.id = $2 AND s.is_active = TRUE
  AND NOT EXISTS (
      SELECT 1 FROM seat_reservations sr
      WHERE sr.event_id = $1 AND sr.seat_id = $2
        AND (
            sr.status = 'BOOKED'
            OR (sr.status IN ('HELD', 'OFFERED') AND sr.expires_at > NOW())
        )
  )
LIMIT 1;

-- name: GetActiveHoldByToken :many
SELECT 
    sr.*,
    s.row_label,
    s.seat_number,
    s.seat_category_id,
    ep.price,
    ep.currency,
    sc.name AS category_name,
    sc.color_code AS category_color
FROM seat_reservations sr
JOIN seats s ON sr.seat_id = s.id
JOIN seat_categories sc ON s.seat_category_id = sc.id
JOIN event_pricing ep ON ep.event_id = sr.event_id AND ep.seat_category_id = s.seat_category_id
WHERE sr.event_id = $1 AND sr.hold_token = $2 AND sr.status = 'HELD' AND sr.expires_at > NOW();

-- name: ReleaseSeatHoldByToken :execrows
UPDATE seat_reservations
SET status = 'RELEASED', updated_at = NOW()
WHERE event_id = $1 AND hold_token = $2 AND status = 'HELD';

-- name: ReleaseSeatHoldBySeatAndUser :execrows
UPDATE seat_reservations
SET status = 'RELEASED', updated_at = NOW()
WHERE event_id = $1 AND seat_id = $2 AND user_id = $3 AND status = 'HELD';

-- name: GetEventSeatMapWithStatus :many
SELECT 
    s.id AS seat_id,
    s.venue_id,
    s.seat_category_id,
    sc.name AS category_name,
    sc.color_code AS category_color,
    s.row_label,
    s.seat_number,
    s.grid_row,
    s.grid_col,
    COALESCE(ep.price, 0.00) AS price,
    COALESCE(ep.currency, 'USD') AS currency,
    CASE 
        WHEN sr.status = 'BOOKED' THEN 'BOOKED'
        WHEN sr.status = 'HELD' AND sr.expires_at > NOW() THEN 'HELD'
        WHEN sr.status = 'OFFERED' AND sr.expires_at > NOW() THEN 'OFFERED'
        ELSE 'AVAILABLE'
    END AS computed_status,
    sr.user_id AS held_by_user_id,
    sr.expires_at AS hold_expires_at,
    sr.hold_token
FROM seats s
JOIN events e ON e.venue_id = s.venue_id
JOIN seat_categories sc ON s.seat_category_id = sc.id
LEFT JOIN event_pricing ep ON ep.event_id = e.id AND ep.seat_category_id = s.seat_category_id
LEFT JOIN seat_reservations sr ON sr.event_id = e.id AND sr.seat_id = s.id 
    AND (
        sr.status = 'BOOKED' 
        OR (sr.status IN ('HELD', 'OFFERED') AND sr.expires_at > NOW())
    )
WHERE e.id = $1 AND s.is_active = TRUE
ORDER BY s.grid_row ASC, s.grid_col ASC;

-- name: GetExpiredSeatHolds :many
SELECT * FROM seat_reservations
WHERE status = 'HELD' AND expires_at <= NOW();

-- name: BulkReleaseExpiredHolds :execrows
UPDATE seat_reservations
SET status = 'RELEASED', updated_at = NOW()
WHERE status = 'HELD' AND expires_at <= NOW();

-- name: ConfirmReservationToBooked :execrows
UPDATE seat_reservations
SET status = 'BOOKED', booking_id = $3, updated_at = NOW()
WHERE event_id = $1 AND hold_token = $2 AND status IN ('HELD', 'OFFERED');

-- name: ReleaseExpiredSeatHold :execrows
UPDATE seat_reservations
SET status = 'RELEASED', updated_at = NOW()
WHERE event_id = $1 AND seat_id = $2 AND status = 'HELD' AND expires_at <= NOW();

