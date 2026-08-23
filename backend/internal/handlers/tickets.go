package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"ticket-booking/backend/internal/config"
	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/middleware"
	"ticket-booking/backend/internal/utils"
)

type TicketHandler struct {
	queries generated.Querier
	pool    *pgxpool.Pool
	cfg     *config.Config
}

func NewTicketHandler(queries generated.Querier, pool *pgxpool.Pool, cfg *config.Config) *TicketHandler {
	return &TicketHandler{
		queries: queries,
		pool:    pool,
		cfg:     cfg,
	}
}

// ============================================================================
// DTOs
// ============================================================================

type VerifyTicketRequest struct {
	QRPayload string `json:"qr_payload,omitempty"`
	TicketID  string `json:"ticket_id,omitempty"`
}

type CheckInTicketRequest struct {
	QRPayload string `json:"qr_payload,omitempty"`
	TicketID  string `json:"ticket_id,omitempty"`
}

type TicketInfo struct {
	ID            string     `json:"id"`
	Status        string     `json:"status"`
	UnitPrice     float64    `json:"unit_price"`
	QRCodePayload string     `json:"qr_code_payload"`
	CreatedAt     time.Time  `json:"created_at"`
	CheckedInAt   *time.Time `json:"checked_in_at,omitempty"`
}

type TicketBookingInfo struct {
	ID               string `json:"id"`
	BookingReference string `json:"booking_reference"`
	Status           string `json:"status"`
	CustomerID       string `json:"customer_id"`
	CustomerName     string `json:"customer_name"`
	CustomerEmail    string `json:"customer_email"`
}

type TicketEventInfo struct {
	ID          string    `json:"id"`
	OrganiserID string    `json:"organiser_id"`
	Title       string    `json:"title"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
	PosterURL   *string   `json:"poster_url,omitempty"`
}

type TicketVenueInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
}

type TicketSeatInfo struct {
	ID             string `json:"id"`
	RowLabel       string `json:"row_label"`
	SeatNumber     string `json:"seat_number"`
	GridRow        int    `json:"grid_row"`
	GridCol        int    `json:"grid_col"`
	SeatCategoryID string `json:"seat_category_id"`
	CategoryName   string `json:"category_name"`
	CategoryColor  string `json:"category_color"`
}

type TicketVerificationResponse struct {
	IsValid           bool              `json:"is_valid"`
	CanCheckIn        bool              `json:"can_check_in"`
	ValidationMessage string            `json:"validation_message"`
	Ticket            TicketInfo        `json:"ticket"`
	Booking           TicketBookingInfo `json:"booking"`
	Event             TicketEventInfo   `json:"event"`
	Venue             TicketVenueInfo   `json:"venue"`
	Seat              TicketSeatInfo    `json:"seat"`
}

type TicketListItemResponse struct {
	ID               string     `json:"id"`
	BookingID        string     `json:"booking_id"`
	BookingReference string     `json:"booking_reference"`
	CustomerID       string     `json:"customer_id"`
	CustomerName     string     `json:"customer_name"`
	CustomerEmail    string     `json:"customer_email"`
	SeatID           string     `json:"seat_id"`
	RowLabel         string     `json:"row_label"`
	SeatNumber       string     `json:"seat_number"`
	CategoryName     string     `json:"category_name"`
	CategoryColor    string     `json:"category_color"`
	UnitPrice        float64    `json:"unit_price"`
	QRCodePayload    string     `json:"qr_code_payload"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	CheckedInAt      *time.Time `json:"checked_in_at,omitempty"`
}

type EventCheckInOverviewResponse struct {
	EventID        string                   `json:"event_id"`
	TotalTickets   int64                    `json:"total_tickets"`
	ValidCount     int64                    `json:"valid_count"`
	CheckedInCount int64                    `json:"checked_in_count"`
	CancelledCount int64                    `json:"cancelled_count"`
	CheckInRate    float64                  `json:"check_in_rate"`
	Tickets        []TicketListItemResponse `json:"tickets,omitempty"`
}

// ============================================================================
// HANDLERS
// ============================================================================

