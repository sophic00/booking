ALTER TABLE tickets ADD COLUMN IF NOT EXISTS checked_in_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_tickets_qr_payload ON tickets(qr_code_payload);
CREATE INDEX IF NOT EXISTS idx_tickets_event_status ON tickets(event_id, status);
