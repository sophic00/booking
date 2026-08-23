-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- ENUMS
-- ============================================================================
CREATE TYPE user_role AS ENUM ('ADMIN', 'ORGANISER', 'CUSTOMER');
CREATE TYPE event_type AS ENUM ('MOVIE', 'CONCERT', 'THEATRE', 'SPORTS', 'OTHER');
CREATE TYPE event_status AS ENUM ('DRAFT', 'PUBLISHED', 'CANCELLED', 'COMPLETED');
CREATE TYPE reservation_status AS ENUM ('HELD', 'OFFERED', 'BOOKED', 'RELEASED', 'CANCELLED');
CREATE TYPE booking_status AS ENUM ('CONFIRMED', 'CANCELLED');
CREATE TYPE ticket_status AS ENUM ('VALID', 'CANCELLED', 'CHECKED_IN');
CREATE TYPE waitlist_status AS ENUM ('WAITING', 'OFFERED', 'ACCEPTED', 'EXPIRED', 'CANCELLED');
CREATE TYPE offer_status AS ENUM ('PENDING', 'ACCEPTED', 'EXPIRED', 'REVOKED');

-- ============================================================================
-- 1. USERS & AUTH (RBAC)
-- ============================================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(150) NOT NULL,
    phone VARCHAR(30),
    role user_role NOT NULL DEFAULT 'CUSTOMER',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 2. VENUES, SEAT CATEGORIES & STATIC SEATS
-- ============================================================================
CREATE TABLE venues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    country VARCHAR(100) NOT NULL DEFAULT 'IN',
    total_capacity INT NOT NULL DEFAULT 0,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE seat_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    color_code VARCHAR(10) NOT NULL DEFAULT '#3B82F6',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE seats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    venue_id UUID NOT NULL REFERENCES venues(id) ON DELETE CASCADE,
    seat_category_id UUID NOT NULL REFERENCES seat_categories(id) ON DELETE RESTRICT,
    row_label VARCHAR(10) NOT NULL,
    seat_number VARCHAR(10) NOT NULL,
    grid_row INT NOT NULL,
    grid_col INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_venue_row_seat UNIQUE (venue_id, row_label, seat_number),
    CONSTRAINT uq_venue_grid_coord UNIQUE (venue_id, grid_row, grid_col)
);

CREATE INDEX idx_seats_venue_category ON seats(venue_id, seat_category_id);

-- ============================================================================
-- 3. EVENTS & PRICING
-- ============================================================================
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organiser_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    venue_id UUID NOT NULL REFERENCES venues(id) ON DELETE RESTRICT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    event_type event_type NOT NULL DEFAULT 'MOVIE',
    poster_url TEXT,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    hold_ttl_seconds INT NOT NULL DEFAULT 600,
    status event_status NOT NULL DEFAULT 'DRAFT',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_event_time_order CHECK (end_time > start_time)
);

CREATE INDEX idx_events_browse ON events(status, start_time) WHERE status = 'PUBLISHED';
CREATE INDEX idx_events_organiser ON events(organiser_id);
CREATE INDEX idx_events_venue ON events(venue_id);

CREATE TABLE event_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    seat_category_id UUID NOT NULL REFERENCES seat_categories(id) ON DELETE RESTRICT,
    price NUMERIC(10, 2) NOT NULL CHECK (price >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_event_category_pricing UNIQUE (event_id, seat_category_id)
);

-- ============================================================================
-- 4. BOOKINGS & TICKETS
-- ============================================================================
CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_reference VARCHAR(32) NOT NULL UNIQUE,
    customer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE RESTRICT,
    total_amount NUMERIC(10, 2) NOT NULL CHECK (total_amount >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status booking_status NOT NULL DEFAULT 'CONFIRMED',
    cancellation_reason TEXT,
    cancelled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bookings_customer ON bookings(customer_id, created_at DESC);
CREATE INDEX idx_bookings_event_status ON bookings(event_id, status);

CREATE TABLE tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE RESTRICT,
    seat_id UUID NOT NULL REFERENCES seats(id) ON DELETE RESTRICT,
    unit_price NUMERIC(10, 2) NOT NULL CHECK (unit_price >= 0),
    qr_code_payload TEXT NOT NULL UNIQUE,
    status ticket_status NOT NULL DEFAULT 'VALID',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_event_seat_booking UNIQUE (event_id, seat_id, booking_id)
);

CREATE INDEX idx_tickets_booking ON tickets(booking_id);

-- ============================================================================
-- 5. SEAT RESERVATIONS (SPARSE CONCURRENCY & HOLD TTL ENGINE)
-- ============================================================================
CREATE TABLE seat_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    seat_id UUID NOT NULL REFERENCES seats(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status reservation_status NOT NULL DEFAULT 'HELD',
    hold_token UUID NOT NULL DEFAULT gen_random_uuid(),
    expires_at TIMESTAMPTZ,
    booking_id UUID REFERENCES bookings(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_event_seat_booked 
ON seat_reservations (event_id, seat_id) 
WHERE status = 'BOOKED';

CREATE UNIQUE INDEX uq_event_seat_active_hold_offer 
ON seat_reservations (event_id, seat_id) 
WHERE status IN ('HELD', 'OFFERED');

CREATE INDEX idx_seat_reservations_expiry 
ON seat_reservations (status, expires_at) 
WHERE status IN ('HELD', 'OFFERED');

CREATE INDEX idx_seat_reservations_event_user 
ON seat_reservations (event_id, user_id, status);

-- ============================================================================
-- 6. WAITLIST & TIME-LIMITED AUTO-ASSIGNMENT OFFERS
-- ============================================================================
CREATE TABLE waitlist_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    seat_category_id UUID NOT NULL REFERENCES seat_categories(id) ON DELETE RESTRICT,
    customer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status waitlist_status NOT NULL DEFAULT 'WAITING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_active_customer_waitlist UNIQUE (event_id, seat_category_id, customer_id, status)
);

CREATE INDEX idx_waitlist_fifo 
ON waitlist_entries (event_id, seat_category_id, created_at ASC) 
WHERE status = 'WAITING';

CREATE TABLE waitlist_offers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    waitlist_entry_id UUID NOT NULL REFERENCES waitlist_entries(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    seat_id UUID NOT NULL REFERENCES seats(id) ON DELETE CASCADE,
    offer_token UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    offered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    status offer_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_waitlist_offers_pending_expiry 
ON waitlist_offers (status, expires_at) 
WHERE status = 'PENDING';
