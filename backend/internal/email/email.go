// Package email delivers transactional ticket and waitlist emails.
//
// When cfg.EmailMock is true (default), no network calls are made: emails are
// logged to stdout so local development and tests work without an SMTP relay.
// Otherwise, messages are sent over SMTP using the configured server.
package email

import (
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/smtp"
	"strings"
	"time"

	"ticket-booking/backend/internal/config"
)

// SeatInfo describes a single seat on a booking confirmation email.
type SeatInfo struct {
	RowLabel     string
	SeatNumber   string
	CategoryName string
	UnitPrice    float64
}

// TicketConfirmation carries everything needed for a booking confirmation email.
type TicketConfirmation struct {
	CustomerName     string
	EventTitle       string
	EventStartTime   *time.Time
	VenueName        string
	VenueCity        string
	BookingRef       string
	Seats            []SeatInfo
	TotalAmount      float64
	Currency         string
	QRPNG            []byte // PNG bytes of the QR code encoding the ticket payload
	QRAttachmentName string
}

// WaitlistOfferNotification carries everything needed for a time-limited offer email.
type WaitlistOfferNotification struct {
	CustomerName   string
	EventTitle     string
	EventStartTime *time.Time
	CategoryName   string
	RowLabel       string
	SeatNumber     string
	Price          float64
	Currency       string
	OfferURL       string // time-limited link to complete the booking
	ExpiresAt      time.Time
}

// Service sends emails via SMTP or logs them when running in mock mode.
type Service struct {
	cfg *config.Config
}

// NewService builds an email Service from configuration.
func NewService(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

// IsMock reports whether emails are only logged instead of delivered.
func (s *Service) IsMock() bool {
	return s.cfg == nil || s.cfg.EmailMock
}

//go:embed templates/*.html
var templatesFS embed.FS

var (
	emailFuncMap = template.FuncMap{
		"formatRFC1123": func(v any) string {
			switch t := v.(type) {
			case time.Time:
				return t.Format(time.RFC1123)
			case *time.Time:
				if t != nil {
					return t.Format(time.RFC1123)
				}
			}
			return ""
		},
	}

	ticketConfirmationTmpl = template.Must(
		template.New("ticket_confirmation.html").
			Funcs(emailFuncMap).
			ParseFS(templatesFS, "templates/ticket_confirmation.html"),
	)

	waitlistOfferTmpl = template.Must(
		template.New("waitlist_offer.html").
			Funcs(emailFuncMap).
			ParseFS(templatesFS, "templates/waitlist_offer.html"),
	)
)

// SendTicketConfirmation delivers the booking confirmation with the QR code ticket attached inline.
func (s *Service) SendTicketConfirmation(to string, info TicketConfirmation) error {
	if s.IsMock() {
		log.Printf("📧 [EMAIL MOCK] Ticket confirmation → to=%s ref=%s event=%q seats=%d qr_png_bytes=%d",
			to, info.BookingRef, info.EventTitle, len(info.Seats), len(info.QRPNG))
		return nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Hi %s,\n\nYour booking is confirmed!\n\n", info.CustomerName))
	sb.WriteString(fmt.Sprintf("Booking Reference: %s\nEvent: %s\n", info.BookingRef, info.EventTitle))
	if info.EventStartTime != nil {
		sb.WriteString(fmt.Sprintf("Show Time: %s\n", info.EventStartTime.Format(time.RFC1123)))
	}
	if info.VenueName != "" {
		sb.WriteString(fmt.Sprintf("Venue: %s, %s\n", info.VenueName, info.VenueCity))
	}
	sb.WriteString("\nSeats:\n")
	for _, seat := range info.Seats {
		sb.WriteString(fmt.Sprintf("  - Row %s, Seat %s (%s): %.2f %s\n",
			seat.RowLabel, seat.SeatNumber, seat.CategoryName, seat.UnitPrice, info.Currency))
	}
	sb.WriteString(fmt.Sprintf("\nTotal Paid: %.2f %s\n", info.TotalAmount, info.Currency))
	sb.WriteString("\nYour QR code ticket is attached to this email. Present it at the venue for entry.\n")
	textBody := sb.String()

	var htmlBuf bytes.Buffer
	if err := ticketConfirmationTmpl.Execute(&htmlBuf, info); err != nil {
		return fmt.Errorf("failed to render ticket confirmation html: %w", err)
	}
	htmlBody := htmlBuf.String()

	var attachments []attachment
	if len(info.QRPNG) > 0 {
		attachments = append(attachments, attachment{
			name:        info.QRAttachmentName,
			contentType: "image/png",
			data:        info.QRPNG,
			inlineCID:   info.QRAttachmentName,
		})
	}

	msg, err := buildMultipartMessage(s.cfg.SMTPFrom, to,
		fmt.Sprintf("Your tickets for %s [%s]", info.EventTitle, info.BookingRef),
		textBody, htmlBody, attachments)
	if err != nil {
		return err
	}
	return s.send(to, msg)
}

// SendWaitlistOffer notifies the next waitlisted customer about a time-limited seat offer.
func (s *Service) SendWaitlistOffer(to string, info WaitlistOfferNotification) error {
	if s.IsMock() {
		log.Printf("📧 [EMAIL MOCK] Waitlist offer → to=%s event=%q seat=%s%s link=%s expires=%s",
			to, info.EventTitle, info.RowLabel, info.SeatNumber, info.OfferURL, info.ExpiresAt.Format(time.RFC3339))
		return nil
	}

	subject := fmt.Sprintf("A seat opened up for %s: offer expires %s",
		info.EventTitle, info.ExpiresAt.Format("3:04 PM MST"))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Hi %s,\n\nGood news! A seat in your waitlisted category just became available.\n\n",
		info.CustomerName))
	sb.WriteString(fmt.Sprintf("Event: %s\nSeat: Row %s, Seat %s (%s)\nPrice: %.2f %s\n\n",
		info.EventTitle, info.RowLabel, info.SeatNumber, info.CategoryName, info.Price, info.Currency))
	sb.WriteString(fmt.Sprintf("Complete your booking before %s using this link:\n%s\n\n",
		info.ExpiresAt.Format(time.RFC1123), info.OfferURL))
	sb.WriteString("If you do not complete the booking in time, the seat will be offered to the next person in line.\n")
	textBody := sb.String()

	var htmlBuf bytes.Buffer
	if err := waitlistOfferTmpl.Execute(&htmlBuf, info); err != nil {
		return fmt.Errorf("failed to render waitlist offer html: %w", err)
	}
	htmlBody := htmlBuf.String()

	msg, err := buildMultipartMessage(s.cfg.SMTPFrom, to, subject, textBody, htmlBody, nil)
	if err != nil {
		return err
	}
	return s.send(to, msg)
}

