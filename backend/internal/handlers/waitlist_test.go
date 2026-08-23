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
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ticket-booking/backend/internal/auth"
	"ticket-booking/backend/internal/config"
	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/middleware"
	"ticket-booking/backend/internal/utils"
)

type mockWaitlistQuerier struct {
	generated.Querier
	getEventByIDFunc                  func(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error)
	countAvailableSeatsInCategoryFunc func(ctx context.Context, arg generated.CountAvailableSeatsInCategoryParams) (int64, error)
	joinWaitlistFunc                  func(ctx context.Context, arg generated.JoinWaitlistParams) (generated.WaitlistEntry, error)
	getWaitlistQueuePositionFunc      func(ctx context.Context, arg generated.GetWaitlistQueuePositionParams) (int32, error)
	getWaitlistOfferByTokenFunc       func(ctx context.Context, token pgtype.UUID) (generated.GetWaitlistOfferByTokenRow, error)
	getActiveHoldOrOfferByTokenFunc   func(ctx context.Context, arg generated.GetActiveHoldOrOfferByTokenParams) ([]generated.GetActiveHoldOrOfferByTokenRow, error)
	createBookingFunc                 func(ctx context.Context, arg generated.CreateBookingParams) (generated.Booking, error)
	createTicketFunc                  func(ctx context.Context, arg generated.CreateTicketParams) (generated.Ticket, error)
	confirmReservationToBookedFunc    func(ctx context.Context, arg generated.ConfirmReservationToBookedParams) (int64, error)
	updateWaitlistEntryStatusFunc     func(ctx context.Context, arg generated.UpdateWaitlistEntryStatusParams) (generated.WaitlistEntry, error)
}

func (m *mockWaitlistQuerier) GetEventByID(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error) {
	if m.getEventByIDFunc != nil {
		return m.getEventByIDFunc(ctx, id)
	}
	return generated.GetEventByIDRow{}, nil
}

func (m *mockWaitlistQuerier) CountAvailableSeatsInCategory(ctx context.Context, arg generated.CountAvailableSeatsInCategoryParams) (int64, error) {
	if m.countAvailableSeatsInCategoryFunc != nil {
		return m.countAvailableSeatsInCategoryFunc(ctx, arg)
	}
	return 0, nil
}

