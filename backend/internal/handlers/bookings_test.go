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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ticket-booking/backend/internal/auth"
	"ticket-booking/backend/internal/config"
	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/middleware"
	"ticket-booking/backend/internal/utils"
)

type mockBookingQuerier struct {
	generated.Querier
	getActiveHoldByTokenFunc func(ctx context.Context, arg generated.GetActiveHoldByTokenParams) ([]generated.GetActiveHoldByTokenRow, error)
	createBookingFunc        func(ctx context.Context, arg generated.CreateBookingParams) (generated.Booking, error)
	createTicketFunc         func(ctx context.Context, arg generated.CreateTicketParams) (generated.Ticket, error)
	getCustomerBookingsFunc  func(ctx context.Context, customerID pgtype.UUID) ([]generated.GetCustomerBookingsRow, error)
	getBookingByIDFunc       func(ctx context.Context, id pgtype.UUID) (generated.GetBookingByIDRow, error)
	getTicketsByBookingIDFunc func(ctx context.Context, bookingID pgtype.UUID) ([]generated.GetTicketsByBookingIDRow, error)
	cancelBookingFunc        func(ctx context.Context, arg generated.CancelBookingParams) (generated.Booking, error)
	cancelBookingTicketsFunc func(ctx context.Context, bookingID pgtype.UUID) ([]generated.Ticket, error)
	cancelBookingReservationsFunc func(ctx context.Context, bookingID pgtype.UUID) ([]generated.SeatReservation, error)
}

func (m *mockBookingQuerier) GetActiveHoldByToken(ctx context.Context, arg generated.GetActiveHoldByTokenParams) ([]generated.GetActiveHoldByTokenRow, error) {
	if m.getActiveHoldByTokenFunc != nil {
		return m.getActiveHoldByTokenFunc(ctx, arg)
	}
	return nil, nil
}

