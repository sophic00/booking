package qrcode

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePNG(t *testing.T) {
	content := "https://ticketbooking.example.com/tickets/12345"
	pngBytes, err := GeneratePNG(content, 256)
	require.NoError(t, err)
	assert.NotEmpty(t, pngBytes)
	// PNG magic header: \x89PNG\r\n\x1a\n
	assert.True(t, len(pngBytes) > 8)
	assert.Equal(t, byte(0x89), pngBytes[0])
	assert.Equal(t, "PNG", string(pngBytes[1:4]))
}

func TestGenerateDataURL(t *testing.T) {
	content := "TICKET|REF:TB-2026-001|SEAT:A1"
	dataURL, err := GenerateDataURL(content, 200)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(dataURL, "data:image/png;base64,"))
	assert.True(t, len(dataURL) > 30)
}

func TestGenerateTicketPayload(t *testing.T) {
	payload := GenerateTicketPayload("TB-20260823-123456", "seat-uuid-1", "ticket-uuid-2")
	assert.Equal(t, "TICKET|REF:TB-20260823-123456|SEAT:seat-uuid-1|ID:ticket-uuid-2", payload)
}