func (s *Service) send(to string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPServer, s.cfg.SMTPPort)
	var auth smtp.Auth
	if s.cfg.SMTPUsername != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPServer)
	}
	return smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{to}, msg)
}

type attachment struct {
	name        string
	contentType string
	data        []byte
	inlineCID   string
}

func buildMultipartMessage(from, to, subject, textBody, htmlBody string, attachments []attachment) ([]byte, error) {
	boundaryBytes := make([]byte, 12)
	if _, err := rand.Read(boundaryBytes); err != nil {
		return nil, fmt.Errorf("failed to generate MIME boundary: %w", err)
	}
	boundary := "tbs-" + base64.RawURLEncoding.EncodeToString(boundaryBytes)

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject)))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%q\r\n\r\n", boundary))

	writePart := func(headers, body string) {
		buf.WriteString("--" + boundary + "\r\n")
		buf.WriteString(headers)
		buf.WriteString("\r\n")
		buf.WriteString(body)
		buf.WriteString("\r\n\r\n")
	}

	writePart("Content-Type: text/plain; charset=\"utf-8\"\r\nContent-Transfer-Encoding: 8bit\r\n", textBody)
	writePart("Content-Type: text/html; charset=\"utf-8\"\r\nContent-Transfer-Encoding: 8bit\r\n", htmlBody)

	for _, att := range attachments {
		headers := fmt.Sprintf(
			"Content-Type: %s; name=%q\r\nContent-Transfer-Encoding: base64\r\nContent-Disposition: inline;\r\n filename=%q\r\nContent-ID: <%s>\r\n",
			att.contentType, att.name, att.name, att.inlineCID)
		encoded := base64.StdEncoding.EncodeToString(att.data)
		var wrapped strings.Builder
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			wrapped.WriteString(encoded[i:end])
			wrapped.WriteString("\r\n")
		}
		writePart(headers, wrapped.String())
	}

	buf.WriteString("--" + boundary + "--\r\n")
	return buf.Bytes(), nil
}