func (m *mockBookingQuerier) CreateBooking(ctx context.Context, arg generated.CreateBookingParams) (generated.Booking, error) {
	if m.createBookingFunc != nil {
		return m.createBookingFunc(ctx, arg)
	}
	return generated.Booking{
		ID:               utils.UUIDToPgtype(uuid.New()),
		BookingReference: arg.BookingReference,
		CustomerID:       arg.CustomerID,
		EventID:          arg.EventID,
		TotalAmount:      arg.TotalAmount,
		Currency:         arg.Currency,
		Status:           generated.BookingStatusCONFIRMED,
		CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *mockBookingQuerier) CreateTicket(ctx context.Context, arg generated.CreateTicketParams) (generated.Ticket, error) {
	if m.createTicketFunc != nil {
		return m.createTicketFunc(ctx, arg)
	}
	return generated.Ticket{
		ID:            utils.UUIDToPgtype(uuid.New()),
		BookingID:     arg.BookingID,
		EventID:       arg.EventID,
		SeatID:        arg.SeatID,
		UnitPrice:     arg.UnitPrice,
		QrCodePayload: arg.QrCodePayload,
		Status:        generated.TicketStatusVALID,
		CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *mockBookingQuerier) GetCustomerBookings(ctx context.Context, customerID pgtype.UUID) ([]generated.GetCustomerBookingsRow, error) {
	if m.getCustomerBookingsFunc != nil {
		return m.getCustomerBookingsFunc(ctx, customerID)
	}
	return nil, nil
}

func (m *mockBookingQuerier) GetBookingByID(ctx context.Context, id pgtype.UUID) (generated.GetBookingByIDRow, error) {
	if m.getBookingByIDFunc != nil {
		return m.getBookingByIDFunc(ctx, id)
	}
	return generated.GetBookingByIDRow{}, nil
}

func (m *mockBookingQuerier) GetTicketsByBookingID(ctx context.Context, bookingID pgtype.UUID) ([]generated.GetTicketsByBookingIDRow, error) {
	if m.getTicketsByBookingIDFunc != nil {
		return m.getTicketsByBookingIDFunc(ctx, bookingID)
	}
	return nil, nil
}

func (m *mockBookingQuerier) CancelBooking(ctx context.Context, arg generated.CancelBookingParams) (generated.Booking, error) {
	if m.cancelBookingFunc != nil {
		return m.cancelBookingFunc(ctx, arg)
	}
	return generated.Booking{
		ID:               arg.ID,
		BookingReference: "TB-2026-CANCELLED",
		Status:           generated.BookingStatusCANCELLED,
		CancelledAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *mockBookingQuerier) CancelBookingTickets(ctx context.Context, bookingID pgtype.UUID) ([]generated.Ticket, error) {
	if m.cancelBookingTicketsFunc != nil {
		return m.cancelBookingTicketsFunc(ctx, bookingID)
	}
	return nil, nil
}

func (m *mockBookingQuerier) CancelBookingReservations(ctx context.Context, bookingID pgtype.UUID) ([]generated.SeatReservation, error) {
	if m.cancelBookingReservationsFunc != nil {
		return m.cancelBookingReservationsFunc(ctx, bookingID)
	}
	return nil, nil
}

func (m *mockBookingQuerier) ConfirmReservationToBooked(ctx context.Context, arg generated.ConfirmReservationToBookedParams) (int64, error) {
	return 1, nil
}

func TestCheckout_Success(t *testing.T) {
	customerID := uuid.New()
	eventID := uuid.New()
	holdToken := uuid.New()
	seatID := uuid.New()
	catID := uuid.New()

	mock := &mockBookingQuerier{
		getActiveHoldByTokenFunc: func(ctx context.Context, arg generated.GetActiveHoldByTokenParams) ([]generated.GetActiveHoldByTokenRow, error) {
			return []generated.GetActiveHoldByTokenRow{
				{
					ID:             utils.UUIDToPgtype(uuid.New()),
					EventID:        arg.EventID,
					SeatID:         utils.UUIDToPgtype(seatID),
					UserID:         utils.UUIDToPgtype(customerID),
					Status:         generated.ReservationStatusHELD,
					RowLabel:       "B",
					SeatNumber:     "5",
					SeatCategoryID: utils.UUIDToPgtype(catID),
					Price:          utils.Float64ToPgtypeNumeric(75.00),
					Currency:       "USD",
					CategoryName:   "Premium",
					CategoryColor:  "#3B82F6",
				},
			}, nil
		},
	}

	handler := NewBookingHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)

	body := CheckoutRequest{
		EventID:   eventID.String(),
		HoldToken: holdToken.String(),
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/checkout", bytes.NewReader(jsonBody))
	token, _, _ := auth.GenerateToken(customerID, "cust@example.com", "CUSTOMER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/bookings/checkout", handler.Checkout)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp struct {
		Success bool            `json:"success"`
		Data    BookingResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "CONFIRMED", resp.Data.Status)
	assert.Equal(t, 75.00, resp.Data.TotalAmount)
	assert.Len(t, resp.Data.Tickets, 1)
	assert.NotEmpty(t, resp.Data.Tickets[0].QRCodePayload)
	assert.NotEmpty(t, resp.Data.Tickets[0].QRCodeDataURL)
}

func TestCheckout_ExpiredHold(t *testing.T) {
	customerID := uuid.New()
	eventID := uuid.New()
	holdToken := uuid.New()

	mock := &mockBookingQuerier{
		getActiveHoldByTokenFunc: func(ctx context.Context, arg generated.GetActiveHoldByTokenParams) ([]generated.GetActiveHoldByTokenRow, error) {
			return nil, nil // No active holds!
		},
	}

	handler := NewBookingHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)

	body := CheckoutRequest{
		EventID:   eventID.String(),
		HoldToken: holdToken.String(),
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings/checkout", bytes.NewReader(jsonBody))
	token, _, _ := auth.GenerateToken(customerID, "cust@example.com", "CUSTOMER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/bookings/checkout", handler.Checkout)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "HOLD_EXPIRED")
}

func TestListCustomerBookings_Success(t *testing.T) {
	customerID := uuid.New()
	bookingID := uuid.New()

	mock := &mockBookingQuerier{
		getCustomerBookingsFunc: func(ctx context.Context, cID pgtype.UUID) ([]generated.GetCustomerBookingsRow, error) {
			return []generated.GetCustomerBookingsRow{
				{
					ID:               utils.UUIDToPgtype(bookingID),
					BookingReference: "TB-20260823-112233",
					CustomerID:       cID,
					EventTitle:       "Avengers Premiere",
					VenueName:        "IMAX Mumbai",
					TotalAmount:      utils.Float64ToPgtypeNumeric(150.00),
					Currency:         "USD",
					Status:           generated.BookingStatusCONFIRMED,
					TicketCount:      2,
					CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
				},
			}, nil
		},
	}

	handler := NewBookingHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/bookings", nil)
	token, _, _ := auth.GenerateToken(customerID, "cust@example.com", "CUSTOMER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Get("/api/v1/customer/bookings", handler.ListCustomerBookings)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool              `json:"success"`
		Data    []BookingResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "TB-20260823-112233", resp.Data[0].BookingReference)
}

func TestCancelBooking_Success(t *testing.T) {
	customerID := uuid.New()
	bookingID := uuid.New()

	mock := &mockBookingQuerier{
		getBookingByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.GetBookingByIDRow, error) {
			return generated.GetBookingByIDRow{
				ID:             id,
				CustomerID:     utils.UUIDToPgtype(customerID),
				Status:         generated.BookingStatusCONFIRMED,
				EventStartTime: pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
			}, nil
		},
		getTicketsByBookingIDFunc: func(ctx context.Context, bookingID pgtype.UUID) ([]generated.GetTicketsByBookingIDRow, error) {
			return []generated.GetTicketsByBookingIDRow{
				{
					ID:     utils.UUIDToPgtype(uuid.New()),
					Status: generated.TicketStatusVALID,
				},
			}, nil
		},
	}
	handler := NewBookingHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/customer/bookings/"+bookingID.String()+"/cancel", nil)
	token, _, _ := auth.GenerateToken(customerID, "cust@example.com", "CUSTOMER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/customer/bookings/{id}/cancel", handler.CancelBooking)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCancelBooking_EventAlreadyStarted(t *testing.T) {
	customerID := uuid.New()
	bookingID := uuid.New()

	mock := &mockBookingQuerier{
		getBookingByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.GetBookingByIDRow, error) {
			return generated.GetBookingByIDRow{
				ID:             id,
				CustomerID:     utils.UUIDToPgtype(customerID),
				Status:         generated.BookingStatusCONFIRMED,
				EventStartTime: pgtype.Timestamptz{Time: time.Now().Add(-2 * time.Hour), Valid: true},
			}, nil
		},
	}
	handler := NewBookingHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/customer/bookings/"+bookingID.String()+"/cancel", nil)
	token, _, _ := auth.GenerateToken(customerID, "cust@example.com", "CUSTOMER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/customer/bookings/{id}/cancel", handler.CancelBooking)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "EVENT_ALREADY_STARTED")
}

func TestCancelBooking_TicketAlreadyCheckedIn(t *testing.T) {
	customerID := uuid.New()
	bookingID := uuid.New()

	mock := &mockBookingQuerier{
		getBookingByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.GetBookingByIDRow, error) {
			return generated.GetBookingByIDRow{
				ID:             id,
				CustomerID:     utils.UUIDToPgtype(customerID),
				Status:         generated.BookingStatusCONFIRMED,
				EventStartTime: pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
			}, nil
		},
		getTicketsByBookingIDFunc: func(ctx context.Context, bookingID pgtype.UUID) ([]generated.GetTicketsByBookingIDRow, error) {
			return []generated.GetTicketsByBookingIDRow{
				{
					ID:          utils.UUIDToPgtype(uuid.New()),
					Status:      generated.TicketStatusCHECKEDIN,
					CheckedInAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				},
			}, nil
		},
	}
	handler := NewBookingHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/customer/bookings/"+bookingID.String()+"/cancel", nil)
	token, _, _ := auth.GenerateToken(customerID, "cust@example.com", "CUSTOMER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/customer/bookings/{id}/cancel", handler.CancelBooking)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "TICKET_ALREADY_CHECKED_IN")
}
