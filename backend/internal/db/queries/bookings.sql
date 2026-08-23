-- name: CreateBooking :one
INSERT INTO bookings (
    booking_reference,
    customer_id,
    event_id,
    total_amount,
    currency,
    status
) VALUES (
    $1, $2, $3, $4, $5, 'CONFIRMED'
) RETURNING *;

-- name: CreateTicket :one
INSERT INTO tickets (
    booking_id,
    event_id,
    seat_id,
    unit_price,
    qr_code_payload,
    status
) VALUES (
    $1, $2, $3, $4, $5, 'VALID'
) RETURNING *;

-- name: GetBookingByID :one
SELECT 
    b.*,
    e.title AS event_title,
    e.start_time AS event_start_time,
    e.end_time AS event_end_time,
    e.poster_url AS event_poster_url,
    e.event_type,
    v.name AS venue_name,
    v.address AS venue_address,
    v.city AS venue_city,
    u.full_name AS customer_name,
    u.email AS customer_email
FROM bookings b
JOIN events e ON b.event_id = e.id
JOIN venues v ON e.venue_id = v.id
JOIN users u ON b.customer_id = u.id
WHERE b.id = $1 LIMIT 1;

-- name: GetBookingByReference :one
SELECT 
    b.*,
    e.title AS event_title,
    e.start_time AS event_start_time,
    e.end_time AS event_end_time,
    e.poster_url AS event_poster_url,
    e.event_type,
    v.name AS venue_name,
    v.address AS venue_address,
    v.city AS venue_city,
    u.full_name AS customer_name,
    u.email AS customer_email
FROM bookings b
JOIN events e ON b.event_id = e.id
JOIN venues v ON e.venue_id = v.id
JOIN users u ON b.customer_id = u.id
WHERE b.booking_reference = $1 LIMIT 1;

-- name: GetTicketsByBookingID :many
SELECT 
    t.*,
    s.row_label,
    s.seat_number,
    s.grid_row,
    s.grid_col,
    sc.id AS seat_category_id,
    sc.name AS category_name,
    sc.color_code AS category_color
FROM tickets t
JOIN seats s ON t.seat_id = s.id
JOIN seat_categories sc ON s.seat_category_id = sc.id
WHERE t.booking_id = $1
ORDER BY s.row_label ASC, s.seat_number ASC;

-- name: GetCustomerBookings :many
SELECT 
    b.*,
    e.title AS event_title,
    e.start_time AS event_start_time,
    e.event_type,
    e.poster_url AS event_poster_url,
    v.name AS venue_name,
    v.city AS venue_city,
    COUNT(t.id) AS ticket_count
FROM bookings b
JOIN events e ON b.event_id = e.id
JOIN venues v ON e.venue_id = v.id
LEFT JOIN tickets t ON t.booking_id = b.id
WHERE b.customer_id = $1
GROUP BY b.id, e.id, v.id
ORDER BY b.created_at DESC;

-- name: CancelBooking :one
UPDATE bookings
SET status = 'CANCELLED',
    cancellation_reason = $2,
    cancelled_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND customer_id = $3 AND status = 'CONFIRMED'
RETURNING *;

-- name: CancelBookingTickets :many
UPDATE tickets
SET status = 'CANCELLED'
WHERE booking_id = $1 AND status = 'VALID'
RETURNING *;

-- name: CancelBookingReservations :many
UPDATE seat_reservations
SET status = 'CANCELLED',
    updated_at = NOW()
WHERE booking_id = $1 AND status = 'BOOKED'
RETURNING *;
