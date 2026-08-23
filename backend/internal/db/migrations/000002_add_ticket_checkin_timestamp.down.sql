DROP INDEX IF EXISTS idx_tickets_event_status;
DROP INDEX IF EXISTS idx_tickets_qr_payload;
ALTER TABLE tickets DROP COLUMN IF EXISTS checked_in_at;
