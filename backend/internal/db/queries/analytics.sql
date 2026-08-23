-- name: GetEventBookingSummary :one
SELECT 
    e.id AS event_id,
    e.title AS event_title,
    e.status AS event_status,
    e.start_time,
    v.total_capacity,
    COUNT(DISTINCT CASE WHEN b.status = 'CONFIRMED' THEN b.id END) AS confirmed_bookings_count,
    COUNT(DISTINCT CASE WHEN b.status = 'CANCELLED' THEN b.id END) AS cancelled_bookings_count,
    COUNT(DISTINCT CASE WHEN t.status IN ('VALID', 'CHECKED_IN') THEN t.id END) AS valid_tickets_count,
    COUNT(DISTINCT CASE WHEN t.status = 'CHECKED_IN' THEN t.id END) AS checked_in_tickets_count,
    COALESCE(SUM(CASE WHEN t.status IN ('VALID', 'CHECKED_IN') THEN t.unit_price ELSE 0 END), 0.00) AS total_revenue,
    COALESCE(
        ROUND((COUNT(DISTINCT CASE WHEN t.status IN ('VALID', 'CHECKED_IN') THEN t.id END)::numeric / NULLIF(v.total_capacity, 0)::numeric) * 100, 2),
        0.00
    ) AS occupancy_percentage,
    (
        SELECT COUNT(*) 
        FROM waitlist_entries we 
        WHERE we.event_id = e.id AND we.status = 'WAITING'
    ) AS waitlist_waiting_count
FROM events e
JOIN venues v ON e.venue_id = v.id
LEFT JOIN bookings b ON b.event_id = e.id
LEFT JOIN tickets t ON t.booking_id = b.id AND b.status = 'CONFIRMED'
WHERE e.id = $1 AND e.organiser_id = $2
GROUP BY e.id, e.title, e.status, e.start_time, v.total_capacity;

-- name: GetEventCategoryBreakdown :many
SELECT 
    sc.id AS seat_category_id,
    sc.name AS category_name,
    sc.color_code,
    COUNT(DISTINCT s.id) AS total_seats,
    COUNT(DISTINCT CASE WHEN t.status IN ('VALID', 'CHECKED_IN') THEN t.id END) AS booked_seats,
    COALESCE(SUM(CASE WHEN t.status IN ('VALID', 'CHECKED_IN') THEN t.unit_price ELSE 0 END), 0.00) AS revenue,
    (
        SELECT COUNT(*) 
        FROM waitlist_entries we 
        WHERE we.event_id = $1 AND we.seat_category_id = sc.id AND we.status = 'WAITING'
    ) AS waitlist_count
FROM seats s
JOIN events e ON e.venue_id = s.venue_id AND e.id = $1
JOIN seat_categories sc ON s.seat_category_id = sc.id
LEFT JOIN tickets t ON t.event_id = e.id AND t.seat_id = s.id AND t.status IN ('VALID', 'CHECKED_IN')
WHERE s.is_active = TRUE
GROUP BY sc.id, sc.name, sc.color_code
ORDER BY sc.name ASC;
