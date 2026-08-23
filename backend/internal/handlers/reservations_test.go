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

type mockReservationQuerier struct {
	generated.Querier
	getEventSeatMapWithStatusFunc    func(ctx context.Context, id pgtype.UUID) ([]generated.GetEventSeatMapWithStatusRow, error)
	getEventByIDFunc                 func(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error)
	checkSeatAvailableFunc           func(ctx context.Context, arg generated.CheckSeatAvailableParams) (generated.CheckSeatAvailableRow, error)
	createSeatReservationFunc        func(ctx context.Context, arg generated.CreateSeatReservationParams) (generated.SeatReservation, error)
	releaseSeatHoldByTokenFunc       func(ctx context.Context, arg generated.ReleaseSeatHoldByTokenParams) (int64, error)
	releaseSeatHoldBySeatAndUserFunc func(ctx context.Context, arg generated.ReleaseSeatHoldBySeatAndUserParams) (int64, error)
	releaseExpiredSeatHoldFunc       func(ctx context.Context, arg generated.ReleaseExpiredSeatHoldParams) (int64, error)
}

func (m *mockReservationQuerier) ReleaseExpiredSeatHold(ctx context.Context, arg generated.ReleaseExpiredSeatHoldParams) (int64, error) {
	if m.releaseExpiredSeatHoldFunc != nil {
		return m.releaseExpiredSeatHoldFunc(ctx, arg)
	}
	return 0, nil
}

