package qrcode

import (
	"encoding/base64"
	"fmt"

	"github.com/skip2/go-qrcode"
)

// GeneratePNG generates raw PNG image bytes encoding the given content.
func GeneratePNG(content string, size int) ([]byte, error) {
	if size <= 0 {
		size = 256
	}
	pngBytes, err := qrcode.Encode(content, qrcode.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	return pngBytes, nil
}

// GenerateDataURL generates a base64 Data URL for embedding the QR code in HTML/JSON.
func GenerateDataURL(content string, size int) (string, error) {
	pngBytes, err := GeneratePNG(content, size)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(pngBytes)
	return fmt.Sprintf("data:image/png;base64,%s", encoded), nil
}

// GenerateTicketPayload creates a standardized verifiable QR payload string for a ticket.
func GenerateTicketPayload(bookingRef, seatID, ticketID string) string {
	return fmt.Sprintf("TICKET|REF:%s|SEAT:%s|ID:%s", bookingRef, seatID, ticketID)
}
