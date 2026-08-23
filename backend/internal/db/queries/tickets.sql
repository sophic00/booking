-- name: GetTicketByID :one
SELECT 
    t.id,
    t.booking_id,
    t.event_id,
    t.seat_id,
    t.unit_price,
    t.qr_code_payload,
    t.status,
    t.created_at,
    t.checked_in_at,
    b.booking_reference,
    b.status AS booking_status,
    b.customer_id,
    u.full_name AS customer_name,
    u.email AS customer_email,
    e.organiser_id,
    e.title AS event_title,
    e.start_time AS event_start_time,
    e.end_time AS event_end_time,
    e.status AS event_status,
    e.poster_url AS event_poster_url,
    v.id AS venue_id,
    v.name AS venue_name,
    v.address AS venue_address,
    v.city AS venue_city,
    s.row_label,
    s.seat_number,
    s.grid_row,
    s.grid_col,
    sc.id AS seat_category_id,
    sc.name AS category_name,
    sc.color_code AS category_color
FROM tickets t
JOIN bookings b ON t.booking_id = b.id
JOIN users u ON b.customer_id = u.id
JOIN events e ON t.event_id = e.id
JOIN venues v ON e.venue_id = v.id
JOIN seats s ON t.seat_id = s.id
JOIN seat_categories sc ON s.seat_category_id = sc.id
WHERE t.id = $1 LIMIT 1;

-- name: GetTicketByQRPayload :one
SELECT 
    t.id,
    t.booking_id,
    t.event_id,
    t.seat_id,
    t.unit_price,
    t.qr_code_payload,
    t.status,
    t.created_at,
    t.checked_in_at,
    b.booking_reference,
    b.status AS booking_status,
    b.customer_id,
    u.full_name AS customer_name,
    u.email AS customer_email,
    e.organiser_id,
    e.title AS event_title,
    e.start_time AS event_start_time,
    e.end_time AS event_end_time,
    e.status AS event_status,
    e.poster_url AS event_poster_url,
    v.id AS venue_id,
    v.name AS venue_name,
    v.address AS venue_address,
    v.city AS venue_city,
    s.row_label,
    s.seat_number,
    s.grid_row,
    s.grid_col,
    sc.id AS seat_category_id,
    sc.name AS category_name,
    sc.color_code AS category_color
FROM tickets t
JOIN bookings b ON t.booking_id = b.id
JOIN users u ON b.customer_id = u.id
JOIN events e ON t.event_id = e.id
JOIN venues v ON e.venue_id = v.id
JOIN seats s ON t.seat_id = s.id
JOIN seat_categories sc ON s.seat_category_id = sc.id
WHERE t.qr_code_payload = $1 LIMIT 1;

-- name: CheckInTicketByID :one
UPDATE tickets
SET status = 'CHECKED_IN',
    checked_in_at = NOW()
WHERE id = $1 AND status = 'VALID'
RETURNING *;

-- name: CheckInTicketByQRPayload :one
UPDATE tickets
SET status = 'CHECKED_IN',
    checked_in_at = NOW()
WHERE qr_code_payload = $1 AND status = 'VALID'
RETURNING *;

-- name: ListTicketsByEventID :many
SELECT 
    t.id,
    t.booking_id,
    t.event_id,
    t.seat_id,
    t.unit_price,
    t.qr_code_payload,
    t.status,
    t.created_at,
    t.checked_in_at,
    b.booking_reference,
    b.customer_id,
    u.full_name AS customer_name,
    u.email AS customer_email,
    s.row_label,
    s.seat_number,
    sc.name AS category_name,
    sc.color_code AS category_color
FROM tickets t
JOIN bookings b ON t.booking_id = b.id
JOIN users u ON b.customer_id = u.id
JOIN seats s ON t.seat_id = s.id
JOIN seat_categories sc ON s.seat_category_id = sc.id
WHERE t.event_id = $1
ORDER BY t.created_at DESC;

-- name: GetEventCheckInStats :one
SELECT 
    COUNT(*)::bigint AS total_tickets,
    COUNT(CASE WHEN status = 'VALID' THEN 1 END)::bigint AS valid_count,
    COUNT(CASE WHEN status = 'CHECKED_IN' THEN 1 END)::bigint AS checked_in_count,
    COUNT(CASE WHEN status = 'CANCELLED' THEN 1 END)::bigint AS cancelled_count
FROM tickets
WHERE event_id = $1;