func (m *mockReservationQuerier) GetEventSeatMapWithStatus(ctx context.Context, id pgtype.UUID) ([]generated.GetEventSeatMapWithStatusRow, error) {
	if m.getEventSeatMapWithStatusFunc != nil {
		return m.getEventSeatMapWithStatusFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockReservationQuerier) GetEventByID(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error) {
	if m.getEventByIDFunc != nil {
		return m.getEventByIDFunc(ctx, id)
	}
	return generated.GetEventByIDRow{
		ID:             id,
		Status:         generated.EventStatusPUBLISHED,
		StartTime:      pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
		HoldTtlSeconds: 600,
	}, nil
}

func (m *mockReservationQuerier) CheckSeatAvailable(ctx context.Context, arg generated.CheckSeatAvailableParams) (generated.CheckSeatAvailableRow, error) {
	if m.checkSeatAvailableFunc != nil {
		return m.checkSeatAvailableFunc(ctx, arg)
	}
	return generated.CheckSeatAvailableRow{
		SeatID:         arg.ID_2,
		SeatCategoryID: utils.UUIDToPgtype(uuid.New()),
		Price:          utils.Float64ToPgtypeNumeric(50.0),
		Currency:       "USD",
	}, nil
}

func (m *mockReservationQuerier) CreateSeatReservation(ctx context.Context, arg generated.CreateSeatReservationParams) (generated.SeatReservation, error) {
	if m.createSeatReservationFunc != nil {
		return m.createSeatReservationFunc(ctx, arg)
	}
	return generated.SeatReservation{
		ID:        utils.UUIDToPgtype(uuid.New()),
		EventID:   arg.EventID,
		SeatID:    arg.SeatID,
		UserID:    arg.UserID,
		Status:    arg.Status,
		HoldToken: arg.HoldToken,
		ExpiresAt: arg.ExpiresAt,
	}, nil
}

func (m *mockReservationQuerier) ReleaseSeatHoldByToken(ctx context.Context, arg generated.ReleaseSeatHoldByTokenParams) (int64, error) {
	if m.releaseSeatHoldByTokenFunc != nil {
		return m.releaseSeatHoldByTokenFunc(ctx, arg)
	}
	return 1, nil
}

func (m *mockReservationQuerier) ReleaseSeatHoldBySeatAndUser(ctx context.Context, arg generated.ReleaseSeatHoldBySeatAndUserParams) (int64, error) {
	if m.releaseSeatHoldBySeatAndUserFunc != nil {
		return m.releaseSeatHoldBySeatAndUserFunc(ctx, arg)
	}
	return 1, nil
}

func TestGetEventSeatMap_Success(t *testing.T) {
	eventID := uuid.New()
	seatID1 := uuid.New()
	seatID2 := uuid.New()

	mock := &mockReservationQuerier{
		getEventSeatMapWithStatusFunc: func(ctx context.Context, id pgtype.UUID) ([]generated.GetEventSeatMapWithStatusRow, error) {
			return []generated.GetEventSeatMapWithStatusRow{
				{
					SeatID:         utils.UUIDToPgtype(seatID1),
					RowLabel:       "A",
					SeatNumber:     "1",
					GridRow:        1,
					GridCol:        1,
					Price:          utils.Float64ToPgtypeNumeric(100.0),
					Currency:       "USD",
					CategoryName:   "VIP",
					CategoryColor:  "#F59E0B",
					ComputedStatus: "AVAILABLE",
				},
				{
					SeatID:         utils.UUIDToPgtype(seatID2),
					RowLabel:       "A",
					SeatNumber:     "2",
					GridRow:        1,
					GridCol:        2,
					Price:          utils.Float64ToPgtypeNumeric(100.0),
					Currency:       "USD",
					CategoryName:   "VIP",
					CategoryColor:  "#F59E0B",
					ComputedStatus: "HELD",
				},
			}, nil
		},
	}

	handler := NewReservationHandler(mock, nil, &config.Config{JWTSecret: "test-secret"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/"+eventID.String()+"/seats", nil)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/api/v1/events/{id}/seats", handler.GetEventSeatMap)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool                  `json:"success"`
		Data    []SeatMapItemResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "AVAILABLE", resp.Data[0].Status)
	assert.Equal(t, "HELD", resp.Data[1].Status)
}

func TestHoldSeats_Success(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()
	seatID1 := uuid.New()
	seatID2 := uuid.New()

	mock := &mockReservationQuerier{}
	handler := NewReservationHandler(mock, nil, &config.Config{JWTSecret: "test-secret"})

	body := HoldSeatsRequest{
		SeatIDs: []string{seatID1.String(), seatID2.String()},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID.String()+"/hold", bytes.NewReader(jsonBody))
	token, _, _ := auth.GenerateToken(userID, "user@example.com", "CUSTOMER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/events/{id}/hold", handler.HoldSeats)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp struct {
		Success bool              `json:"success"`
		Data    HoldSeatsResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.Data.HoldToken)
	assert.Equal(t, 2, resp.Data.SeatCount)
	assert.Equal(t, 100.0, resp.Data.TotalPrice)
}

func TestHoldSeats_SeatUnavailable(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()
	seatID := uuid.New()

	mock := &mockReservationQuerier{
		checkSeatAvailableFunc: func(ctx context.Context, arg generated.CheckSeatAvailableParams) (generated.CheckSeatAvailableRow, error) {
			return generated.CheckSeatAvailableRow{}, errorsNew("seat booked or held")
		},
	}
	handler := NewReservationHandler(mock, nil, &config.Config{JWTSecret: "test-secret"})

	body := HoldSeatsRequest{
		SeatIDs: []string{seatID.String()},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID.String()+"/hold", bytes.NewReader(jsonBody))
	token, _, _ := auth.GenerateToken(userID, "user@example.com", "CUSTOMER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/events/{id}/hold", handler.HoldSeats)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "SEAT_UNAVAILABLE")
}

func TestReleaseHold_Success(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()
	holdToken := uuid.New()

	mock := &mockReservationQuerier{}
	handler := NewReservationHandler(mock, nil, &config.Config{JWTSecret: "test-secret"})

	body := ReleaseHoldRequest{
		HoldToken: holdToken.String(),
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID.String()+"/release", bytes.NewReader(jsonBody))
	token, _, _ := auth.GenerateToken(userID, "user@example.com", "CUSTOMER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/events/{id}/release", handler.ReleaseHold)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHoldSeats_EventAlreadyStarted(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()
	seatID := uuid.New()

	mock := &mockReservationQuerier{
		getEventByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error) {
			return generated.GetEventByIDRow{
				ID:             id,
				Status:         generated.EventStatusPUBLISHED,
				StartTime:      pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour), Valid: true},
				HoldTtlSeconds: 600,
			}, nil
		},
	}
	handler := NewReservationHandler(mock, nil, &config.Config{JWTSecret: "test-secret"})

	body := HoldSeatsRequest{
		SeatIDs: []string{seatID.String()},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+eventID.String()+"/hold", bytes.NewReader(jsonBody))
	token, _, _ := auth.GenerateToken(userID, "user@example.com", "CUSTOMER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/events/{id}/hold", handler.HoldSeats)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "EVENT_ALREADY_STARTED")
}

