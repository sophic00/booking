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
	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/middleware"
	"ticket-booking/backend/internal/utils"
)

type mockEventQuerier struct {
	generated.Querier
	createEventFunc               func(ctx context.Context, arg generated.CreateEventParams) (generated.Event, error)
	updateEventFunc               func(ctx context.Context, arg generated.UpdateEventParams) (generated.Event, error)
	publishEventFunc              func(ctx context.Context, arg generated.PublishEventParams) (generated.Event, error)
	cancelEventFunc               func(ctx context.Context, arg generated.CancelEventParams) (generated.Event, error)
	getEventByIDFunc              func(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error)
	listOrganiserEventsFunc       func(ctx context.Context, organiserID pgtype.UUID) ([]generated.ListOrganiserEventsRow, error)
	listPublishedEventsFunc       func(ctx context.Context, arg generated.ListPublishedEventsParams) ([]generated.ListPublishedEventsRow, error)
	setEventPricingFunc           func(ctx context.Context, arg generated.SetEventPricingParams) (generated.EventPricing, error)
	getEventPricingByEventIDFunc  func(ctx context.Context, eventID pgtype.UUID) ([]generated.GetEventPricingByEventIDRow, error)
	getEventBookingSummaryFunc    func(ctx context.Context, arg generated.GetEventBookingSummaryParams) (generated.GetEventBookingSummaryRow, error)
	getEventCategoryBreakdownFunc func(ctx context.Context, eventID pgtype.UUID) ([]generated.GetEventCategoryBreakdownRow, error)
}

func (m *mockEventQuerier) CreateEvent(ctx context.Context, arg generated.CreateEventParams) (generated.Event, error) {
	if m.createEventFunc != nil {
		return m.createEventFunc(ctx, arg)
	}
	return generated.Event{}, nil
}

func (m *mockEventQuerier) UpdateEvent(ctx context.Context, arg generated.UpdateEventParams) (generated.Event, error) {
	if m.updateEventFunc != nil {
		return m.updateEventFunc(ctx, arg)
	}
	return generated.Event{}, nil
}

func (m *mockEventQuerier) PublishEvent(ctx context.Context, arg generated.PublishEventParams) (generated.Event, error) {
	if m.publishEventFunc != nil {
		return m.publishEventFunc(ctx, arg)
	}
	return generated.Event{}, nil
}

func (m *mockEventQuerier) CancelEvent(ctx context.Context, arg generated.CancelEventParams) (generated.Event, error) {
	if m.cancelEventFunc != nil {
		return m.cancelEventFunc(ctx, arg)
	}
	return generated.Event{}, nil
}

func (m *mockEventQuerier) GetEventByID(ctx context.Context, id pgtype.UUID) (generated.GetEventByIDRow, error) {
	if m.getEventByIDFunc != nil {
		return m.getEventByIDFunc(ctx, id)
	}
	return generated.GetEventByIDRow{}, pgx.ErrNoRows
}

func (m *mockEventQuerier) ListOrganiserEvents(ctx context.Context, organiserID pgtype.UUID) ([]generated.ListOrganiserEventsRow, error) {
	if m.listOrganiserEventsFunc != nil {
		return m.listOrganiserEventsFunc(ctx, organiserID)
	}
	return nil, nil
}

func (m *mockEventQuerier) ListPublishedEvents(ctx context.Context, arg generated.ListPublishedEventsParams) ([]generated.ListPublishedEventsRow, error) {
	if m.listPublishedEventsFunc != nil {
		return m.listPublishedEventsFunc(ctx, arg)
	}
	return nil, nil
}

func (m *mockEventQuerier) SetEventPricing(ctx context.Context, arg generated.SetEventPricingParams) (generated.EventPricing, error) {
	if m.setEventPricingFunc != nil {
		return m.setEventPricingFunc(ctx, arg)
	}
	return generated.EventPricing{}, nil
}

func (m *mockEventQuerier) GetEventPricingByEventID(ctx context.Context, eventID pgtype.UUID) ([]generated.GetEventPricingByEventIDRow, error) {
	if m.getEventPricingByEventIDFunc != nil {
		return m.getEventPricingByEventIDFunc(ctx, eventID)
	}
	return nil, nil
}