// VerifyTicket inspects a ticket by QR payload or Ticket ID without mutating state.
func (h *TicketHandler) VerifyTicket(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	userRole, _ := middleware.GetUserRole(r.Context())

	var qrPayload, ticketIDStr string

	if r.Method == http.MethodPost {
		var req VerifyTicketRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		qrPayload = strings.TrimSpace(req.QRPayload)
		ticketIDStr = strings.TrimSpace(req.TicketID)
	}

	// Fallback to query parameters
	if qrPayload == "" {
		qrPayload = strings.TrimSpace(r.URL.Query().Get("payload"))
	}
	if qrPayload == "" {
		qrPayload = strings.TrimSpace(r.URL.Query().Get("qr_payload"))
	}
	if ticketIDStr == "" {
		ticketIDStr = strings.TrimSpace(r.URL.Query().Get("id"))
	}
	if ticketIDStr == "" {
		ticketIDStr = strings.TrimSpace(r.URL.Query().Get("ticket_id"))
	}

	if qrPayload == "" && ticketIDStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "must provide qr_payload or ticket_id")
		return
	}

	var row ticketDetailsCommon
	if ticketIDStr != "" {
		ticketUUID, err := uuid.Parse(ticketIDStr)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_TICKET_ID", "invalid ticket ID format")
			return
		}
		tRow, err := h.queries.GetTicketByID(r.Context(), utils.UUIDToPgtype(ticketUUID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				RespondError(w, http.StatusNotFound, "TICKET_NOT_FOUND", "ticket not found")
				return
			}
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve ticket")
			return
		}
		row = mapGetTicketByIDRow(tRow)
	} else {
		tRow, err := h.queries.GetTicketByQRPayload(r.Context(), qrPayload)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				RespondError(w, http.StatusNotFound, "TICKET_NOT_FOUND", "ticket with specified QR payload not found")
				return
			}
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve ticket")
			return
		}
		row = mapGetTicketByQRPayloadRow(tRow)
	}

	// Authorization Check
	if userRole == "CUSTOMER" {
		if utils.PgtypeToUUID(row.CustomerID) != userID {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "access denied: you do not own this ticket")
			return
		}
	} else if userRole == "ORGANISER" {
		if utils.PgtypeToUUID(row.OrganiserID) != userID {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "access denied: you are not the organiser for this event")
			return
		}
	}

	resp := buildVerificationResponse(row)
	RespondSuccess(w, http.StatusOK, resp, "Ticket verified successfully")
}

