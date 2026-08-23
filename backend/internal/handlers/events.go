package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/middleware"
	"ticket-booking/backend/internal/utils"
)

type EventHandler struct {
	queries generated.Querier
	pool    *pgxpool.Pool
}

func NewEventHandler(queries generated.Querier, pool *pgxpool.Pool) *EventHandler {
	return &EventHandler{
		queries: queries,
		pool:    pool,
	}
}

// ============================================================================
// DTOs
// ============================================================================

type CategoryPricingItem struct {
	SeatCategoryID string  `json:"seat_category_id"`
	Price          float64 `json:"price"`
	Currency       *string `json:"currency,omitempty"`
}

type CreateEventRequest struct {
	VenueID        string                `json:"venue_id"`
	Title          string                `json:"title"`
	Description    *string               `json:"description,omitempty"`
	EventType      string                `json:"event_type"`
	PosterURL      *string               `json:"poster_url,omitempty"`
	StartTime      time.Time             `json:"start_time"`
	EndTime        time.Time             `json:"end_time"`
	HoldTTLSeconds *int                  `json:"hold_ttl_seconds,omitempty"`
	Pricing        []CategoryPricingItem `json:"pricing,omitempty"`
}

type UpdateEventRequest struct {
	Title          string    `json:"title"`
	Description    *string   `json:"description,omitempty"`
	EventType      string    `json:"event_type"`
	PosterURL      *string   `json:"poster_url,omitempty"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	HoldTTLSeconds int       `json:"hold_ttl_seconds"`
	Status         string    `json:"status"`
}

type SetEventPricingRequest struct {
	Pricing []CategoryPricingItem `json:"pricing"`
}

type EventPricingResponse struct {
	ID                  string  `json:"id"`
	EventID             string  `json:"event_id"`
	SeatCategoryID      string  `json:"seat_category_id"`
	CategoryName        string  `json:"category_name,omitempty"`
	CategoryDescription *string `json:"category_description,omitempty"`
	CategoryColor       string  `json:"category_color,omitempty"`
	Price               float64 `json:"price"`
	Currency            string  `json:"currency"`
}

type EventResponse struct {
	ID             string                 `json:"id"`
	OrganiserID    string                 `json:"organiser_id"`
	OrganiserName  *string                `json:"organiser_name,omitempty"`
	OrganiserEmail *string                `json:"organiser_email,omitempty"`
	VenueID        string                 `json:"venue_id"`
	VenueName      *string                `json:"venue_name,omitempty"`
	VenueAddress   *string                `json:"venue_address,omitempty"`
	VenueCity      *string                `json:"venue_city,omitempty"`
	VenueCapacity  *int                   `json:"venue_capacity,omitempty"`
	Title          string                 `json:"title"`
	Description    *string                `json:"description,omitempty"`
	EventType      string                 `json:"event_type"`
	PosterURL      *string                `json:"poster_url,omitempty"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        time.Time              `json:"end_time"`
	HoldTTLSeconds int                    `json:"hold_ttl_seconds"`
	Status         string                 `json:"status"`
	Pricing        []EventPricingResponse `json:"pricing,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

type CategoryBreakdownResponse struct {
	SeatCategoryID string  `json:"seat_category_id"`
	CategoryName   string  `json:"category_name"`
	ColorCode      string  `json:"color_code"`
	TotalSeats     int64   `json:"total_seats"`
	BookedSeats    int64   `json:"booked_seats"`
	Revenue        float64 `json:"revenue"`
	WaitlistCount  int64   `json:"waitlist_count"`
}

type EventAnalyticsResponse struct {
	EventID                string                      `json:"event_id"`
	EventTitle             string                      `json:"event_title"`
	EventStatus            string                      `json:"event_status"`
	StartTime              time.Time                   `json:"start_time"`
	TotalCapacity          int                         `json:"total_capacity"`
	ConfirmedBookingsCount int64                       `json:"confirmed_bookings_count"`
	CancelledBookingsCount int64                       `json:"cancelled_bookings_count"`
	ValidTicketsCount      int64                       `json:"valid_tickets_count"`
	CheckedInTicketsCount  int64                       `json:"checked_in_tickets_count"`
	TotalRevenue           float64                     `json:"total_revenue"`
	OccupancyPercentage    float64                     `json:"occupancy_percentage"`
	WaitlistWaitingCount   int64                       `json:"waitlist_waiting_count"`
	CategoryBreakdown      []CategoryBreakdownResponse `json:"category_breakdown"`
}

// ============================================================================
// HELPERS
// ============================================================================

func parseNumericValue(v interface{}) float64 {
	if v == nil {
		return 0.0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case int:
		return float64(val)
	case string:
		var f float64
		_, _ = fmt.Sscanf(val, "%f", &f)
		return f
	case []byte:
		var f float64
		_, _ = fmt.Sscanf(string(val), "%f", &f)
		return f
	case pgtype.Numeric:
		return utils.PgtypeNumericToFloat64(val)
	default:
		return 0.0
	}
}

func parseEventType(s string) (generated.EventType, error) {
	upper := strings.ToUpper(strings.TrimSpace(s))
	switch upper {
	case "MOVIE":
		return generated.EventTypeMOVIE, nil
	case "CONCERT":
		return generated.EventTypeCONCERT, nil
	case "THEATRE":
		return generated.EventTypeTHEATRE, nil
	case "SPORTS":
		return generated.EventTypeSPORTS, nil
	case "OTHER":
		return generated.EventTypeOTHER, nil
	default:
		return "", errors.New("invalid event type: must be MOVIE, CONCERT, THEATRE, SPORTS, or OTHER")
	}
}

func parseEventStatus(s string) (generated.EventStatus, error) {
	upper := strings.ToUpper(strings.TrimSpace(s))
	switch upper {
	case "DRAFT":
		return generated.EventStatusDRAFT, nil
	case "PUBLISHED":
		return generated.EventStatusPUBLISHED, nil
	case "CANCELLED":
		return generated.EventStatusCANCELLED, nil
	case "COMPLETED":
		return generated.EventStatusCOMPLETED, nil
	default:
		return "", errors.New("invalid event status: must be DRAFT, PUBLISHED, CANCELLED, or COMPLETED")
	}
}

func formatEventResponse(e generated.Event, pricing []EventPricingResponse) EventResponse {
	var desc, poster *string
	if e.Description.Valid && e.Description.String != "" {
		desc = &e.Description.String
	}
	if e.PosterUrl.Valid && e.PosterUrl.String != "" {
		poster = &e.PosterUrl.String
	}

	return EventResponse{
		ID:             utils.PgtypeToUUID(e.ID).String(),
		OrganiserID:    utils.PgtypeToUUID(e.OrganiserID).String(),
		VenueID:        utils.PgtypeToUUID(e.VenueID).String(),
		Title:          e.Title,
		Description:    desc,
		EventType:      string(e.EventType),
		PosterURL:      poster,
		StartTime:      utils.PgtypeToTime(e.StartTime),
		EndTime:        utils.PgtypeToTime(e.EndTime),
		HoldTTLSeconds: int(e.HoldTtlSeconds),
		Status:         string(e.Status),
		Pricing:        pricing,
		CreatedAt:      utils.PgtypeToTime(e.CreatedAt),
		UpdatedAt:      utils.PgtypeToTime(e.UpdatedAt),
	}
}

// ============================================================================
// ORGANISER EVENT HANDLERS
// ============================================================================

// CreateEvent handles organiser event creation wizard.
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	organiserID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_TITLE", "event title is required")
		return
	}

	venueUUID, err := uuid.Parse(req.VenueID)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_VENUE_ID", "invalid venue ID format")
		return
	}

	eventType, err := parseEventType(req.EventType)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_EVENT_TYPE", err.Error())
		return
	}

	if req.StartTime.IsZero() || req.EndTime.IsZero() {
		RespondError(w, http.StatusBadRequest, "INVALID_TIME", "valid start_time and end_time are required")
		return
	}

	if !req.EndTime.After(req.StartTime) {
		RespondError(w, http.StatusBadRequest, "INVALID_TIME_ORDER", "end_time must be after start_time")
		return
	}

	holdTTL := int32(600)
	if req.HoldTTLSeconds != nil && *req.HoldTTLSeconds > 0 {
		holdTTL = int32(*req.HoldTTLSeconds)
	}

	descText := pgtype.Text{Valid: false}
	if req.Description != nil && strings.TrimSpace(*req.Description) != "" {
		descText = pgtype.Text{String: strings.TrimSpace(*req.Description), Valid: true}
	}

	posterText := pgtype.Text{Valid: false}
	if req.PosterURL != nil && strings.TrimSpace(*req.PosterURL) != "" {
		posterText = pgtype.Text{String: strings.TrimSpace(*req.PosterURL), Valid: true}
	}

	event, err := h.queries.CreateEvent(r.Context(), generated.CreateEventParams{
		OrganiserID:    utils.UUIDToPgtype(organiserID),
		VenueID:        utils.UUIDToPgtype(venueUUID),
		Title:          title,
		Description:    descText,
		EventType:      eventType,
		PosterUrl:      posterText,
		StartTime:      utils.TimeToPgtype(req.StartTime),
		EndTime:        utils.TimeToPgtype(req.EndTime),
		HoldTtlSeconds: holdTTL,
		Status:         generated.EventStatusDRAFT,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to create event")
		return
	}

	var createdPricing []EventPricingResponse
	if len(req.Pricing) > 0 {
		for _, p := range req.Pricing {
			catUUID, err := uuid.Parse(p.SeatCategoryID)
			if err != nil || p.Price < 0 {
				continue
			}
			currency := "USD"
			if p.Currency != nil && strings.TrimSpace(*p.Currency) != "" {
				currency = strings.ToUpper(strings.TrimSpace(*p.Currency))
			}
			pricingRow, err := h.queries.SetEventPricing(r.Context(), generated.SetEventPricingParams{
				EventID:        event.ID,
				SeatCategoryID: utils.UUIDToPgtype(catUUID),
				Price:          utils.Float64ToPgtypeNumeric(p.Price),
				Currency:       currency,
			})
			if err == nil {
				createdPricing = append(createdPricing, EventPricingResponse{
					ID:             utils.PgtypeToUUID(pricingRow.ID).String(),
					EventID:        utils.PgtypeToUUID(pricingRow.EventID).String(),
					SeatCategoryID: utils.PgtypeToUUID(pricingRow.SeatCategoryID).String(),
					Price:          p.Price,
					Currency:       pricingRow.Currency,
				})
			}
		}
	}

	RespondSuccess(w, http.StatusCreated, formatEventResponse(event, createdPricing), "Event created successfully")
}

// ListOrganiserEvents lists all events owned by the authenticated organiser.
func (h *EventHandler) ListOrganiserEvents(w http.ResponseWriter, r *http.Request) {
	organiserID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	events, err := h.queries.ListOrganiserEvents(r.Context(), utils.UUIDToPgtype(organiserID))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve organiser events")
		return
	}

	res := make([]EventResponse, 0, len(events))
	for _, e := range events {
		var desc, poster *string
		if e.Description.Valid && e.Description.String != "" {
			desc = &e.Description.String
		}
		if e.PosterUrl.Valid && e.PosterUrl.String != "" {
			poster = &e.PosterUrl.String
		}
		vCap := int(e.VenueCapacity)
		res = append(res, EventResponse{
			ID:             utils.PgtypeToUUID(e.ID).String(),
			OrganiserID:    utils.PgtypeToUUID(e.OrganiserID).String(),
			VenueID:        utils.PgtypeToUUID(e.VenueID).String(),
			VenueName:      &e.VenueName,
			VenueCity:      &e.VenueCity,
			VenueCapacity:  &vCap,
			Title:          e.Title,
			Description:    desc,
			EventType:      string(e.EventType),
			PosterURL:      poster,
			StartTime:      utils.PgtypeToTime(e.StartTime),
			EndTime:        utils.PgtypeToTime(e.EndTime),
			HoldTTLSeconds: int(e.HoldTtlSeconds),
			Status:         string(e.Status),
			CreatedAt:      utils.PgtypeToTime(e.CreatedAt),
			UpdatedAt:      utils.PgtypeToTime(e.UpdatedAt),
		})
	}

	RespondSuccess(w, http.StatusOK, res, "Organiser events retrieved")
}

// GetOrganiserEvent retrieves an event owned by the organiser with pricing tiers.
func (h *EventHandler) GetOrganiserEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid event ID format")
		return
	}

	event, err := h.queries.GetEventByID(r.Context(), utils.UUIDToPgtype(eventID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "event not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve event")
		return
	}

	pricingRows, _ := h.queries.GetEventPricingByEventID(r.Context(), utils.UUIDToPgtype(eventID))
	pricing := make([]EventPricingResponse, 0, len(pricingRows))
	for _, pr := range pricingRows {
		var catDesc *string
		if pr.CategoryDescription.Valid && pr.CategoryDescription.String != "" {
			catDesc = &pr.CategoryDescription.String
		}
		pricing = append(pricing, EventPricingResponse{
			ID:                  utils.PgtypeToUUID(pr.ID).String(),
			EventID:             utils.PgtypeToUUID(pr.EventID).String(),
			SeatCategoryID:      utils.PgtypeToUUID(pr.SeatCategoryID).String(),
			CategoryName:        pr.CategoryName,
			CategoryDescription: catDesc,
			CategoryColor:       pr.CategoryColor,
			Price:               utils.PgtypeNumericToFloat64(pr.Price),
			Currency:            pr.Currency,
		})
	}

	var desc, poster, vState *string
	if event.Description.Valid && event.Description.String != "" {
		desc = &event.Description.String
	}
	if event.PosterUrl.Valid && event.PosterUrl.String != "" {
		poster = &event.PosterUrl.String
	}
	if event.VenueState.Valid && event.VenueState.String != "" {
		vState = &event.VenueState.String
	}
	vCap := int(event.VenueCapacity)
	_ = vState

	res := EventResponse{
		ID:             utils.PgtypeToUUID(event.ID).String(),
		OrganiserID:    utils.PgtypeToUUID(event.OrganiserID).String(),
		OrganiserName:  &event.OrganiserName,
		OrganiserEmail: &event.OrganiserEmail,
		VenueID:        utils.PgtypeToUUID(event.VenueID).String(),
		VenueName:      &event.VenueName,
		VenueAddress:   &event.VenueAddress,
		VenueCity:      &event.VenueCity,
		VenueCapacity:  &vCap,
		Title:          event.Title,
		Description:    desc,
		EventType:      string(event.EventType),
		PosterURL:      poster,
		StartTime:      utils.PgtypeToTime(event.StartTime),
		EndTime:        utils.PgtypeToTime(event.EndTime),
		HoldTTLSeconds: int(event.HoldTtlSeconds),
		Status:         string(event.Status),
		Pricing:        pricing,
		CreatedAt:      utils.PgtypeToTime(event.CreatedAt),
		UpdatedAt:      utils.PgtypeToTime(event.UpdatedAt),
	}

	RespondSuccess(w, http.StatusOK, res, "Event details retrieved")
}

// UpdateEvent updates event parameters.
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid event ID format")
		return
	}

	organiserID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_TITLE", "event title cannot be empty")
		return
	}

	eventType, err := parseEventType(req.EventType)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_EVENT_TYPE", err.Error())
		return
	}

	eventStatus, err := parseEventStatus(req.Status)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_EVENT_STATUS", err.Error())
		return
	}

	if !req.EndTime.After(req.StartTime) {
		RespondError(w, http.StatusBadRequest, "INVALID_TIME_ORDER", "end_time must be after start_time")
		return
	}

	descText := pgtype.Text{Valid: false}
	if req.Description != nil && strings.TrimSpace(*req.Description) != "" {
		descText = pgtype.Text{String: strings.TrimSpace(*req.Description), Valid: true}
	}

	posterText := pgtype.Text{Valid: false}
	if req.PosterURL != nil && strings.TrimSpace(*req.PosterURL) != "" {
		posterText = pgtype.Text{String: strings.TrimSpace(*req.PosterURL), Valid: true}
	}

	holdTTL := int32(req.HoldTTLSeconds)
	if holdTTL <= 0 {
		holdTTL = 600
	}

	event, err := h.queries.UpdateEvent(r.Context(), generated.UpdateEventParams{
		ID:             utils.UUIDToPgtype(eventID),
		Title:          title,
		Description:    descText,
		EventType:      eventType,
		PosterUrl:      posterText,
		StartTime:      utils.TimeToPgtype(req.StartTime),
		EndTime:        utils.TimeToPgtype(req.EndTime),
		HoldTtlSeconds: holdTTL,
		Status:         eventStatus,
		OrganiserID:    utils.UUIDToPgtype(organiserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "event not found or access denied")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to update event")
		return
	}

	RespondSuccess(w, http.StatusOK, formatEventResponse(event, nil), "Event updated successfully")
}

// PublishEvent publishes a draft event.
func (h *EventHandler) PublishEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid event ID format")
		return
	}

	organiserID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	event, err := h.queries.PublishEvent(r.Context(), generated.PublishEventParams{
		ID:          utils.UUIDToPgtype(eventID),
		OrganiserID: utils.UUIDToPgtype(organiserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "event not found or access denied")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to publish event")
		return
	}

	RespondSuccess(w, http.StatusOK, formatEventResponse(event, nil), "Event published successfully")
}

// CancelEvent cancels an event.
func (h *EventHandler) CancelEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid event ID format")
		return
	}

	organiserID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	event, err := h.queries.CancelEvent(r.Context(), generated.CancelEventParams{
		ID:          utils.UUIDToPgtype(eventID),
		OrganiserID: utils.UUIDToPgtype(organiserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "event not found or access denied")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to cancel event")
		return
	}

	RespondSuccess(w, http.StatusOK, formatEventResponse(event, nil), "Event cancelled successfully")
}

// SetEventPricing sets or updates per-category pricing for an event.
func (h *EventHandler) SetEventPricing(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid event ID format")
		return
	}

	var req SetEventPricingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	if len(req.Pricing) == 0 {
		RespondError(w, http.StatusBadRequest, "EMPTY_PRICING", "pricing list cannot be empty")
		return
	}

	res := make([]EventPricingResponse, 0, len(req.Pricing))
	for _, p := range req.Pricing {
		catUUID, err := uuid.Parse(p.SeatCategoryID)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_CATEGORY_ID", "invalid seat_category_id: "+p.SeatCategoryID)
			return
		}
		if p.Price < 0 {
			RespondError(w, http.StatusBadRequest, "INVALID_PRICE", "price must be non-negative")
			return
		}

		currency := "USD"
		if p.Currency != nil && strings.TrimSpace(*p.Currency) != "" {
			currency = strings.ToUpper(strings.TrimSpace(*p.Currency))
		}

		pricingRow, err := h.queries.SetEventPricing(r.Context(), generated.SetEventPricingParams{
			EventID:        utils.UUIDToPgtype(eventID),
			SeatCategoryID: utils.UUIDToPgtype(catUUID),
			Price:          utils.Float64ToPgtypeNumeric(p.Price),
			Currency:       currency,
		})
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to set pricing for category "+p.SeatCategoryID)
			return
		}

		res = append(res, EventPricingResponse{
			ID:             utils.PgtypeToUUID(pricingRow.ID).String(),
			EventID:        utils.PgtypeToUUID(pricingRow.EventID).String(),
			SeatCategoryID: utils.PgtypeToUUID(pricingRow.SeatCategoryID).String(),
			Price:          p.Price,
			Currency:       pricingRow.Currency,
		})
	}

	RespondSuccess(w, http.StatusOK, res, "Event pricing updated successfully")
}

// GetEventPricing retrieves pricing for an event.
func (h *EventHandler) GetEventPricing(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid event ID format")
		return
	}

	pricingRows, err := h.queries.GetEventPricingByEventID(r.Context(), utils.UUIDToPgtype(eventID))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve event pricing")
		return
	}

	res := make([]EventPricingResponse, 0, len(pricingRows))
	for _, pr := range pricingRows {
		var desc *string
		if pr.CategoryDescription.Valid && pr.CategoryDescription.String != "" {
			desc = &pr.CategoryDescription.String
		}
		res = append(res, EventPricingResponse{
			ID:                  utils.PgtypeToUUID(pr.ID).String(),
			EventID:             utils.PgtypeToUUID(pr.EventID).String(),
			SeatCategoryID:      utils.PgtypeToUUID(pr.SeatCategoryID).String(),
			CategoryName:        pr.CategoryName,
			CategoryDescription: desc,
			CategoryColor:       pr.CategoryColor,
			Price:               utils.PgtypeNumericToFloat64(pr.Price),
			Currency:            pr.Currency,
		})
	}

	RespondSuccess(w, http.StatusOK, res, "Event pricing retrieved")
}

// GetEventAnalytics returns revenue and booking summary for an organiser event.
func (h *EventHandler) GetEventAnalytics(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid event ID format")
		return
	}

	organiserID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	summary, err := h.queries.GetEventBookingSummary(r.Context(), generated.GetEventBookingSummaryParams{
		ID:          utils.UUIDToPgtype(eventID),
		OrganiserID: utils.UUIDToPgtype(organiserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "event not found or access denied")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to calculate event analytics")
		return
	}

	breakdownRows, err := h.queries.GetEventCategoryBreakdown(r.Context(), utils.UUIDToPgtype(eventID))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to calculate category breakdown")
		return
	}

	breakdown := make([]CategoryBreakdownResponse, 0, len(breakdownRows))
	for _, br := range breakdownRows {
		breakdown = append(breakdown, CategoryBreakdownResponse{
			SeatCategoryID: utils.PgtypeToUUID(br.SeatCategoryID).String(),
			CategoryName:   br.CategoryName,
			ColorCode:      br.ColorCode,
			TotalSeats:     br.TotalSeats,
			BookedSeats:    br.BookedSeats,
			Revenue:        parseNumericValue(br.Revenue),
			WaitlistCount:  br.WaitlistCount,
		})
	}

	analytics := EventAnalyticsResponse{
		EventID:                utils.PgtypeToUUID(summary.EventID).String(),
		EventTitle:             summary.EventTitle,
		EventStatus:            string(summary.EventStatus),
		StartTime:              utils.PgtypeToTime(summary.StartTime),
		TotalCapacity:          int(summary.TotalCapacity),
		ConfirmedBookingsCount: summary.ConfirmedBookingsCount,
		CancelledBookingsCount: summary.CancelledBookingsCount,
		ValidTicketsCount:      summary.ValidTicketsCount,
		CheckedInTicketsCount:  summary.CheckedInTicketsCount,
		TotalRevenue:           parseNumericValue(summary.TotalRevenue),
		OccupancyPercentage:    parseNumericValue(summary.OccupancyPercentage),
		WaitlistWaitingCount:   summary.WaitlistWaitingCount,
		CategoryBreakdown:      breakdown,
	}

	RespondSuccess(w, http.StatusOK, analytics, "Event analytics calculated")
}

// ============================================================================
// PUBLIC EVENT DISCOVERY HANDLERS
// ============================================================================

// ListPublishedEvents lists and filters published events for customers.
func (h *EventHandler) ListPublishedEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := int32(20)
	if l := q.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = int32(val)
		}
	}

	offset := int32(0)
	if o := q.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = int32(val)
		}
	}

	var eventType generated.NullEventType
	if et := q.Get("event_type"); et != "" {
		if parsed, err := parseEventType(et); err == nil {
			eventType = generated.NullEventType{EventType: parsed, Valid: true}
		}
	}

	var search pgtype.Text
	if s := strings.TrimSpace(q.Get("search")); s != "" {
		search = pgtype.Text{String: s, Valid: true}
	}

	var fromDate pgtype.Timestamptz
	if fd := q.Get("from_date"); fd != "" {
		if t, err := time.Parse(time.RFC3339, fd); err == nil {
			fromDate = utils.TimeToPgtype(t)
		}
	}

	var toDate pgtype.Timestamptz
	if td := q.Get("to_date"); td != "" {
		if t, err := time.Parse(time.RFC3339, td); err == nil {
			toDate = utils.TimeToPgtype(t)
		}
	}

	events, err := h.queries.ListPublishedEvents(r.Context(), generated.ListPublishedEventsParams{
		Limit:     limit,
		Offset:    offset,
		EventType: eventType,
		Search:    search,
		FromDate:  fromDate,
		ToDate:    toDate,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to list published events")
		return
	}

	res := make([]EventResponse, 0, len(events))
	for _, e := range events {
		var desc, poster *string
		if e.Description.Valid && e.Description.String != "" {
			desc = &e.Description.String
		}
		if e.PosterUrl.Valid && e.PosterUrl.String != "" {
			poster = &e.PosterUrl.String
		}
		vCap := int(e.VenueCapacity)
		res = append(res, EventResponse{
			ID:             utils.PgtypeToUUID(e.ID).String(),
			OrganiserID:    utils.PgtypeToUUID(e.OrganiserID).String(),
			VenueID:        utils.PgtypeToUUID(e.VenueID).String(),
			VenueName:      &e.VenueName,
			VenueCity:      &e.VenueCity,
			VenueCapacity:  &vCap,
			Title:          e.Title,
			Description:    desc,
			EventType:      string(e.EventType),
			PosterURL:      poster,
			StartTime:      utils.PgtypeToTime(e.StartTime),
			EndTime:        utils.PgtypeToTime(e.EndTime),
			HoldTTLSeconds: int(e.HoldTtlSeconds),
			Status:         string(e.Status),
			CreatedAt:      utils.PgtypeToTime(e.CreatedAt),
			UpdatedAt:      utils.PgtypeToTime(e.UpdatedAt),
		})
	}

	RespondSuccess(w, http.StatusOK, res, "Published events retrieved")
}

// GetPublicEvent retrieves a single published event for customers including pricing.
func (h *EventHandler) GetPublicEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid event ID format")
		return
	}

	event, err := h.queries.GetEventByID(r.Context(), utils.UUIDToPgtype(eventID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "event not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve event")
		return
	}

	if event.Status != generated.EventStatusPUBLISHED {
		RespondError(w, http.StatusNotFound, "EVENT_NOT_AVAILABLE", "event is not published")
		return
	}

	pricingRows, _ := h.queries.GetEventPricingByEventID(r.Context(), utils.UUIDToPgtype(eventID))
	pricing := make([]EventPricingResponse, 0, len(pricingRows))
	for _, pr := range pricingRows {
		var catDesc *string
		if pr.CategoryDescription.Valid && pr.CategoryDescription.String != "" {
			catDesc = &pr.CategoryDescription.String
		}
		pricing = append(pricing, EventPricingResponse{
			ID:                  utils.PgtypeToUUID(pr.ID).String(),
			EventID:             utils.PgtypeToUUID(pr.EventID).String(),
			SeatCategoryID:      utils.PgtypeToUUID(pr.SeatCategoryID).String(),
			CategoryName:        pr.CategoryName,
			CategoryDescription: catDesc,
			CategoryColor:       pr.CategoryColor,
			Price:               utils.PgtypeNumericToFloat64(pr.Price),
			Currency:            pr.Currency,
		})
	}

	var desc, poster, vState *string
	if event.Description.Valid && event.Description.String != "" {
		desc = &event.Description.String
	}
	if event.PosterUrl.Valid && event.PosterUrl.String != "" {
		poster = &event.PosterUrl.String
	}
	if event.VenueState.Valid && event.VenueState.String != "" {
		vState = &event.VenueState.String
	}
	vCap := int(event.VenueCapacity)
	_ = vState

	res := EventResponse{
		ID:             utils.PgtypeToUUID(event.ID).String(),
		OrganiserID:    utils.PgtypeToUUID(event.OrganiserID).String(),
		OrganiserName:  &event.OrganiserName,
		OrganiserEmail: &event.OrganiserEmail,
		VenueID:        utils.PgtypeToUUID(event.VenueID).String(),
		VenueName:      &event.VenueName,
		VenueAddress:   &event.VenueAddress,
		VenueCity:      &event.VenueCity,
		VenueCapacity:  &vCap,
		Title:          event.Title,
		Description:    desc,
		EventType:      string(event.EventType),
		PosterURL:      poster,
		StartTime:      utils.PgtypeToTime(event.StartTime),
		EndTime:        utils.PgtypeToTime(event.EndTime),
		HoldTTLSeconds: int(event.HoldTtlSeconds),
		Status:         string(event.Status),
		Pricing:        pricing,
		CreatedAt:      utils.PgtypeToTime(event.CreatedAt),
		UpdatedAt:      utils.PgtypeToTime(event.UpdatedAt),
	}

	RespondSuccess(w, http.StatusOK, res, "Event details retrieved")
}