func (m *mockWaitlistQuerier) JoinWaitlist(ctx context.Context, arg generated.JoinWaitlistParams) (generated.WaitlistEntry, error) {
	if m.joinWaitlistFunc != nil {
		return m.joinWaitlistFunc(ctx, arg)
	}
	return generated.WaitlistEntry{
		ID:             utils.UUIDToPgtype(uuid.New()),
		EventID:        arg.EventID,
		SeatCategoryID: arg.SeatCategoryID,
		CustomerID:     arg.CustomerID,
		Status:         generated.WaitlistStatusWAITING,
		CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *mockWaitlistQuerier) GetWaitlistQueuePosition(ctx context.Context, arg generated.GetWaitlistQueuePositionParams) (int32, error) {
	if m.getWaitlistQueuePositionFunc != nil {
		return m.getWaitlistQueuePositionFunc(ctx, arg)
	}
	return 1, nil
}

func (m *mockWaitlistQuerier) GetWaitlistOfferByToken(ctx context.Context, token pgtype.UUID) (generated.GetWaitlistOfferByTokenRow, error) {
	if m.getWaitlistOfferByTokenFunc != nil {
		return m.getWaitlistOfferByTokenFunc(ctx, token)
	}
	return generated.GetWaitlistOfferByTokenRow{}, nil
}

func (m *mockWaitlistQuerier) GetActiveHoldOrOfferByToken(ctx context.Context, arg generated.GetActiveHoldOrOfferByTokenParams) ([]generated.GetActiveHoldOrOfferByTokenRow, error) {
	if m.getActiveHoldOrOfferByTokenFunc != nil {
		return m.getActiveHoldOrOfferByTokenFunc(ctx, arg)
	}
	return nil, nil
}

func (m *mockWaitlistQuerier) CreateBooking(ctx context.Context, arg generated.CreateBookingParams) (generated.Booking, error) {
	if m.createBookingFunc != nil {
		return m.createBookingFunc(ctx, arg)
	}
	return generated.Booking{
		ID:               utils.UUIDToPgtype(uuid.New()),
		BookingReference: arg.BookingReference,
		CustomerID:       arg.CustomerID,
		EventID:          arg.EventID,
		Status:           generated.BookingStatusCONFIRMED,
		CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

func (m *mockWaitlistQuerier) CreateTicket(ctx context.Context, arg generated.CreateTicketParams) (generated.Ticket, error) {
	if m.createTicketFunc != nil {
		return m.createTicketFunc(ctx, arg)
	}
	return generated.Ticket{
		ID:            utils.UUIDToPgtype(uuid.New()),
		BookingID:     arg.BookingID,
		EventID:       arg.EventID,
		SeatID:        arg.SeatID,
		QrCodePayload: arg.QrCodePayload,
		Status:        generated.TicketStatusVALID,
	}, nil
}

func (m *mockWaitlistQuerier) ConfirmReservationToBooked(ctx context.Context, arg generated.ConfirmReservationToBookedParams) (int64, error) {
	if m.confirmReservationToBookedFunc != nil {
		return m.confirmReservationToBookedFunc(ctx, arg)
	}
	return 1, nil
}

func (m *mockWaitlistQuerier) UpdateWaitlistEntryStatus(ctx context.Context, arg generated.UpdateWaitlistEntryStatusParams) (generated.WaitlistEntry, error) {
	if m.updateWaitlistEntryStatusFunc != nil {
		return m.updateWaitlistEntryStatusFunc(ctx, arg)
	}
	return generated.WaitlistEntry{ID: arg.ID, Status: arg.Status}, nil
}

// ============================================================================

func authenticatedWaitlistRequest(t *testing.T, userID uuid.UUID, method, path string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	token, _, _ := auth.GenerateToken(userID, "cust@example.com", "CUSTOMER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	return req, httptest.NewRecorder()
}

func waitlistTestRouter(handler *WaitlistHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/events/{id}/waitlist", handler.JoinWaitlist)
	r.Post("/api/v1/waitlist/offers/{token}/accept", handler.AcceptOffer)
	return r
}

func TestJoinWaitlist_Success(t *testing.T) {
	eventID := uuid.New()
	categoryID := uuid.New()

	mock := &mockWaitlistQuerier{
		getEventByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error) {
			return generated.GetEventByIDRow{Status: generated.EventStatusPUBLISHED}, nil
		},
		countAvailableSeatsInCategoryFunc: func(ctx context.Context, arg generated.CountAvailableSeatsInCategoryParams) (int64, error) {
			return 0, nil // sold out
		},
		joinWaitlistFunc: func(ctx context.Context, arg generated.JoinWaitlistParams) (generated.WaitlistEntry, error) {
			return generated.WaitlistEntry{
				ID:             utils.UUIDToPgtype(uuid.New()),
				EventID:        arg.EventID,
				SeatCategoryID: arg.SeatCategoryID,
				CustomerID:     arg.CustomerID,
				Status:         generated.WaitlistStatusWAITING,
				CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
		getWaitlistQueuePositionFunc: func(ctx context.Context, arg generated.GetWaitlistQueuePositionParams) (int32, error) {
			return 4, nil // fourth in line
		},
	}

	handler := NewWaitlistHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)
	customerID := uuid.New()
	req, rr := authenticatedWaitlistRequest(t, customerID, http.MethodPost,
		"/api/v1/events/"+eventID.String()+"/waitlist",
		JoinWaitlistRequest{SeatCategoryID: categoryID.String()})

	waitlistTestRouter(handler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	var resp struct {
		Success bool                  `json:"success"`
		Data    WaitlistEntryResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "WAITING", resp.Data.Status)
	assert.Equal(t, 4, resp.Data.QueuePosition)
}

func TestJoinWaitlist_CategoryNotSoldOut(t *testing.T) {
	eventID := uuid.New()
	categoryID := uuid.New()

	mock := &mockWaitlistQuerier{
		getEventByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error) {
			return generated.GetEventByIDRow{Status: generated.EventStatusPUBLISHED}, nil
		},
		countAvailableSeatsInCategoryFunc: func(ctx context.Context, arg generated.CountAvailableSeatsInCategoryParams) (int64, error) {
			return 12, nil // seats still available
		},
	}

	handler := NewWaitlistHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)
	customerID := uuid.New()
	req, rr := authenticatedWaitlistRequest(t, customerID, http.MethodPost,
		"/api/v1/events/"+eventID.String()+"/waitlist",
		JoinWaitlistRequest{SeatCategoryID: categoryID.String()})

	waitlistTestRouter(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "CATEGORY_AVAILABLE")
}

func TestJoinWaitlist_AlreadyOnWaitlist(t *testing.T) {
	eventID := uuid.New()
	categoryID := uuid.New()

	mock := &mockWaitlistQuerier{
		getEventByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error) {
			return generated.GetEventByIDRow{Status: generated.EventStatusPUBLISHED}, nil
		},
		countAvailableSeatsInCategoryFunc: func(ctx context.Context, arg generated.CountAvailableSeatsInCategoryParams) (int64, error) {
			return 0, nil
		},
		joinWaitlistFunc: func(ctx context.Context, arg generated.JoinWaitlistParams) (generated.WaitlistEntry, error) {
			return generated.WaitlistEntry{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation}
		},
	}

	handler := NewWaitlistHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)
	customerID := uuid.New()
	req, rr := authenticatedWaitlistRequest(t, customerID, http.MethodPost,
		"/api/v1/events/"+eventID.String()+"/waitlist",
		JoinWaitlistRequest{SeatCategoryID: categoryID.String()})

	waitlistTestRouter(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "ALREADY_ON_WAITLIST")
}

func TestJoinWaitlist_Unauthenticated(t *testing.T) {
	handler := NewWaitlistHandler(&mockWaitlistQuerier{}, nil, &config.Config{JWTSecret: "test-secret"}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+uuid.New().String()+"/waitlist", bytes.NewReader([]byte(`{}`)))
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/api/v1/events/{id}/waitlist", handler.JoinWaitlist)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func sampleOfferRow(customerID uuid.UUID, status generated.OfferStatus, expiresAt time.Time) generated.GetWaitlistOfferByTokenRow {
	return generated.GetWaitlistOfferByTokenRow{
		ID:              utils.UUIDToPgtype(uuid.New()),
		WaitlistEntryID: utils.UUIDToPgtype(uuid.New()),
		EventID:         utils.UUIDToPgtype(uuid.New()),
		SeatID:          utils.UUIDToPgtype(uuid.New()),
		OfferToken:      utils.UUIDToPgtype(uuid.New()),
		ExpiresAt:       pgtype.Timestamptz{Time: expiresAt, Valid: true},
		Status:          status,
		CustomerID:      utils.UUIDToPgtype(customerID),
		CustomerEmail:   "cust@example.com",
		CustomerName:    "Test Customer",
		EventTitle:      "Test Event",
		CategoryName:    "Premium",
	}
}

func TestAcceptOffer_Success(t *testing.T) {
	customerID := uuid.New()
	offer := sampleOfferRow(customerID, generated.OfferStatusPENDING, time.Now().Add(5*time.Minute))
	seatID := uuid.New()

	mock := &mockWaitlistQuerier{
		getWaitlistOfferByTokenFunc: func(ctx context.Context, token pgtype.UUID) (generated.GetWaitlistOfferByTokenRow, error) {
			assert.Equal(t, offer.OfferToken, token)
			return offer, nil
		},
		getActiveHoldOrOfferByTokenFunc: func(ctx context.Context, arg generated.GetActiveHoldOrOfferByTokenParams) ([]generated.GetActiveHoldOrOfferByTokenRow, error) {
			return []generated.GetActiveHoldOrOfferByTokenRow{{
				UserID:       utils.UUIDToPgtype(customerID),
				SeatID:       utils.UUIDToPgtype(seatID),
				Status:       generated.ReservationStatusOFFERED,
				Price:        utils.Float64ToPgtypeNumeric(150.00),
				Currency:     "USD",
				RowLabel:     "A",
				SeatNumber:   "12",
				CategoryName: "Premium",
			}}, nil
		},
		confirmReservationToBookedFunc: func(ctx context.Context, arg generated.ConfirmReservationToBookedParams) (int64, error) {
			assert.Equal(t, offer.EventID, arg.EventID)
			assert.Equal(t, offer.OfferToken, arg.HoldToken)
			return 1, nil
		},
		updateWaitlistEntryStatusFunc: func(ctx context.Context, arg generated.UpdateWaitlistEntryStatusParams) (generated.WaitlistEntry, error) {
			assert.Equal(t, generated.WaitlistStatusACCEPTED, arg.Status)
			return generated.WaitlistEntry{ID: arg.ID, Status: arg.Status}, nil
		},
	}

	handler := NewWaitlistHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)
	req, rr := authenticatedWaitlistRequest(t, utils.PgtypeToUUID(offer.CustomerID), http.MethodPost,
		"/api/v1/waitlist/offers/"+utils.PgtypeToUUID(offer.OfferToken).String()+"/accept", nil)

	waitlistTestRouter(handler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())

	var resp struct {
		Success bool            `json:"success"`
		Data    BookingResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "CONFIRMED", resp.Data.Status)
	assert.Equal(t, 1, resp.Data.TicketCount)
	assert.Equal(t, 150.00, resp.Data.TotalAmount)
}

