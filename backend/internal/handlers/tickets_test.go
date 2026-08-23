package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ticket-booking/backend/internal/auth"
	"ticket-booking/backend/internal/config"
	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/middleware"
	"ticket-booking/backend/internal/utils"
)

type mockTicketQuerier struct {
	generated.Querier
	getTicketByIDFunc          func(ctx context.Context, id pgtype.UUID) (generated.GetTicketByIDRow, error)
	getTicketByQRPayloadFunc   func(ctx context.Context, qrCodePayload string) (generated.GetTicketByQRPayloadRow, error)
	checkInTicketByIDFunc      func(ctx context.Context, id pgtype.UUID) (generated.Ticket, error)
	checkInTicketByQRPayloadFunc func(ctx context.Context, qrCodePayload string) (generated.Ticket, error)
	listTicketsByEventIDFunc   func(ctx context.Context, eventID pgtype.UUID) ([]generated.ListTicketsByEventIDRow, error)
	getEventCheckInStatsFunc   func(ctx context.Context, eventID pgtype.UUID) (generated.GetEventCheckInStatsRow, error)
	getEventByIDFunc           func(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error)
}

func (m *mockTicketQuerier) GetTicketByID(ctx context.Context, id pgtype.UUID) (generated.GetTicketByIDRow, error) {
	if m.getTicketByIDFunc != nil {
		return m.getTicketByIDFunc(ctx, id)
	}
	return generated.GetTicketByIDRow{}, pgx.ErrNoRows
}

func (m *mockTicketQuerier) GetTicketByQRPayload(ctx context.Context, qrCodePayload string) (generated.GetTicketByQRPayloadRow, error) {
	if m.getTicketByQRPayloadFunc != nil {
		return m.getTicketByQRPayloadFunc(ctx, qrCodePayload)
	}
	return generated.GetTicketByQRPayloadRow{}, pgx.ErrNoRows
}

func (m *mockTicketQuerier) CheckInTicketByID(ctx context.Context, id pgtype.UUID) (generated.Ticket, error) {
	if m.checkInTicketByIDFunc != nil {
		return m.checkInTicketByIDFunc(ctx, id)
	}
	return generated.Ticket{}, pgx.ErrNoRows
}

func (m *mockTicketQuerier) CheckInTicketByQRPayload(ctx context.Context, qrCodePayload string) (generated.Ticket, error) {
	if m.checkInTicketByQRPayloadFunc != nil {
		return m.checkInTicketByQRPayloadFunc(ctx, qrCodePayload)
	}
	return generated.Ticket{}, pgx.ErrNoRows
}

func (m *mockTicketQuerier) ListTicketsByEventID(ctx context.Context, eventID pgtype.UUID) ([]generated.ListTicketsByEventIDRow, error) {
	if m.listTicketsByEventIDFunc != nil {
		return m.listTicketsByEventIDFunc(ctx, eventID)
	}
	return []generated.ListTicketsByEventIDRow{}, nil
}

func (m *mockTicketQuerier) GetEventCheckInStats(ctx context.Context, eventID pgtype.UUID) (generated.GetEventCheckInStatsRow, error) {
	if m.getEventCheckInStatsFunc != nil {
		return m.getEventCheckInStatsFunc(ctx, eventID)
	}
	return generated.GetEventCheckInStatsRow{}, nil
}

func (m *mockTicketQuerier) GetEventByID(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error) {
	if m.getEventByIDFunc != nil {
		return m.getEventByIDFunc(ctx, id)
	}
	return generated.GetEventByIDRow{}, pgx.ErrNoRows
}

func setupTicketRouter(handler *TicketHandler, secret string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Authenticate(secret))
	r.Post("/api/v1/tickets/verify", handler.VerifyTicket)
	r.Get("/api/v1/tickets/verify", handler.VerifyTicket)
	r.Group(func(r chi.Router) {
		r.Use(middleware.OrganiserOrAdmin)
		r.Post("/api/v1/tickets/check-in", handler.CheckInTicket)
		r.Post("/api/v1/tickets/{id}/check-in", handler.CheckInTicket)
		r.Get("/api/v1/organiser/events/{id}/tickets", handler.ListEventTickets)
	})
	return r
}