// CheckInTicket atomically transitions ticket status from VALID to CHECKED_IN.
func (h *TicketHandler) CheckInTicket(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	userRole, _ := middleware.GetUserRole(r.Context())

	// Only Organiser or Admin may perform gate check-in
	if userRole != "ORGANISER" && userRole != "ADMIN" {
		RespondError(w, http.StatusForbidden, "FORBIDDEN", "only event organisers and admins can perform check-in")
		return
	}

	var qrPayload, ticketIDStr string
	// Check URL param first
	if urlID := chi.URLParam(r, "id"); urlID != "" {
		ticketIDStr = strings.TrimSpace(urlID)
	}

	if r.Body != nil {
		var req CheckInTicketRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.QRPayload != "" {
			qrPayload = strings.TrimSpace(req.QRPayload)
		}
		if req.TicketID != "" && ticketIDStr == "" {
			ticketIDStr = strings.TrimSpace(req.TicketID)
		}
	}

	if qrPayload == "" && ticketIDStr == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "must provide qr_payload or ticket_id")
		return
	}

	// 1. Fetch current ticket state to verify event ownership and check-in eligibility
	var row ticketDetailsCommon
	var ticketPgUUID pgtype.UUID

	if ticketIDStr != "" {
		ticketUUID, err := uuid.Parse(ticketIDStr)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "INVALID_TICKET_ID", "invalid ticket ID format")
			return
		}
		ticketPgUUID = utils.UUIDToPgtype(ticketUUID)
		tRow, err := h.queries.GetTicketByID(r.Context(), ticketPgUUID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				RespondError(w, http.StatusNotFound, "TICKET_NOT_FOUND", "ticket not found")
				return
			}
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve ticket")
			return
		}
		row = mapGetTicketByIDRow(tRow)
	} else {
		tRow, err := h.queries.GetTicketByQRPayload(r.Context(), qrPayload)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				RespondError(w, http.StatusNotFound, "TICKET_NOT_FOUND", "ticket with specified QR payload not found")
				return
			}
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve ticket")
			return
		}
		row = mapGetTicketByQRPayloadRow(tRow)
		ticketPgUUID = row.ID
	}

	// Verify organiser ownership
	if userRole == "ORGANISER" && utils.PgtypeToUUID(row.OrganiserID) != userID {
		RespondError(w, http.StatusForbidden, "FORBIDDEN", "access denied: you are not the organiser for this event")
		return
	}

	// Validate ticket status
	if row.Status == generated.TicketStatusCHECKEDIN {
		checkInTimeStr := "previously"
		if row.CheckedInAt.Valid {
			checkInTimeStr = row.CheckedInAt.Time.Format(time.RFC3339)
		}
		RespondError(w, http.StatusConflict, "ALREADY_CHECKED_IN", fmt.Sprintf("ticket was already checked in (%s)", checkInTimeStr))
		return
	}

	if row.Status == generated.TicketStatusCANCELLED {
		RespondError(w, http.StatusBadRequest, "TICKET_CANCELLED", "cannot check in a cancelled ticket")
		return
	}

	if row.BookingStatus != generated.BookingStatusCONFIRMED {
		RespondError(w, http.StatusBadRequest, "BOOKING_CANCELLED", "cannot check in ticket: booking is cancelled")
		return
	}

	if row.EventStatus == generated.EventStatusCANCELLED {
		RespondError(w, http.StatusBadRequest, "EVENT_CANCELLED", "cannot check in ticket: event has been cancelled")
		return
	}

	// 2. Perform atomic check-in state transition
	var updatedTicket generated.Ticket
	if ticketIDStr != "" {
		updatedTicket, err = h.queries.CheckInTicketByID(r.Context(), ticketPgUUID)
	} else {
		updatedTicket, err = h.queries.CheckInTicketByQRPayload(r.Context(), qrPayload)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusConflict, "ALREADY_CHECKED_IN", "ticket was checked in by another scanner")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to update check-in status")
		return
	}

	// Update local struct representation for response
	row.Status = updatedTicket.Status
	row.CheckedInAt = updatedTicket.CheckedInAt

	resp := buildVerificationResponse(row)
	resp.ValidationMessage = "Ticket checked in successfully"
	RespondSuccess(w, http.StatusOK, resp, "Ticket checked in successfully")
}

// ListEventTickets returns all tickets and gate check-in statistics for an event (Organiser/Admin).
func (h *TicketHandler) ListEventTickets(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "id")
	eventUUID, err := uuid.Parse(eventIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_EVENT_ID", "invalid event ID format")
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}
	userRole, _ := middleware.GetUserRole(r.Context())

	eventPgUUID := utils.UUIDToPgtype(eventUUID)

	// Verify Event and Organiser permissions
	event, err := h.queries.GetEventByID(r.Context(), eventPgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "event not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve event")
		return
	}

	if userRole == "ORGANISER" && utils.PgtypeToUUID(event.OrganiserID) != userID {
		RespondError(w, http.StatusForbidden, "FORBIDDEN", "access denied: you are not the organiser for this event")
		return
	}

	stats, err := h.queries.GetEventCheckInStats(r.Context(), eventPgUUID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve check-in statistics")
		return
	}

	tickets, err := h.queries.ListTicketsByEventID(r.Context(), eventPgUUID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to list event tickets")
		return
	}

	statusFilter := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("status")))

	items := make([]TicketListItemResponse, 0, len(tickets))
	for _, t := range tickets {
		if statusFilter != "" && string(t.Status) != statusFilter {
			continue
		}

		var chkIn *time.Time
		if t.CheckedInAt.Valid {
			tVal := utils.PgtypeToTime(t.CheckedInAt)
			chkIn = &tVal
		}

		items = append(items, TicketListItemResponse{
			ID:               utils.PgtypeToUUID(t.ID).String(),
			BookingID:        utils.PgtypeToUUID(t.BookingID).String(),
			BookingReference: t.BookingReference,
			CustomerID:       utils.PgtypeToUUID(t.CustomerID).String(),
			CustomerName:     t.CustomerName,
			CustomerEmail:    t.CustomerEmail,
			SeatID:           utils.PgtypeToUUID(t.SeatID).String(),
			RowLabel:         t.RowLabel,
			SeatNumber:       t.SeatNumber,
			CategoryName:     t.CategoryName,
			CategoryColor:    t.CategoryColor,
			UnitPrice:        utils.PgtypeNumericToFloat64(t.UnitPrice),
			QRCodePayload:    t.QrCodePayload,
			Status:           string(t.Status),
			CreatedAt:        utils.PgtypeToTime(t.CreatedAt),
			CheckedInAt:      chkIn,
		})
	}

	var checkInRate float64
	attendedTotal := stats.ValidCount + stats.CheckedInCount
	if attendedTotal > 0 {
		checkInRate = (float64(stats.CheckedInCount) / float64(attendedTotal)) * 100
	}

	overview := EventCheckInOverviewResponse{
		EventID:        eventUUID.String(),
		TotalTickets:   stats.TotalTickets,
		ValidCount:     stats.ValidCount,
		CheckedInCount: stats.CheckedInCount,
		CancelledCount: stats.CancelledCount,
		CheckInRate:    checkInRate,
		Tickets:        items,
	}

	RespondSuccess(w, http.StatusOK, overview, "Event tickets and check-in overview retrieved")
}

