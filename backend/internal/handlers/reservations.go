package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ticket-booking/backend/internal/auth"
	"ticket-booking/backend/internal/config"
	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/middleware"
	"ticket-booking/backend/internal/utils"
)

type ReservationHandler struct {
	queries generated.Querier
	pool    *pgxpool.Pool
	cfg     *config.Config
}

func NewReservationHandler(queries generated.Querier, pool *pgxpool.Pool, cfg *config.Config) *ReservationHandler {
	return &ReservationHandler{
		queries: queries,
		pool:    pool,
		cfg:     cfg,
	}
}

// ============================================================================
// DTOs
// ============================================================================

type SeatMapItemResponse struct {
	SeatID         string     `json:"seat_id"`
	VenueID        string     `json:"venue_id"`
	SeatCategoryID string     `json:"seat_category_id"`
	CategoryName   string     `json:"category_name"`
	CategoryColor  string     `json:"category_color"`
	RowLabel       string     `json:"row_label"`
	SeatNumber     string     `json:"seat_number"`
	GridRow        int        `json:"grid_row"`
	GridCol        int        `json:"grid_col"`
	Price          float64    `json:"price"`
	Currency       string     `json:"currency"`
	Status         string     `json:"status"` // AVAILABLE, HELD, OFFERED, BOOKED
	HeldByUserID   *string    `json:"held_by_user_id,omitempty"`
	HoldExpiresAt  *time.Time `json:"hold_expires_at,omitempty"`
	IsMyHold       bool       `json:"is_my_hold"`
}

type HoldSeatsRequest struct {
	SeatIDs []string `json:"seat_ids"`
}

type HeldSeatDetail struct {
	SeatID         string  `json:"seat_id"`
	SeatCategoryID string  `json:"seat_category_id"`
	Price          float64 `json:"price"`
	Currency       string  `json:"currency"`
}

type HoldSeatsResponse struct {
	HoldToken      string           `json:"hold_token"`
	EventID        string           `json:"event_id"`
	ExpiresAt      time.Time        `json:"expires_at"`
	HoldTTLSeconds int              `json:"hold_ttl_seconds"`
	SeatCount      int              `json:"seat_count"`
	TotalPrice     float64          `json:"total_price"`
	Currency       string           `json:"currency"`
	Seats          []HeldSeatDetail `json:"seats"`
}

type ReleaseHoldRequest struct {
	HoldToken string  `json:"hold_token"`
	SeatID    *string `json:"seat_id,omitempty"`
}

// ============================================================================
// HANDLERS
// ============================================================================

// GetEventSeatMap returns visual seat map for an event with real-time seat availability.
func (h *ReservationHandler) GetEventSeatMap(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid event ID format")
		return
	}

	// Optional check for authenticated user
	var currentUserID uuid.UUID
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenStr := authHeader[7:]
		if claims, err := auth.ValidateToken(tokenStr, h.cfg.JWTSecret); err == nil && claims != nil {
			currentUserID = claims.UserID
		}
	}

	rows, err := h.queries.GetEventSeatMapWithStatus(r.Context(), utils.UUIDToPgtype(eventID))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve event seat map")
		return
	}

	res := make([]SeatMapItemResponse, 0, len(rows))
	for _, row := range rows {
		var heldBy *string
		var holdExp *time.Time
		isMyHold := false

		if row.HeldByUserID.Valid {
			uID := utils.PgtypeToUUID(row.HeldByUserID)
			uStr := uID.String()
			heldBy = &uStr
			if currentUserID != uuid.Nil && currentUserID == uID {
				isMyHold = true
			}
		}
		if row.HoldExpiresAt.Valid {
			t := utils.PgtypeToTime(row.HoldExpiresAt)
			holdExp = &t
		}

		res = append(res, SeatMapItemResponse{
			SeatID:         utils.PgtypeToUUID(row.SeatID).String(),
			VenueID:        utils.PgtypeToUUID(row.VenueID).String(),
			SeatCategoryID: utils.PgtypeToUUID(row.SeatCategoryID).String(),
			CategoryName:   row.CategoryName,
			CategoryColor:  row.CategoryColor,
			RowLabel:       row.RowLabel,
			SeatNumber:     row.SeatNumber,
			GridRow:        int(row.GridRow),
			GridCol:        int(row.GridCol),
			Price:          utils.PgtypeNumericToFloat64(row.Price),
			Currency:       row.Currency,
			Status:         row.ComputedStatus,
			HeldByUserID:   heldBy,
			HoldExpiresAt:  holdExp,
			IsMyHold:       isMyHold,
		})
	}

	RespondSuccess(w, http.StatusOK, res, "Event seat map retrieved")
}