func createSampleTicketRow(ticketID, bookingID, customerID, eventID, organiserID, seatID uuid.UUID, qrPayload string, status generated.TicketStatus) generated.GetTicketByQRPayloadRow {
	return generated.GetTicketByQRPayloadRow{
		ID:               utils.UUIDToPgtype(ticketID),
		BookingID:        utils.UUIDToPgtype(bookingID),
		EventID:          utils.UUIDToPgtype(eventID),
		SeatID:           utils.UUIDToPgtype(seatID),
		UnitPrice:        utils.Float64ToPgtypeNumeric(85.00),
		QrCodePayload:    qrPayload,
		Status:           status,
		CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
		BookingReference: "TB-20260823-XYZ",
		BookingStatus:    generated.BookingStatusCONFIRMED,
		CustomerID:       utils.UUIDToPgtype(customerID),
		CustomerName:     "John Doe",
		CustomerEmail:    "john@example.com",
		OrganiserID:      utils.UUIDToPgtype(organiserID),
		EventTitle:       "Hans Zimmer Live",
		EventStartTime:   pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
		EventEndTime:     pgtype.Timestamptz{Time: time.Now().Add(27 * time.Hour), Valid: true},
		EventStatus:      generated.EventStatusPUBLISHED,
		VenueID:          utils.UUIDToPgtype(uuid.New()),
		VenueName:        "Royal Arena",
		VenueAddress:     "100 Arena Way",
		VenueCity:        "London",
		RowLabel:         "A",
		SeatNumber:       "12",
		GridRow:          1,
		GridCol:          12,
		SeatCategoryID:   utils.UUIDToPgtype(uuid.New()),
		CategoryName:     "VIP",
		CategoryColor:    "#EF4444",
	}
}

