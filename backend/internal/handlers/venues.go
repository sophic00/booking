package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/middleware"
	"ticket-booking/backend/internal/utils"
)

var hexColorRegex = regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)

type VenueHandler struct {
	queries generated.Querier
	pool    *pgxpool.Pool
}

func NewVenueHandler(queries generated.Querier, pool *pgxpool.Pool) *VenueHandler {
	return &VenueHandler{
		queries: queries,
		pool:    pool,
	}
}

// ============================================================================
// DTOs
// ============================================================================

type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	ColorCode   *string `json:"color_code,omitempty"`
}

type CategoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	ColorCode   string    `json:"color_code"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateVenueRequest struct {
	Name          string  `json:"name"`
	Address       string  `json:"address"`
	City          string  `json:"city"`
	State         *string `json:"state,omitempty"`
	Country       *string `json:"country,omitempty"`
	TotalCapacity *int    `json:"total_capacity,omitempty"`
}

type UpdateVenueRequest struct {
	Name          string  `json:"name"`
	Address       string  `json:"address"`
	City          string  `json:"city"`
	State         *string `json:"state,omitempty"`
	Country       string  `json:"country"`
	TotalCapacity int     `json:"total_capacity"`
}

type VenueResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	City          string    `json:"city"`
	State         *string   `json:"state,omitempty"`
	Country       string    `json:"country"`
	TotalCapacity int       `json:"total_capacity"`
	CreatedBy     *string   `json:"created_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateSeatRequest struct {
	SeatCategoryID string `json:"seat_category_id"`
	RowLabel       string `json:"row_label"`
	SeatNumber     string `json:"seat_number"`
	GridRow        int    `json:"grid_row"`
	GridCol        int    `json:"grid_col"`
	IsActive       *bool  `json:"is_active,omitempty"`
}

type BatchCreateSeatsRequest struct {
	Replace bool                `json:"replace"`
	Seats   []CreateSeatRequest `json:"seats"`
}