// HoldSeats atomically places a temporary hold on the selected seats.
func (h *ReservationHandler) HoldSeats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid event ID format")
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req HoldSeatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	if len(req.SeatIDs) == 0 {
		RespondError(w, http.StatusBadRequest, "EMPTY_SEATS", "please select at least one seat to hold")
		return
	}

	if len(req.SeatIDs) > 10 {
		RespondError(w, http.StatusBadRequest, "MAX_SEATS_EXCEEDED", "cannot hold more than 10 seats at once")
		return
	}

	// Validate seat UUIDs
	seatUUIDs := make([]uuid.UUID, 0, len(req.SeatIDs))
	seen := make(map[uuid.UUID]bool)
	for _, sIDStr := range req.SeatIDs {
		sUUID, err := uuid.Parse(sIDStr)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_SEAT_ID", "invalid seat ID: "+sIDStr)
			return
		}
		if seen[sUUID] {
			RespondError(w, http.StatusBadRequest, "DUPLICATE_SEAT", "duplicate seat ID in request: "+sIDStr)
			return
		}
		seen[sUUID] = true
		seatUUIDs = append(seatUUIDs, sUUID)
	}

	// Fetch event to get configurable hold_ttl_seconds
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
		RespondError(w, http.StatusBadRequest, "EVENT_NOT_AVAILABLE", "seats cannot be held for non-published events")
		return
	}

	if event.StartTime.Valid && event.StartTime.Time.Before(time.Now()) {
		RespondError(w, http.StatusBadRequest, "EVENT_ALREADY_STARTED", "cannot hold seats for an event that has already started")
		return
	}

	holdTTL := event.HoldTtlSeconds
	if holdTTL <= 0 {
		holdTTL = 600
	}
	expiresAt := time.Now().Add(time.Duration(holdTTL) * time.Second)
	holdToken := uuid.New()

	eventPgUUID := utils.UUIDToPgtype(eventID)
	userPgUUID := utils.UUIDToPgtype(userID)
	holdTokenPgUUID := utils.UUIDToPgtype(holdToken)
	expiresAtPgTime := utils.TimeToPgtype(expiresAt)

	var heldSeats []HeldSeatDetail
	var totalPrice float64
	currency := "USD"

	// Transactional hold execution
	if h.pool != nil {
		tx, err := h.pool.Begin(r.Context())
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to start transaction")
			return
		}
		defer tx.Rollback(r.Context())

		qtx := generated.New(tx)

		for _, sUUID := range seatUUIDs {
			seatPgUUID := utils.UUIDToPgtype(sUUID)

			// Clean up any expired holds for this seat first
			_, _ = qtx.ReleaseExpiredSeatHold(r.Context(), generated.ReleaseExpiredSeatHoldParams{
				EventID: eventPgUUID,
				SeatID:  seatPgUUID,
			})

			avail, err := qtx.CheckSeatAvailable(r.Context(), generated.CheckSeatAvailableParams{
				ID:   eventPgUUID,
				ID_2: seatPgUUID,
			})
			if err != nil {
				RespondError(w, http.StatusConflict, "SEAT_UNAVAILABLE", "one or more selected seats are no longer available")
				return
			}

			price := utils.PgtypeNumericToFloat64(avail.Price)
			totalPrice += price
			currency = avail.Currency

			_, err = qtx.CreateSeatReservation(r.Context(), generated.CreateSeatReservationParams{
				EventID:   eventPgUUID,
				SeatID:    seatPgUUID,
				UserID:    userPgUUID,
				Status:    generated.ReservationStatusHELD,
				HoldToken: holdTokenPgUUID,
				ExpiresAt: expiresAtPgTime,
			})
			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
					RespondError(w, http.StatusConflict, "SEAT_UNAVAILABLE", "one or more selected seats were just reserved by another user")
					return
				}
				RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to place seat hold")
				return
			}

			heldSeats = append(heldSeats, HeldSeatDetail{
				SeatID:         sUUID.String(),
				SeatCategoryID: utils.PgtypeToUUID(avail.SeatCategoryID).String(),
				Price:          price,
				Currency:       avail.Currency,
			})
		}

		if err := tx.Commit(r.Context()); err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to commit seat hold")
			return
		}
	} else {
		// Non-pool execution for mock tests
		for _, sUUID := range seatUUIDs {
			seatPgUUID := utils.UUIDToPgtype(sUUID)
			_, _ = h.queries.ReleaseExpiredSeatHold(r.Context(), generated.ReleaseExpiredSeatHoldParams{
				EventID: eventPgUUID,
				SeatID:  seatPgUUID,
			})

			avail, err := h.queries.CheckSeatAvailable(r.Context(), generated.CheckSeatAvailableParams{
				ID:   eventPgUUID,
				ID_2: seatPgUUID,
			})
			if err != nil {
				RespondError(w, http.StatusConflict, "SEAT_UNAVAILABLE", "one or more selected seats are no longer available")
				return
			}

			price := utils.PgtypeNumericToFloat64(avail.Price)
			totalPrice += price
			currency = avail.Currency

			_, err = h.queries.CreateSeatReservation(r.Context(), generated.CreateSeatReservationParams{
				EventID:   eventPgUUID,
				SeatID:    seatPgUUID,
				UserID:    userPgUUID,
				Status:    generated.ReservationStatusHELD,
				HoldToken: holdTokenPgUUID,
				ExpiresAt: expiresAtPgTime,
			})
			if err != nil {
				RespondError(w, http.StatusConflict, "SEAT_UNAVAILABLE", "one or more selected seats are no longer available")
				return
			}

			heldSeats = append(heldSeats, HeldSeatDetail{
				SeatID:         sUUID.String(),
				SeatCategoryID: utils.PgtypeToUUID(avail.SeatCategoryID).String(),
				Price:          price,
				Currency:       avail.Currency,
			})
		}
	}

	RespondSuccess(w, http.StatusCreated, HoldSeatsResponse{
		HoldToken:      holdToken.String(),
		EventID:        eventID.String(),
		ExpiresAt:      expiresAt,
		HoldTTLSeconds: int(holdTTL),
		SeatCount:      len(heldSeats),
		TotalPrice:     totalPrice,
		Currency:       currency,
		Seats:          heldSeats,
	}, "Seats successfully held")
}