func TestAcceptOffer_ForbiddenForOtherUser(t *testing.T) {
	offer := sampleOfferRow(uuid.New(), generated.OfferStatusPENDING, time.Now().Add(5*time.Minute))

	mock := &mockWaitlistQuerier{
		getWaitlistOfferByTokenFunc: func(ctx context.Context, token pgtype.UUID) (generated.GetWaitlistOfferByTokenRow, error) {
			return offer, nil
		},
	}

	handler := NewWaitlistHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)
	// Authenticate as a customer who is NOT the offer owner.
	req, rr := authenticatedWaitlistRequest(t, uuid.New(), http.MethodPost,
		"/api/v1/waitlist/offers/"+utils.PgtypeToUUID(offer.OfferToken).String()+"/accept", nil)

	waitlistTestRouter(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "FORBIDDEN")
}

func TestAcceptOffer_Expired(t *testing.T) {
	offer := sampleOfferRow(uuid.New(), generated.OfferStatusPENDING, time.Now().Add(-time.Minute))

	mock := &mockWaitlistQuerier{
		getWaitlistOfferByTokenFunc: func(ctx context.Context, token pgtype.UUID) (generated.GetWaitlistOfferByTokenRow, error) {
			return offer, nil
		},
	}

	handler := NewWaitlistHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)
	req, rr := authenticatedWaitlistRequest(t, utils.PgtypeToUUID(offer.CustomerID), http.MethodPost,
		"/api/v1/waitlist/offers/"+utils.PgtypeToUUID(offer.OfferToken).String()+"/accept", nil)

	waitlistTestRouter(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusGone, rr.Code)
	assert.Contains(t, rr.Body.String(), "OFFER_EXPIRED")
}

func TestAcceptOffer_AlreadyAccepted(t *testing.T) {
	offer := sampleOfferRow(uuid.New(), generated.OfferStatusACCEPTED, time.Now().Add(5*time.Minute))

	mock := &mockWaitlistQuerier{
		getWaitlistOfferByTokenFunc: func(ctx context.Context, token pgtype.UUID) (generated.GetWaitlistOfferByTokenRow, error) {
			return offer, nil
		},
	}

	handler := NewWaitlistHandler(mock, nil, &config.Config{JWTSecret: "test-secret"}, nil)
	req, rr := authenticatedWaitlistRequest(t, utils.PgtypeToUUID(offer.CustomerID), http.MethodPost,
		"/api/v1/waitlist/offers/"+utils.PgtypeToUUID(offer.OfferToken).String()+"/accept", nil)

	waitlistTestRouter(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "OFFER_NOT_PENDING")
}