// ============================================================================
// INTERNAL MAPPING HELPERS
// ============================================================================

type ticketDetailsCommon struct {
	ID             pgtype.UUID
	BookingID      pgtype.UUID
	EventID        pgtype.UUID
	SeatID         pgtype.UUID
	UnitPrice      pgtype.Numeric
	QrCodePayload  string
	Status         generated.TicketStatus
	CreatedAt      pgtype.Timestamptz
	CheckedInAt    pgtype.Timestamptz
	BookingRef     string
	BookingStatus  generated.BookingStatus
	CustomerID     pgtype.UUID
	CustomerName   string
	CustomerEmail  string
	OrganiserID    pgtype.UUID
	EventTitle     string
	EventStartTime pgtype.Timestamptz
	EventEndTime   pgtype.Timestamptz
	EventStatus    generated.EventStatus
	EventPosterURL pgtype.Text
	VenueID        pgtype.UUID
	VenueName      string
	VenueAddress   string
	VenueCity      string
	RowLabel       string
	SeatNumber     string
	GridRow        int32
	GridCol        int32
	SeatCategoryID pgtype.UUID
	CategoryName   string
	CategoryColor  string
}

func mapGetTicketByIDRow(r generated.GetTicketByIDRow) ticketDetailsCommon {
	return ticketDetailsCommon{
		ID:             r.ID,
		BookingID:      r.BookingID,
		EventID:        r.EventID,
		SeatID:         r.SeatID,
		UnitPrice:      r.UnitPrice,
		QrCodePayload:  r.QrCodePayload,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt,
		CheckedInAt:    r.CheckedInAt,
		BookingRef:     r.BookingReference,
		BookingStatus:  r.BookingStatus,
		CustomerID:     r.CustomerID,
		CustomerName:   r.CustomerName,
		CustomerEmail:  r.CustomerEmail,
		OrganiserID:    r.OrganiserID,
		EventTitle:     r.EventTitle,
		EventStartTime: r.EventStartTime,
		EventEndTime:   r.EventEndTime,
		EventStatus:    r.EventStatus,
		EventPosterURL: r.EventPosterUrl,
		VenueID:        r.VenueID,
		VenueName:      r.VenueName,
		VenueAddress:   r.VenueAddress,
		VenueCity:      r.VenueCity,
		RowLabel:       r.RowLabel,
		SeatNumber:     r.SeatNumber,
		GridRow:        r.GridRow,
		GridCol:        r.GridCol,
		SeatCategoryID: r.SeatCategoryID,
		CategoryName:   r.CategoryName,
		CategoryColor:  r.CategoryColor,
	}
}