// ReleaseHold manually releases seats when checkout is abandoned.
func (h *ReservationHandler) ReleaseHold(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid event ID format")
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req ReleaseHoldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	eventPgUUID := utils.UUIDToPgtype(eventID)

	if req.SeatID != nil && *req.SeatID != "" {
		seatUUID, err := uuid.Parse(*req.SeatID)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_SEAT_ID", "invalid seat ID format")
			return
		}
		_, _ = h.queries.ReleaseSeatHoldBySeatAndUser(r.Context(), generated.ReleaseSeatHoldBySeatAndUserParams{
			EventID: eventPgUUID,
			SeatID:  utils.UUIDToPgtype(seatUUID),
			UserID:  utils.UUIDToPgtype(userID),
		})
	} else if req.HoldToken != "" {
		tokenUUID, err := uuid.Parse(req.HoldToken)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_TOKEN", "invalid hold token format")
			return
		}
		_, _ = h.queries.ReleaseSeatHoldByToken(r.Context(), generated.ReleaseSeatHoldByTokenParams{
			EventID:   eventPgUUID,
			HoldToken: utils.UUIDToPgtype(tokenUUID),
		})
	} else {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "either hold_token or seat_id is required")
		return
	}

	RespondSuccess(w, http.StatusOK, map[string]bool{"released": true}, "Seat hold released successfully")
}
