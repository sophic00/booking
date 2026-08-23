package email

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"ticket-booking/backend/internal/config"
)

func TestTicketConfirmationTemplate(t *testing.T) {
	now := time.Date(2026, 8, 24, 18, 30, 0, 0, time.UTC)
	info := TicketConfirmation{
		CustomerName:     "Alice Smith",
		EventTitle:       "Rock Festival 2026",
		EventStartTime:   &now,
		VenueName:        "Grand Arena",
		VenueCity:        "London",
		BookingRef:       "REF-12345",
		Seats: []SeatInfo{
			{RowLabel: "A", SeatNumber: "1", CategoryName: "VIP", UnitPrice: 120.0},
			{RowLabel: "A", SeatNumber: "2", CategoryName: "VIP", UnitPrice: 120.0},
		},
		TotalAmount:      240.0,
		Currency:         "GBP",
		QRAttachmentName: "ticket-ref12345.png",
	}

	var buf bytes.Buffer
	err := ticketConfirmationTmpl.Execute(&buf, info)
	if err != nil {
		t.Fatalf("failed to execute ticket confirmation template: %v", err)
	}

	rendered := buf.String()
	t.Logf("Rendered HTML:\n%s", rendered)

	if !strings.Contains(rendered, "Alice Smith") {
		t.Errorf("expected customer name in rendered HTML")
	}
	if !strings.Contains(rendered, "REF-12345") {
		t.Errorf("expected booking ref in rendered HTML")
	}
	if !strings.Contains(rendered, "Rock Festival 2026") {
		t.Errorf("expected event title in rendered HTML")
	}
	if !strings.Contains(rendered, "Grand Arena") || !strings.Contains(rendered, "London") {
		t.Errorf("expected venue name and city in rendered HTML")
	}
	if !strings.Contains(rendered, `src="cid:ticket-ref12345.png"`) {
		t.Errorf("expected cid image src in rendered HTML")
	}
	if !strings.Contains(rendered, "240.00 GBP") {
		t.Errorf("expected total amount formatted in rendered HTML")
	}
	if strings.Contains(rendered, "ZgotmplZ") {
		t.Errorf("unexpected ZgotmplZ sanitization in rendered HTML: %s", rendered)
	}
}

func TestWaitlistOfferTemplate(t *testing.T) {
	now := time.Date(2026, 8, 24, 18, 30, 0, 0, time.UTC)
	expiresAt := now.Add(30 * time.Minute)
	info := WaitlistOfferNotification{
		CustomerName:   "Bob Jones",
		EventTitle:     "Jazz Night",
		EventStartTime: &now,
		CategoryName:   "Balcony",
		RowLabel:       "B",
		SeatNumber:     "14",
		Price:          45.5,
		Currency:       "USD",
		OfferURL:       "https://tickets.example.com/waitlist/offer?token=xyz",
		ExpiresAt:      expiresAt,
	}

	var buf bytes.Buffer
	err := waitlistOfferTmpl.Execute(&buf, info)
	if err != nil {
		t.Fatalf("failed to execute waitlist offer template: %v", err)
	}

	rendered := buf.String()
	t.Logf("Rendered HTML:\n%s", rendered)

	if !strings.Contains(rendered, "Bob Jones") {
		t.Errorf("expected customer name in rendered HTML")
	}
	if !strings.Contains(rendered, "Balcony") {
		t.Errorf("expected category name in rendered HTML")
	}
	if !strings.Contains(rendered, "Jazz Night") {
		t.Errorf("expected event title in rendered HTML")
	}
	if !strings.Contains(rendered, "Row B, Seat 14") {
		t.Errorf("expected seat details in rendered HTML")
	}
	if !strings.Contains(rendered, "45.50 USD") {
		t.Errorf("expected price in rendered HTML")
	}
	if !strings.Contains(rendered, `href="https://tickets.example.com/waitlist/offer?token=xyz"`) {
		t.Errorf("expected offer URL link in rendered HTML")
	}
	if !strings.Contains(rendered, expiresAt.Format(time.RFC1123)) {
		t.Errorf("expected formatted expiresAt in rendered HTML")
	}
	if strings.Contains(rendered, "ZgotmplZ") {
		t.Errorf("unexpected ZgotmplZ sanitization in rendered HTML: %s", rendered)
	}
}

func TestEmailService_Mock(t *testing.T) {
	svc := NewService(&config.Config{EmailMock: true})
	if !svc.IsMock() {
		t.Fatalf("expected IsMock to be true")
	}

	err := svc.SendTicketConfirmation("alice@example.com", TicketConfirmation{
		CustomerName: "Alice",
		BookingRef:   "REF-1",
		EventTitle:   "Concert",
	})
	if err != nil {
		t.Fatalf("expected SendTicketConfirmation to succeed in mock mode: %v", err)
	}

	err = svc.SendWaitlistOffer("bob@example.com", WaitlistOfferNotification{
		CustomerName: "Bob",
		EventTitle:   "Concert",
		ExpiresAt:    time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("expected SendWaitlistOffer to succeed in mock mode: %v", err)
	}
}
