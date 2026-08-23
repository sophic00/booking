-- ============================================================================
-- DOWN MIGRATION: 000001_init_schema.down.sql
-- ============================================================================

DROP TABLE IF EXISTS waitlist_offers CASCADE;
DROP TABLE IF EXISTS waitlist_entries CASCADE;
DROP TABLE IF EXISTS seat_reservations CASCADE;
DROP TABLE IF EXISTS tickets CASCADE;
DROP TABLE IF EXISTS bookings CASCADE;
DROP TABLE IF EXISTS event_pricing CASCADE;
DROP TABLE IF EXISTS events CASCADE;
DROP TABLE IF EXISTS seats CASCADE;
DROP TABLE IF EXISTS seat_categories CASCADE;
DROP TABLE IF EXISTS venues CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP TYPE IF EXISTS offer_status CASCADE;
DROP TYPE IF EXISTS waitlist_status CASCADE;
DROP TYPE IF EXISTS ticket_status CASCADE;
DROP TYPE IF EXISTS booking_status CASCADE;
DROP TYPE IF EXISTS reservation_status CASCADE;
DROP TYPE IF EXISTS event_status CASCADE;
DROP TYPE IF EXISTS event_type CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;