func TestVerifyTicket_ByQRPayload_Success(t *testing.T) {
	secret := "test-secret"
	customerID := uuid.New()
	organiserID := uuid.New()
	eventID := uuid.New()
	bookingID := uuid.New()
	ticketID := uuid.New()
	seatID := uuid.New()
	qrPayload := "TICKET|REF:TB-20260823-XYZ|SEAT:12|ID:" + ticketID.String()

	row := createSampleTicketRow(ticketID, bookingID, customerID, eventID, organiserID, seatID, qrPayload, generated.TicketStatusVALID)

	mock := &mockTicketQuerier{
		getTicketByQRPayloadFunc: func(ctx context.Context, payload string) (generated.GetTicketByQRPayloadRow, error) {
			assert.Equal(t, qrPayload, payload)
			return row, nil
		},
	}

	handler := NewTicketHandler(mock, nil, &config.Config{JWTSecret: secret})
	router := setupTicketRouter(handler, secret)

	body, _ := json.Marshal(VerifyTicketRequest{QRPayload: qrPayload})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/verify", bytes.NewReader(body))
	token, _, _ := auth.GenerateToken(customerID, "john@example.com", "CUSTOMER", secret, 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool                       `json:"success"`
		Data    TicketVerificationResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.True(t, resp.Data.IsValid)
	assert.True(t, resp.Data.CanCheckIn)
	assert.Equal(t, ticketID.String(), resp.Data.Ticket.ID)
	assert.Equal(t, "VALID", resp.Data.Ticket.Status)
	assert.Equal(t, "John Doe", resp.Data.Booking.CustomerName)
	assert.Equal(t, "Hans Zimmer Live", resp.Data.Event.Title)
	assert.Equal(t, "VIP", resp.Data.Seat.CategoryName)
}

func TestVerifyTicket_ByTicketID_Success(t *testing.T) {
	secret := "test-secret"
	customerID := uuid.New()
	organiserID := uuid.New()
	eventID := uuid.New()
	bookingID := uuid.New()
	ticketID := uuid.New()
	seatID := uuid.New()
	qrPayload := "TICKET|REF:TB-20260823-XYZ|SEAT:12|ID:" + ticketID.String()

	row := createSampleTicketRow(ticketID, bookingID, customerID, eventID, organiserID, seatID, qrPayload, generated.TicketStatusVALID)

	mock := &mockTicketQuerier{
		getTicketByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.GetTicketByIDRow, error) {
			assert.Equal(t, utils.UUIDToPgtype(ticketID), id)
			return generated.GetTicketByIDRow(row), nil
		},
	}

	handler := NewTicketHandler(mock, nil, &config.Config{JWTSecret: secret})
	router := setupTicketRouter(handler, secret)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/verify?id="+ticketID.String(), nil)
	token, _, _ := auth.GenerateToken(organiserID, "org@example.com", "ORGANISER", secret, 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool                       `json:"success"`
		Data    TicketVerificationResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.True(t, resp.Data.IsValid)
	assert.Equal(t, ticketID.String(), resp.Data.Ticket.ID)
}

func TestVerifyTicket_AlreadyCheckedIn(t *testing.T) {
	secret := "test-secret"
	customerID := uuid.New()
	organiserID := uuid.New()
	eventID := uuid.New()
	bookingID := uuid.New()
	ticketID := uuid.New()
	seatID := uuid.New()
	qrPayload := "TICKET|REF:TB-20260823-XYZ|SEAT:12|ID:" + ticketID.String()

	row := createSampleTicketRow(ticketID, bookingID, customerID, eventID, organiserID, seatID, qrPayload, generated.TicketStatusCHECKEDIN)
	checkInTime := time.Now().Add(-10 * time.Minute)
	row.CheckedInAt = pgtype.Timestamptz{Time: checkInTime, Valid: true}

	mock := &mockTicketQuerier{
		getTicketByQRPayloadFunc: func(ctx context.Context, payload string) (generated.GetTicketByQRPayloadRow, error) {
			return row, nil
		},
	}

	handler := NewTicketHandler(mock, nil, &config.Config{JWTSecret: secret})
	router := setupTicketRouter(handler, secret)

	body, _ := json.Marshal(VerifyTicketRequest{QRPayload: qrPayload})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/verify", bytes.NewReader(body))
	token, _, _ := auth.GenerateToken(organiserID, "org@example.com", "ORGANISER", secret, 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool                       `json:"success"`
		Data    TicketVerificationResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Data.IsValid)
	assert.False(t, resp.Data.CanCheckIn)
	assert.Contains(t, resp.Data.ValidationMessage, "already been checked in")
}

func TestVerifyTicket_ForbiddenCustomer(t *testing.T) {
	secret := "test-secret"
	customerID := uuid.New()
	otherCustomerID := uuid.New()
	organiserID := uuid.New()
	eventID := uuid.New()
	bookingID := uuid.New()
	ticketID := uuid.New()
	seatID := uuid.New()
	qrPayload := "TICKET|REF:TB-20260823-XYZ|SEAT:12|ID:" + ticketID.String()

	row := createSampleTicketRow(ticketID, bookingID, customerID, eventID, organiserID, seatID, qrPayload, generated.TicketStatusVALID)

	mock := &mockTicketQuerier{
		getTicketByQRPayloadFunc: func(ctx context.Context, payload string) (generated.GetTicketByQRPayloadRow, error) {
			return row, nil
		},
	}

	handler := NewTicketHandler(mock, nil, &config.Config{JWTSecret: secret})
	router := setupTicketRouter(handler, secret)

	body, _ := json.Marshal(VerifyTicketRequest{QRPayload: qrPayload})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/verify", bytes.NewReader(body))
	token, _, _ := auth.GenerateToken(otherCustomerID, "other@example.com", "CUSTOMER", secret, 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestCheckInTicket_Success(t *testing.T) {
	secret := "test-secret"
	customerID := uuid.New()
	organiserID := uuid.New()
	eventID := uuid.New()
	bookingID := uuid.New()
	ticketID := uuid.New()
	seatID := uuid.New()
	qrPayload := "TICKET|REF:TB-20260823-XYZ|SEAT:12|ID:" + ticketID.String()

	row := createSampleTicketRow(ticketID, bookingID, customerID, eventID, organiserID, seatID, qrPayload, generated.TicketStatusVALID)

	mock := &mockTicketQuerier{
		getTicketByQRPayloadFunc: func(ctx context.Context, payload string) (generated.GetTicketByQRPayloadRow, error) {
			return row, nil
		},
		checkInTicketByQRPayloadFunc: func(ctx context.Context, payload string) (generated.Ticket, error) {
			assert.Equal(t, qrPayload, payload)
			return generated.Ticket{
				ID:            utils.UUIDToPgtype(ticketID),
				BookingID:     utils.UUIDToPgtype(bookingID),
				EventID:       utils.UUIDToPgtype(eventID),
				SeatID:        utils.UUIDToPgtype(seatID),
				UnitPrice:     utils.Float64ToPgtypeNumeric(85.00),
				QrCodePayload: qrPayload,
				Status:        generated.TicketStatusCHECKEDIN,
				CreatedAt:     pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour), Valid: true},
				CheckedInAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	handler := NewTicketHandler(mock, nil, &config.Config{JWTSecret: secret})
	router := setupTicketRouter(handler, secret)

	body, _ := json.Marshal(CheckInTicketRequest{QRPayload: qrPayload})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/check-in", bytes.NewReader(body))
	token, _, _ := auth.GenerateToken(organiserID, "org@example.com", "ORGANISER", secret, 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool                       `json:"success"`
		Data    TicketVerificationResponse `json:"data"`
		Message string                     `json:"message"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "CHECKED_IN", resp.Data.Ticket.Status)
	assert.NotNil(t, resp.Data.Ticket.CheckedInAt)
	assert.Equal(t, "Ticket checked in successfully", resp.Message)
}

func TestCheckInTicket_ByID_Success(t *testing.T) {
	secret := "test-secret"
	customerID := uuid.New()
	organiserID := uuid.New()
	eventID := uuid.New()
	bookingID := uuid.New()
	ticketID := uuid.New()
	seatID := uuid.New()
	qrPayload := "TICKET|REF:TB-20260823-XYZ|SEAT:12|ID:" + ticketID.String()

	row := createSampleTicketRow(ticketID, bookingID, customerID, eventID, organiserID, seatID, qrPayload, generated.TicketStatusVALID)

	mock := &mockTicketQuerier{
		getTicketByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.GetTicketByIDRow, error) {
			return generated.GetTicketByIDRow(row), nil
		},
		checkInTicketByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.Ticket, error) {
			assert.Equal(t, utils.UUIDToPgtype(ticketID), id)
			return generated.Ticket{
				ID:            utils.UUIDToPgtype(ticketID),
				BookingID:     utils.UUIDToPgtype(bookingID),
				EventID:       utils.UUIDToPgtype(eventID),
				SeatID:        utils.UUIDToPgtype(seatID),
				UnitPrice:     utils.Float64ToPgtypeNumeric(85.00),
				QrCodePayload: qrPayload,
				Status:        generated.TicketStatusCHECKEDIN,
				CreatedAt:     pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour), Valid: true},
				CheckedInAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	handler := NewTicketHandler(mock, nil, &config.Config{JWTSecret: secret})
	router := setupTicketRouter(handler, secret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/"+ticketID.String()+"/check-in", nil)
	token, _, _ := auth.GenerateToken(organiserID, "org@example.com", "ORGANISER", secret, 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool                       `json:"success"`
		Data    TicketVerificationResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "CHECKED_IN", resp.Data.Ticket.Status)
}

func TestCheckInTicket_DuplicateReplay_Conflict(t *testing.T) {
	secret := "test-secret"
	customerID := uuid.New()
	organiserID := uuid.New()
	eventID := uuid.New()
	bookingID := uuid.New()
	ticketID := uuid.New()
	seatID := uuid.New()
	qrPayload := "TICKET|REF:TB-20260823-XYZ|SEAT:12|ID:" + ticketID.String()

	row := createSampleTicketRow(ticketID, bookingID, customerID, eventID, organiserID, seatID, qrPayload, generated.TicketStatusCHECKEDIN)
	row.CheckedInAt = pgtype.Timestamptz{Time: time.Now().Add(-5 * time.Minute), Valid: true}

	mock := &mockTicketQuerier{
		getTicketByQRPayloadFunc: func(ctx context.Context, payload string) (generated.GetTicketByQRPayloadRow, error) {
			return row, nil
		},
	}

	handler := NewTicketHandler(mock, nil, &config.Config{JWTSecret: secret})
	router := setupTicketRouter(handler, secret)

	body, _ := json.Marshal(CheckInTicketRequest{QRPayload: qrPayload})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/check-in", bytes.NewReader(body))
	token, _, _ := auth.GenerateToken(organiserID, "org@example.com", "ORGANISER", secret, 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "ALREADY_CHECKED_IN")
}

func TestCheckInTicket_CancelledTicket(t *testing.T) {
	secret := "test-secret"
	customerID := uuid.New()
	organiserID := uuid.New()
	eventID := uuid.New()
	bookingID := uuid.New()
	ticketID := uuid.New()
	seatID := uuid.New()
	qrPayload := "TICKET|REF:TB-20260823-XYZ|SEAT:12|ID:" + ticketID.String()

	row := createSampleTicketRow(ticketID, bookingID, customerID, eventID, organiserID, seatID, qrPayload, generated.TicketStatusCANCELLED)

	mock := &mockTicketQuerier{
		getTicketByQRPayloadFunc: func(ctx context.Context, payload string) (generated.GetTicketByQRPayloadRow, error) {
			return row, nil
		},
	}

	handler := NewTicketHandler(mock, nil, &config.Config{JWTSecret: secret})
	router := setupTicketRouter(handler, secret)

	body, _ := json.Marshal(CheckInTicketRequest{QRPayload: qrPayload})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/check-in", bytes.NewReader(body))
	token, _, _ := auth.GenerateToken(organiserID, "org@example.com", "ORGANISER", secret, 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "TICKET_CANCELLED")
}

func TestCheckInTicket_ForbiddenOtherOrganiser(t *testing.T) {
	secret := "test-secret"
	customerID := uuid.New()
	organiserID := uuid.New()
	otherOrganiserID := uuid.New()
	eventID := uuid.New()
	bookingID := uuid.New()
	ticketID := uuid.New()
	seatID := uuid.New()
	qrPayload := "TICKET|REF:TB-20260823-XYZ|SEAT:12|ID:" + ticketID.String()

	row := createSampleTicketRow(ticketID, bookingID, customerID, eventID, organiserID, seatID, qrPayload, generated.TicketStatusVALID)

	mock := &mockTicketQuerier{
		getTicketByQRPayloadFunc: func(ctx context.Context, payload string) (generated.GetTicketByQRPayloadRow, error) {
			return row, nil
		},
	}

	handler := NewTicketHandler(mock, nil, &config.Config{JWTSecret: secret})
	router := setupTicketRouter(handler, secret)

	body, _ := json.Marshal(CheckInTicketRequest{QRPayload: qrPayload})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets/check-in", bytes.NewReader(body))
	token, _, _ := auth.GenerateToken(otherOrganiserID, "otherorg@example.com", "ORGANISER", secret, 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "FORBIDDEN")
}

func TestListEventTickets_Success(t *testing.T) {
	secret := "test-secret"
	organiserID := uuid.New()
	eventID := uuid.New()
	ticketID := uuid.New()

	mock := &mockTicketQuerier{
		getEventByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error) {
			return generated.GetEventByIDRow{
				ID:          id,
				OrganiserID: utils.UUIDToPgtype(organiserID),
				Title:       "Concert Event",
			}, nil
		},
		getEventCheckInStatsFunc: func(ctx context.Context, id pgtype.UUID) (generated.GetEventCheckInStatsRow, error) {
			return generated.GetEventCheckInStatsRow{
				TotalTickets:   10,
				ValidCount:     8,
				CheckedInCount: 2,
				CancelledCount: 0,
			}, nil
		},
		listTicketsByEventIDFunc: func(ctx context.Context, id pgtype.UUID) ([]generated.ListTicketsByEventIDRow, error) {
			return []generated.ListTicketsByEventIDRow{
				{
					ID:               utils.UUIDToPgtype(ticketID),
					BookingID:        utils.UUIDToPgtype(uuid.New()),
					EventID:          id,
					SeatID:           utils.UUIDToPgtype(uuid.New()),
					UnitPrice:        utils.Float64ToPgtypeNumeric(50.00),
					QrCodePayload:    "TICKET|PAYLOAD",
					Status:           generated.TicketStatusCHECKEDIN,
					CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
					CheckedInAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
					BookingReference: "TB-2026-AAA",
					CustomerID:       utils.UUIDToPgtype(uuid.New()),
					CustomerName:     "Alice Smith",
					CustomerEmail:    "alice@example.com",
					RowLabel:         "B",
					SeatNumber:       "3",
					CategoryName:     "Standard",
					CategoryColor:    "#3B82F6",
				},
			}, nil
		},
	}

	handler := NewTicketHandler(mock, nil, &config.Config{JWTSecret: secret})
	router := setupTicketRouter(handler, secret)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organiser/events/"+eventID.String()+"/tickets", nil)
	token, _, _ := auth.GenerateToken(organiserID, "org@example.com", "ORGANISER", secret, 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool                         `json:"success"`
		Data    EventCheckInOverviewResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, int64(10), resp.Data.TotalTickets)
	assert.Equal(t, int64(2), resp.Data.CheckedInCount)
	assert.Equal(t, 20.0, resp.Data.CheckInRate)
	assert.Len(t, resp.Data.Tickets, 1)
}