func mapGetTicketByQRPayloadRow(r generated.GetTicketByQRPayloadRow) ticketDetailsCommon {
	return ticketDetailsCommon{
		ID:             r.ID,
		BookingID:      r.BookingID,
		EventID:        r.EventID,
		SeatID:         r.SeatID,
		UnitPrice:      r.UnitPrice,
		QrCodePayload:  r.QrCodePayload,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt,
		CheckedInAt:    r.CheckedInAt,
		BookingRef:     r.BookingReference,
		BookingStatus:  r.BookingStatus,
		CustomerID:     r.CustomerID,
		CustomerName:   r.CustomerName,
		CustomerEmail:  r.CustomerEmail,
		OrganiserID:    r.OrganiserID,
		EventTitle:     r.EventTitle,
		EventStartTime: r.EventStartTime,
		EventEndTime:   r.EventEndTime,
		EventStatus:    r.EventStatus,
		EventPosterURL: r.EventPosterUrl,
		VenueID:        r.VenueID,
		VenueName:      r.VenueName,
		VenueAddress:   r.VenueAddress,
		VenueCity:      r.VenueCity,
		RowLabel:       r.RowLabel,
		SeatNumber:     r.SeatNumber,
		GridRow:        r.GridRow,
		GridCol:        r.GridCol,
		SeatCategoryID: r.SeatCategoryID,
		CategoryName:   r.CategoryName,
		CategoryColor:  r.CategoryColor,
	}
}

func buildVerificationResponse(r ticketDetailsCommon) TicketVerificationResponse {
	var poster *string
	if r.EventPosterURL.Valid && r.EventPosterURL.String != "" {
		poster = &r.EventPosterURL.String
	}

	var checkedInTime *time.Time
	if r.CheckedInAt.Valid {
		t := utils.PgtypeToTime(r.CheckedInAt)
		checkedInTime = &t
	}

	isValid := false
	canCheckIn := false
	var msg string

	if r.Status == generated.TicketStatusVALID {
		if r.BookingStatus != generated.BookingStatusCONFIRMED {
			msg = "Booking is cancelled"
		} else if r.EventStatus == generated.EventStatusCANCELLED {
			msg = "Event has been cancelled"
		} else {
			isValid = true
			canCheckIn = true
			msg = "Ticket is valid and ready for entry"
		}
	} else if r.Status == generated.TicketStatusCHECKEDIN {
		if checkedInTime != nil {
			msg = fmt.Sprintf("Ticket has already been checked in at %s", checkedInTime.Format(time.RFC3339))
		} else {
			msg = "Ticket has already been checked in"
		}
	} else if r.Status == generated.TicketStatusCANCELLED {
		msg = "Ticket has been cancelled"
	}

	return TicketVerificationResponse{
		IsValid:           isValid,
		CanCheckIn:        canCheckIn,
		ValidationMessage: msg,
		Ticket: TicketInfo{
			ID:            utils.PgtypeToUUID(r.ID).String(),
			Status:        string(r.Status),
			UnitPrice:     utils.PgtypeNumericToFloat64(r.UnitPrice),
			QRCodePayload: r.QrCodePayload,
			CreatedAt:     utils.PgtypeToTime(r.CreatedAt),
			CheckedInAt:   checkedInTime,
		},
		Booking: TicketBookingInfo{
			ID:               utils.PgtypeToUUID(r.BookingID).String(),
			BookingReference: r.BookingRef,
			Status:           string(r.BookingStatus),
			CustomerID:       utils.PgtypeToUUID(r.CustomerID).String(),
			CustomerName:     r.CustomerName,
			CustomerEmail:    r.CustomerEmail,
		},
		Event: TicketEventInfo{
			ID:          utils.PgtypeToUUID(r.EventID).String(),
			OrganiserID: utils.PgtypeToUUID(r.OrganiserID).String(),
			Title:       r.EventTitle,
			StartTime:   utils.PgtypeToTime(r.EventStartTime),
			EndTime:     utils.PgtypeToTime(r.EventEndTime),
			Status:      string(r.EventStatus),
			PosterURL:   poster,
		},
		Venue: TicketVenueInfo{
			ID:      utils.PgtypeToUUID(r.VenueID).String(),
			Name:    r.VenueName,
			Address: r.VenueAddress,
			City:    r.VenueCity,
		},
		Seat: TicketSeatInfo{
			ID:             utils.PgtypeToUUID(r.SeatID).String(),
			RowLabel:       r.RowLabel,
			SeatNumber:     r.SeatNumber,
			GridRow:        int(r.GridRow),
			GridCol:        int(r.GridCol),
			SeatCategoryID: utils.PgtypeToUUID(r.SeatCategoryID).String(),
			CategoryName:   r.CategoryName,
			CategoryColor:  r.CategoryColor,
		},
	}
}