type SeatResponse struct {
	ID             string    `json:"id"`
	VenueID        string    `json:"venue_id"`
	SeatCategoryID string    `json:"seat_category_id"`
	CategoryName   string    `json:"category_name,omitempty"`
	CategoryColor  string    `json:"category_color,omitempty"`
	RowLabel       string    `json:"row_label"`
	SeatNumber     string    `json:"seat_number"`
	GridRow        int       `json:"grid_row"`
	GridCol        int       `json:"grid_col"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

// ============================================================================
// HELPERS
// ============================================================================

func formatCategoryResponse(c generated.SeatCategory) CategoryResponse {
	var desc *string
	if c.Description.Valid && c.Description.String != "" {
		desc = &c.Description.String
	}
	return CategoryResponse{
		ID:          utils.PgtypeToUUID(c.ID).String(),
		Name:        c.Name,
		Description: desc,
		ColorCode:   c.ColorCode,
		CreatedAt:   utils.PgtypeToTime(c.CreatedAt),
	}
}

func formatVenueResponse(v generated.Venue) VenueResponse {
	var state *string
	if v.State.Valid && v.State.String != "" {
		state = &v.State.String
	}
	var createdBy *string
	if v.CreatedBy.Valid {
		idStr := utils.PgtypeToUUID(v.CreatedBy).String()
		createdBy = &idStr
	}
	return VenueResponse{
		ID:            utils.PgtypeToUUID(v.ID).String(),
		Name:          v.Name,
		Address:       v.Address,
		City:          v.City,
		State:         state,
		Country:       v.Country,
		TotalCapacity: int(v.TotalCapacity),
		CreatedBy:     createdBy,
		CreatedAt:     utils.PgtypeToTime(v.CreatedAt),
		UpdatedAt:     utils.PgtypeToTime(v.UpdatedAt),
	}
}

// ============================================================================
// SEAT CATEGORY HANDLERS
// ============================================================================

// ListCategories lists all available seat categories.
func (h *VenueHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.queries.ListSeatCategories(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve seat categories")
		return
	}

	res := make([]CategoryResponse, 0, len(categories))
	for _, c := range categories {
		res = append(res, formatCategoryResponse(c))
	}

	RespondSuccess(w, http.StatusOK, res, "Seat categories retrieved")
}

// CreateCategory handles admin category creation.
func (h *VenueHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_NAME", "category name is required")
		return
	}

	colorCode := "#3B82F6"
	if req.ColorCode != nil && strings.TrimSpace(*req.ColorCode) != "" {
		trimmed := strings.TrimSpace(*req.ColorCode)
		if !hexColorRegex.MatchString(trimmed) {
			RespondError(w, http.StatusBadRequest, "INVALID_COLOR", "color code must be a valid hex color (e.g. #3B82F6)")
			return
		}
		colorCode = trimmed
	}

	descText := pgtype.Text{Valid: false}
	if req.Description != nil && strings.TrimSpace(*req.Description) != "" {
		descText = pgtype.Text{String: strings.TrimSpace(*req.Description), Valid: true}
	}

	category, err := h.queries.CreateSeatCategory(r.Context(), generated.CreateSeatCategoryParams{
		Name:        name,
		Description: descText,
		ColorCode:   colorCode,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			RespondError(w, http.StatusConflict, "CATEGORY_EXISTS", "seat category with this name already exists")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to create seat category")
		return
	}

	RespondSuccess(w, http.StatusCreated, formatCategoryResponse(category), "Seat category created")
}

// ============================================================================
// VENUE HANDLERS
// ============================================================================

// ListVenues lists all registered venues.
func (h *VenueHandler) ListVenues(w http.ResponseWriter, r *http.Request) {
	venues, err := h.queries.ListVenues(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve venues")
		return
	}

	res := make([]VenueResponse, 0, len(venues))
	for _, v := range venues {
		res = append(res, formatVenueResponse(v))
	}

	RespondSuccess(w, http.StatusOK, res, "Venues retrieved")
}

// GetVenue retrieves a single venue by its ID.
func (h *VenueHandler) GetVenue(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	venueID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid venue ID format")
		return
	}

	venue, err := h.queries.GetVenueByID(r.Context(), utils.UUIDToPgtype(venueID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "VENUE_NOT_FOUND", "venue not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve venue")
		return
	}

	RespondSuccess(w, http.StatusOK, formatVenueResponse(venue), "Venue retrieved")
}

// CreateVenue handles admin venue creation.
func (h *VenueHandler) CreateVenue(w http.ResponseWriter, r *http.Request) {
	var req CreateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	name := strings.TrimSpace(req.Name)
	address := strings.TrimSpace(req.Address)
	city := strings.TrimSpace(req.City)
	if name == "" || address == "" || city == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "name, address, and city are required")
		return
	}

	stateText := pgtype.Text{Valid: false}
	if req.State != nil && strings.TrimSpace(*req.State) != "" {
		stateText = pgtype.Text{String: strings.TrimSpace(*req.State), Valid: true}
	}

	country := "IN"
	if req.Country != nil && strings.TrimSpace(*req.Country) != "" {
		country = strings.TrimSpace(*req.Country)
	}

	capacity := int32(0)
	if req.TotalCapacity != nil && *req.TotalCapacity >= 0 {
		capacity = int32(*req.TotalCapacity)
	}

	var createdBy pgtype.UUID
	adminID, err := middleware.GetUserID(r.Context())
	if err == nil {
		createdBy = utils.UUIDToPgtype(adminID)
	}

	venue, err := h.queries.CreateVenue(r.Context(), generated.CreateVenueParams{
		Name:          name,
		Address:       address,
		City:          city,
		State:         stateText,
		Country:       country,
		TotalCapacity: capacity,
		CreatedBy:     createdBy,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to create venue")
		return
	}

	RespondSuccess(w, http.StatusCreated, formatVenueResponse(venue), "Venue created successfully")
}

// UpdateVenue handles admin venue updates.
func (h *VenueHandler) UpdateVenue(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	venueID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid venue ID format")
		return
	}

	var req UpdateVenueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	name := strings.TrimSpace(req.Name)
	address := strings.TrimSpace(req.Address)
	city := strings.TrimSpace(req.City)
	if name == "" || address == "" || city == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "name, address, and city are required")
		return
	}

	stateText := pgtype.Text{Valid: false}
	if req.State != nil && strings.TrimSpace(*req.State) != "" {
		stateText = pgtype.Text{String: strings.TrimSpace(*req.State), Valid: true}
	}

	country := strings.TrimSpace(req.Country)
	if country == "" {
		country = "IN"
	}

	venue, err := h.queries.UpdateVenue(r.Context(), generated.UpdateVenueParams{
		ID:            utils.UUIDToPgtype(venueID),
		Name:          name,
		Address:       address,
		City:          city,
		State:         stateText,
		Country:       country,
		TotalCapacity: int32(req.TotalCapacity),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "VENUE_NOT_FOUND", "venue not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to update venue")
		return
	}

	RespondSuccess(w, http.StatusOK, formatVenueResponse(venue), "Venue updated successfully")
}

// DeleteVenue handles admin venue deletion.
func (h *VenueHandler) DeleteVenue(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	venueID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid venue ID format")
		return
	}

	if err := h.queries.DeleteVenue(r.Context(), utils.UUIDToPgtype(venueID)); err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to delete venue")
		return
	}

	RespondSuccess(w, http.StatusOK, nil, "Venue deleted successfully")
}

// ============================================================================
// VISUAL SEAT LAYOUT HANDLERS
// ============================================================================

// GetVenueSeats retrieves all active seats for a venue layout.
func (h *VenueHandler) GetVenueSeats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	venueID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid venue ID format")
		return
	}

	seats, err := h.queries.GetSeatsByVenueID(r.Context(), utils.UUIDToPgtype(venueID))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve venue seats")
		return
	}

	res := make([]SeatResponse, 0, len(seats))
	for _, s := range seats {
		res = append(res, SeatResponse{
			ID:             utils.PgtypeToUUID(s.ID).String(),
			VenueID:        utils.PgtypeToUUID(s.VenueID).String(),
			SeatCategoryID: utils.PgtypeToUUID(s.SeatCategoryID).String(),
			CategoryName:   s.CategoryName,
			CategoryColor:  s.CategoryColor,
			RowLabel:       s.RowLabel,
			SeatNumber:     s.SeatNumber,
			GridRow:        int(s.GridRow),
			GridCol:        int(s.GridCol),
			IsActive:       s.IsActive,
			CreatedAt:      utils.PgtypeToTime(s.CreatedAt),
		})
	}

	RespondSuccess(w, http.StatusOK, res, "Venue seats retrieved")
}

// CreateSeat creates a single seat in a venue.
func (h *VenueHandler) CreateSeat(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	venueID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid venue ID format")
		return
	}

	var req CreateSeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	categoryUUID, err := uuid.Parse(req.SeatCategoryID)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_CATEGORY_ID", "invalid category ID")
		return
	}

	rowLabel := strings.TrimSpace(req.RowLabel)
	seatNumber := strings.TrimSpace(req.SeatNumber)
	if rowLabel == "" || seatNumber == "" || req.GridRow < 1 || req.GridCol < 1 {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "row label, seat number, and positive grid coordinates (row/col >= 1) are required")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	seat, err := h.queries.CreateSeat(r.Context(), generated.CreateSeatParams{
		VenueID:        utils.UUIDToPgtype(venueID),
		SeatCategoryID: utils.UUIDToPgtype(categoryUUID),
		RowLabel:       rowLabel,
		SeatNumber:     seatNumber,
		GridRow:        int32(req.GridRow),
		GridCol:        int32(req.GridCol),
		IsActive:       isActive,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			RespondError(w, http.StatusConflict, "SEAT_EXISTS", "seat with this row/number or grid coordinate already exists in this venue")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to create seat")
		return
	}

	RespondSuccess(w, http.StatusCreated, SeatResponse{
		ID:             utils.PgtypeToUUID(seat.ID).String(),
		VenueID:        utils.PgtypeToUUID(seat.VenueID).String(),
		SeatCategoryID: utils.PgtypeToUUID(seat.SeatCategoryID).String(),
		RowLabel:       seat.RowLabel,
		SeatNumber:     seat.SeatNumber,
		GridRow:        int(seat.GridRow),
		GridCol:        int(seat.GridCol),
		IsActive:       seat.IsActive,
		CreatedAt:      utils.PgtypeToTime(seat.CreatedAt),
	}, "Seat created successfully")
}

// BatchCreateSeats configures or replaces the full visual seat grid for a venue.
func (h *VenueHandler) BatchCreateSeats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	venueID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid venue ID format")
		return
	}

	var req BatchCreateSeatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	if len(req.Seats) == 0 {
		RespondError(w, http.StatusBadRequest, "EMPTY_SEATS", "seats list cannot be empty")
		return
	}

	// Validate all seats before attempting DB operations
	type parsedSeat struct {
		categoryUUID uuid.UUID
		rowLabel     string
		seatNumber   string
		gridRow      int32
		gridCol      int32
		isActive     bool
	}
	parsed := make([]parsedSeat, 0, len(req.Seats))
	activeCount := 0

	coordMap := make(map[string]bool)
	labelMap := make(map[string]bool)

	for _, s := range req.Seats {
		catUUID, err := uuid.Parse(s.SeatCategoryID)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_CATEGORY_ID", "invalid seat_category_id: "+s.SeatCategoryID)
			return
		}
		rowLabel := strings.TrimSpace(s.RowLabel)
		seatNum := strings.TrimSpace(s.SeatNumber)
		if rowLabel == "" || seatNum == "" || s.GridRow < 1 || s.GridCol < 1 {
			RespondError(w, http.StatusBadRequest, "INVALID_SEAT_DATA", "each seat requires row_label, seat_number, grid_row >= 1, and grid_col >= 1")
			return
		}

		coordKey := strings.Join([]string{string(rune(s.GridRow)), string(rune(s.GridCol))}, ":")
		if coordMap[coordKey] {
			RespondError(w, http.StatusBadRequest, "DUPLICATE_COORDINATES", "duplicate grid coordinates found in batch payload")
			return
		}
		coordMap[coordKey] = true

		labelKey := rowLabel + ":" + seatNum
		if labelMap[labelKey] {
			RespondError(w, http.StatusBadRequest, "DUPLICATE_SEAT_LABEL", "duplicate row_label and seat_number found in batch payload: "+labelKey)
			return
		}
		labelMap[labelKey] = true

		isActive := true
		if s.IsActive != nil {
			isActive = *s.IsActive
		}
		if isActive {
			activeCount++
		}

		parsed = append(parsed, parsedSeat{
			categoryUUID: catUUID,
			rowLabel:     rowLabel,
			seatNumber:   seatNum,
			gridRow:      int32(s.GridRow),
			gridCol:      int32(s.GridCol),
			isActive:     isActive,
		})
	}

	venuePgUUID := utils.UUIDToPgtype(venueID)

	// Execute inside a database transaction if pool is available
	if h.pool != nil {
		tx, err := h.pool.Begin(r.Context())
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to start transaction")
			return
		}
		defer tx.Rollback(r.Context())

		qtx := generated.New(tx)

		if req.Replace {
			if err := qtx.DeleteSeatsByVenueID(r.Context(), venuePgUUID); err != nil {
				RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to clear existing seats")
				return
			}
		}

		for _, ps := range parsed {
			_, err := qtx.CreateSeat(r.Context(), generated.CreateSeatParams{
				VenueID:        venuePgUUID,
				SeatCategoryID: utils.UUIDToPgtype(ps.categoryUUID),
				RowLabel:       ps.rowLabel,
				SeatNumber:     ps.seatNumber,
				GridRow:        ps.gridRow,
				GridCol:        ps.gridCol,
				IsActive:       ps.isActive,
			})
			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
					RespondError(w, http.StatusConflict, "SEAT_EXISTS", "seat already exists with conflicting coordinates or label")
					return
				}
				RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to insert seats")
				return
			}
		}

		// Update total capacity in venue
		v, err := qtx.GetVenueByID(r.Context(), venuePgUUID)
		if err == nil {
			_, _ = qtx.UpdateVenue(r.Context(), generated.UpdateVenueParams{
				ID:            v.ID,
				Name:          v.Name,
				Address:       v.Address,
				City:          v.City,
				State:         v.State,
				Country:       v.Country,
				TotalCapacity: int32(activeCount),
			})
		}

		if err := tx.Commit(r.Context()); err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to commit transaction")
			return
		}
	} else {
		// Non-pool execution (e.g. testing with mockQuerier)
		if req.Replace {
			_ = h.queries.DeleteSeatsByVenueID(r.Context(), venuePgUUID)
		}
		for _, ps := range parsed {
			_, err := h.queries.CreateSeat(r.Context(), generated.CreateSeatParams{
				VenueID:        venuePgUUID,
				SeatCategoryID: utils.UUIDToPgtype(ps.categoryUUID),
				RowLabel:       ps.rowLabel,
				SeatNumber:     ps.seatNumber,
				GridRow:        ps.gridRow,
				GridCol:        ps.gridCol,
				IsActive:       ps.isActive,
			})
			if err != nil {
				RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to insert seat")
				return
			}
		}
	}

	RespondSuccess(w, http.StatusCreated, map[string]interface{}{
		"venue_id":        venueID.String(),
		"total_configured": len(parsed),
		"active_capacity":  activeCount,
	}, "Seat layout successfully configured")
}

// DeleteVenueSeats clears all seats for a venue.
func (h *VenueHandler) DeleteVenueSeats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	venueID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid venue ID format")
		return
	}

	if err := h.queries.DeleteSeatsByVenueID(r.Context(), utils.UUIDToPgtype(venueID)); err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to delete seats")
		return
	}

	RespondSuccess(w, http.StatusOK, nil, "All seats deleted for venue")
}