func (m *mockEventQuerier) GetEventBookingSummary(ctx context.Context, arg generated.GetEventBookingSummaryParams) (generated.GetEventBookingSummaryRow, error) {
	if m.getEventBookingSummaryFunc != nil {
		return m.getEventBookingSummaryFunc(ctx, arg)
	}
	return generated.GetEventBookingSummaryRow{}, pgx.ErrNoRows
}

func (m *mockEventQuerier) GetEventCategoryBreakdown(ctx context.Context, eventID pgtype.UUID) ([]generated.GetEventCategoryBreakdownRow, error) {
	if m.getEventCategoryBreakdownFunc != nil {
		return m.getEventCategoryBreakdownFunc(ctx, eventID)
	}
	return nil, nil
}

func TestCreateEvent_Success(t *testing.T) {
	organiserID := uuid.New()
	venueID := uuid.New()
	eventID := uuid.New()
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(3 * time.Hour)

	mock := &mockEventQuerier{
		createEventFunc: func(ctx context.Context, arg generated.CreateEventParams) (generated.Event, error) {
			assert.Equal(t, "Coldplay Live in Mumbai", arg.Title)
			assert.Equal(t, generated.EventTypeCONCERT, arg.EventType)
			assert.Equal(t, utils.UUIDToPgtype(venueID), arg.VenueID)
			return generated.Event{
				ID:             utils.UUIDToPgtype(eventID),
				OrganiserID:    arg.OrganiserID,
				VenueID:        arg.VenueID,
				Title:          arg.Title,
				EventType:      arg.EventType,
				StartTime:      arg.StartTime,
				EndTime:        arg.EndTime,
				HoldTtlSeconds: arg.HoldTtlSeconds,
				Status:         generated.EventStatusDRAFT,
				CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	handler := NewEventHandler(mock, nil)

	body := CreateEventRequest{
		VenueID:   venueID.String(),
		Title:     "Coldplay Live in Mumbai",
		EventType: "CONCERT",
		StartTime: startTime,
		EndTime:   endTime,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/organiser/events", bytes.NewReader(jsonBody))
	token, _, _ := auth.GenerateToken(organiserID, "org@example.com", "ORGANISER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/organiser/events", handler.CreateEvent)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp struct {
		Success bool          `json:"success"`
		Data    EventResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Coldplay Live in Mumbai", resp.Data.Title)
	assert.Equal(t, "CONCERT", resp.Data.EventType)
	assert.Equal(t, "DRAFT", resp.Data.Status)
}

func TestCreateEvent_InvalidTimes(t *testing.T) {
	organiserID := uuid.New()
	venueID := uuid.New()
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(-2 * time.Hour) // End before start!

	handler := NewEventHandler(&mockEventQuerier{}, nil)

	body := CreateEventRequest{
		VenueID:   venueID.String(),
		Title:     "Invalid Time Event",
		EventType: "MOVIE",
		StartTime: startTime,
		EndTime:   endTime,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/organiser/events", bytes.NewReader(jsonBody))
	token, _, _ := auth.GenerateToken(organiserID, "org@example.com", "ORGANISER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/organiser/events", handler.CreateEvent)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "INVALID_TIME_ORDER")
}

func TestPublishEvent_Success(t *testing.T) {
	organiserID := uuid.New()
	eventID := uuid.New()

	mock := &mockEventQuerier{
		publishEventFunc: func(ctx context.Context, arg generated.PublishEventParams) (generated.Event, error) {
			return generated.Event{
				ID:          arg.ID,
				OrganiserID: arg.OrganiserID,
				Title:       "Inception Special Screening",
				EventType:   generated.EventTypeMOVIE,
				Status:      generated.EventStatusPUBLISHED,
				CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	handler := NewEventHandler(mock, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/organiser/events/"+eventID.String()+"/publish", nil)
	token, _, _ := auth.GenerateToken(organiserID, "org@example.com", "ORGANISER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Post("/api/v1/organiser/events/{id}/publish", handler.PublishEvent)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool          `json:"success"`
		Data    EventResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "PUBLISHED", resp.Data.Status)
}

func TestSetEventPricing_Success(t *testing.T) {
	eventID := uuid.New()
	catID1 := uuid.New()
	catID2 := uuid.New()

	mock := &mockEventQuerier{
		setEventPricingFunc: func(ctx context.Context, arg generated.SetEventPricingParams) (generated.EventPricing, error) {
			return generated.EventPricing{
				ID:             utils.UUIDToPgtype(uuid.New()),
				EventID:        arg.EventID,
				SeatCategoryID: arg.SeatCategoryID,
				Price:          arg.Price,
				Currency:       arg.Currency,
				CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	handler := NewEventHandler(mock, nil)

	body := SetEventPricingRequest{
		Pricing: []CategoryPricingItem{
			{SeatCategoryID: catID1.String(), Price: 120.50},
			{SeatCategoryID: catID2.String(), Price: 45.00},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/organiser/events/"+eventID.String()+"/pricing", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/api/v1/organiser/events/{id}/pricing", handler.SetEventPricing)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool                   `json:"success"`
		Data    []EventPricingResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Data, 2)
}

func TestGetEventAnalytics_Success(t *testing.T) {
	organiserID := uuid.New()
	eventID := uuid.New()
	catID := uuid.New()

	mock := &mockEventQuerier{
		getEventBookingSummaryFunc: func(ctx context.Context, arg generated.GetEventBookingSummaryParams) (generated.GetEventBookingSummaryRow, error) {
			return generated.GetEventBookingSummaryRow{
				EventID:                arg.ID,
				EventTitle:             "Rock Concert 2026",
				EventStatus:            generated.EventStatusPUBLISHED,
				StartTime:              pgtype.Timestamptz{Time: time.Now(), Valid: true},
				TotalCapacity:          1000,
				ConfirmedBookingsCount: 150,
				CancelledBookingsCount: 10,
				ValidTicketsCount:      400,
				TotalRevenue:           25000.50,
				OccupancyPercentage:    40.00,
				WaitlistWaitingCount:   25,
			}, nil
		},
		getEventCategoryBreakdownFunc: func(ctx context.Context, eID pgtype.UUID) ([]generated.GetEventCategoryBreakdownRow, error) {
			return []generated.GetEventCategoryBreakdownRow{
				{
					SeatCategoryID: utils.UUIDToPgtype(catID),
					CategoryName:   "VIP",
					ColorCode:      "#F59E0B",
					TotalSeats:     200,
					BookedSeats:    180,
					Revenue:        18000.00,
					WaitlistCount:  20,
				},
			}, nil
		},
	}

	handler := NewEventHandler(mock, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organiser/events/"+eventID.String()+"/analytics", nil)
	token, _, _ := auth.GenerateToken(organiserID, "org@example.com", "ORGANISER", "test-secret", 1)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware.Authenticate("test-secret"))
	r.Get("/api/v1/organiser/events/{id}/analytics", handler.GetEventAnalytics)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool                   `json:"success"`
		Data    EventAnalyticsResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Rock Concert 2026", resp.Data.EventTitle)
	assert.Equal(t, int64(400), resp.Data.ValidTicketsCount)
	assert.Equal(t, 25000.50, resp.Data.TotalRevenue)
	assert.Equal(t, 40.00, resp.Data.OccupancyPercentage)
	assert.Len(t, resp.Data.CategoryBreakdown, 1)
	assert.Equal(t, "VIP", resp.Data.CategoryBreakdown[0].CategoryName)
}

func TestListPublishedEvents_Success(t *testing.T) {
	vName := "Starlight Arena"
	vCity := "New York"
	vCap := int32(500)

	mock := &mockEventQuerier{
		listPublishedEventsFunc: func(ctx context.Context, arg generated.ListPublishedEventsParams) ([]generated.ListPublishedEventsRow, error) {
			return []generated.ListPublishedEventsRow{
				{
					ID:            utils.UUIDToPgtype(uuid.New()),
					OrganiserID:   utils.UUIDToPgtype(uuid.New()),
					VenueID:       utils.UUIDToPgtype(uuid.New()),
					Title:         "Jazz Night",
					EventType:     generated.EventTypeCONCERT,
					StartTime:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
					EndTime:       pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour), Valid: true},
					Status:        generated.EventStatusPUBLISHED,
					VenueName:     vName,
					VenueCity:     vCity,
					VenueCapacity: vCap,
					CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
					UpdatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
				},
			}, nil
		},
	}

	handler := NewEventHandler(mock, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events?search=Jazz", nil)
	rr := httptest.NewRecorder()

	handler.ListPublishedEvents(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool            `json:"success"`
		Data    []EventResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "Jazz Night", resp.Data[0].Title)
}
