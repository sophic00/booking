-- name: JoinWaitlist :one
INSERT INTO waitlist_entries (
    event_id,
    seat_category_id,
    customer_id,
    status
) VALUES (
    $1, $2, $3, 'WAITING'
) RETURNING *;

-- name: GetCustomerWaitlistEntry :one
SELECT * FROM waitlist_entries
WHERE event_id = $1 AND seat_category_id = $2 AND customer_id = $3 AND status = 'WAITING'
LIMIT 1;

-- name: GetCustomerWaitlists :many
SELECT 
    we.*,
    e.title AS event_title,
    e.start_time AS event_start_time,
    e.event_type,
    sc.name AS category_name,
    sc.color_code AS category_color,
    (
        SELECT COUNT(*) + 1 
        FROM waitlist_entries prior 
        WHERE prior.event_id = we.event_id 
          AND prior.seat_category_id = we.seat_category_id 
          AND prior.status = 'WAITING' 
          AND prior.created_at < we.created_at
    ) AS queue_position,
    wo.offer_token,
    wo.expires_at AS offer_expires_at
FROM waitlist_entries we
JOIN events e ON we.event_id = e.id
JOIN seat_categories sc ON we.seat_category_id = sc.id
LEFT JOIN waitlist_offers wo ON wo.waitlist_entry_id = we.id AND wo.status = 'PENDING' AND wo.expires_at > NOW()
WHERE we.customer_id = $1 AND we.status IN ('WAITING', 'OFFERED')
ORDER BY we.created_at DESC;

-- name: GetNextWaitlistCandidate :one
SELECT 
    we.*,
    u.email AS customer_email,
    u.full_name AS customer_name
FROM waitlist_entries we
JOIN users u ON we.customer_id = u.id
WHERE we.event_id = $1 AND we.seat_category_id = $2 AND we.status = 'WAITING'
ORDER BY we.created_at ASC
LIMIT 1
FOR UPDATE OF we SKIP LOCKED;

-- name: UpdateWaitlistEntryStatus :one
UPDATE waitlist_entries
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CreateWaitlistOffer :one
INSERT INTO waitlist_offers (
    waitlist_entry_id,
    event_id,
    seat_id,
    offer_token,
    offered_at,
    expires_at,
    status
) VALUES (
    $1, $2, $3, $4, NOW(), $5, 'PENDING'
) RETURNING *;

-- name: GetWaitlistOfferByToken :one
SELECT 
    wo.*,
    we.customer_id,
    we.seat_category_id,
    u.email AS customer_email,
    u.full_name AS customer_name,
    e.title AS event_title,
    e.start_time AS event_start_time,
    e.end_time AS event_end_time,
    s.row_label,
    s.seat_number,
    sc.name AS category_name,
    COALESCE(ep.price, 0.00) AS price,
    COALESCE(ep.currency, 'USD') AS currency
FROM waitlist_offers wo
JOIN waitlist_entries we ON wo.waitlist_entry_id = we.id
JOIN users u ON we.customer_id = u.id
JOIN events e ON wo.event_id = e.id
JOIN seats s ON wo.seat_id = s.id
JOIN seat_categories sc ON s.seat_category_id = sc.id
LEFT JOIN event_pricing ep ON ep.event_id = e.id AND ep.seat_category_id = s.seat_category_id
WHERE wo.offer_token = $1
LIMIT 1;

-- name: AcceptWaitlistOffer :one
UPDATE waitlist_offers
SET status = 'ACCEPTED',
    updated_at = NOW()
WHERE id = $1 AND status = 'PENDING' AND expires_at > NOW()
RETURNING *;

-- name: ExpireWaitlistOffer :one
UPDATE waitlist_offers
SET status = 'EXPIRED',
    updated_at = NOW()
WHERE id = $1 AND status = 'PENDING'
RETURNING *;

-- name: GetPendingExpiredOffers :many
SELECT 
    wo.*,
    we.seat_category_id
FROM waitlist_offers wo
JOIN waitlist_entries we ON wo.waitlist_entry_id = we.id
WHERE wo.status = 'PENDING' AND wo.expires_at <= NOW();

-- name: RevokeWaitlistOffer :one
UPDATE waitlist_offers
SET status = 'REVOKED',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CountAvailableSeatsInCategory :one
SELECT COUNT(*)
FROM seats s
JOIN events e ON e.venue_id = s.venue_id AND e.id = $1
WHERE s.seat_category_id = $2
  AND s.is_active = TRUE
  AND NOT EXISTS (
      SELECT 1 FROM seat_reservations sr
      WHERE sr.event_id = e.id AND sr.seat_id = s.id
        AND (
            sr.status = 'BOOKED'
            OR (sr.status IN ('HELD', 'OFFERED') AND sr.expires_at > NOW())
        )
  );

-- name: GetWaitlistQueuePosition :one
SELECT COUNT(*)::int + 1 AS position
FROM waitlist_entries
WHERE event_id = $1
  AND seat_category_id = $2
  AND status = 'WAITING'
  AND created_at < $3;

-- name: GetBookingFreedSeats :many
SELECT sr.seat_id, s.seat_category_id
FROM seat_reservations sr
JOIN seats s ON s.id = sr.seat_id
WHERE sr.booking_id = $1;

-- name: GetActiveHoldOrOfferByToken :many
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
WHERE sr.event_id = $1 AND sr.hold_token = $2
  AND sr.status IN ('HELD', 'OFFERED')
  AND sr.expires_at > NOW();

-- name: RevokeOfferedSeatReservation :execrows
UPDATE seat_reservations
SET status = 'RELEASED', updated_at = NOW()
WHERE event_id = $1 AND seat_id = $2 AND status = 'OFFERED';
