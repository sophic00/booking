package handlers

import (
	"crypto/rand"
	"encoding/hex"
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
	"ticket-booking/backend/internal/qrcode"
	"ticket-booking/backend/internal/utils"
)

type BookingHandler struct {
	queries generated.Querier
	pool    *pgxpool.Pool
	cfg     *config.Config
}

func NewBookingHandler(queries generated.Querier, pool *pgxpool.Pool, cfg *config.Config) *BookingHandler {
	return &BookingHandler{
		queries: queries,
		pool:    pool,
		cfg:     cfg,
	}
}

// ============================================================================
// DTOs
// ============================================================================

type CheckoutRequest struct {
	EventID   string `json:"event_id"`
	HoldToken string `json:"hold_token"`
}

type TicketDetailResponse struct {
	ID             string  `json:"id"`
	BookingID      string  `json:"booking_id"`
	SeatID         string  `json:"seat_id"`
	RowLabel       string  `json:"row_label"`
	SeatNumber     string  `json:"seat_number"`
	GridRow        int     `json:"grid_row"`
	GridCol        int     `json:"grid_col"`
	SeatCategoryID string  `json:"seat_category_id"`
	CategoryName   string  `json:"category_name"`
	CategoryColor  string  `json:"category_color"`
	UnitPrice      float64 `json:"unit_price"`
	QRCodePayload  string  `json:"qr_code_payload"`
	QRCodeDataURL  string  `json:"qr_code_data_url"`
	Status         string  `json:"status"`
}

type BookingResponse struct {
	ID                 string                 `json:"id"`
	BookingReference   string                 `json:"booking_reference"`
	CustomerID         string                 `json:"customer_id"`
	CustomerName       *string                `json:"customer_name,omitempty"`
	CustomerEmail      *string                `json:"customer_email,omitempty"`
	EventID            string                 `json:"event_id"`
	EventTitle         *string                `json:"event_title,omitempty"`
	EventStartTime     *time.Time             `json:"event_start_time,omitempty"`
	VenueName          *string                `json:"venue_name,omitempty"`
	VenueCity          *string                `json:"venue_city,omitempty"`
	TotalAmount        float64                `json:"total_amount"`
	Currency           string                 `json:"currency"`
	Status             string                 `json:"status"`
	CancellationReason *string                `json:"cancellation_reason,omitempty"`
	CancelledAt        *time.Time             `json:"cancelled_at,omitempty"`
	TicketCount        int                    `json:"ticket_count,omitempty"`
	Tickets            []TicketDetailResponse `json:"tickets,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
}

type CancelBookingRequest struct {
	Reason *string `json:"reason,omitempty"`
}

// ============================================================================
// HELPERS
// ============================================================================

func generateBookingReference() string {
	bytes := make([]byte, 4)
	_, _ = rand.Read(bytes)
	randomHex := strings.ToUpper(hex.EncodeToString(bytes))
	return fmt.Sprintf("TB-%s-%s", time.Now().Format("20060102"), randomHex)
}

// ============================================================================
// HANDLERS
// ============================================================================

// Checkout confirms an active seat hold into a permanent booking with tickets and QR codes.
func (h *BookingHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	customerID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	eventUUID, err := uuid.Parse(req.EventID)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_EVENT_ID", "invalid event ID format")
		return
	}

	holdTokenUUID, err := uuid.Parse(req.HoldToken)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_HOLD_TOKEN", "invalid hold token format")
		return
	}

	eventPgUUID := utils.UUIDToPgtype(eventUUID)
	holdTokenPgUUID := utils.UUIDToPgtype(holdTokenUUID)
	customerPgUUID := utils.UUIDToPgtype(customerID)

	// Fetch active hold items
	heldSeats, err := h.queries.GetActiveHoldByToken(r.Context(), generated.GetActiveHoldByTokenParams{
		EventID:   eventPgUUID,
		HoldToken: holdTokenPgUUID,
	})
	if err != nil || len(heldSeats) == 0 {
		RespondError(w, http.StatusBadRequest, "HOLD_EXPIRED", "your seat hold has expired or does not exist")
		return
	}

	// Verify that held seats belong to the requesting user
	for _, hs := range heldSeats {
		if hs.UserID.Valid && utils.PgtypeToUUID(hs.UserID) != customerID {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "this seat hold belongs to another account")
			return
		}
	}

	// Calculate total amount
	var totalAmount float64
	currency := "USD"
	for _, hs := range heldSeats {
		price := utils.PgtypeNumericToFloat64(hs.Price)
		totalAmount += price
		if hs.Currency != "" {
			currency = hs.Currency
		}
	}

	bookingRef := generateBookingReference()
	var createdBooking generated.Booking
	var createdTickets []TicketDetailResponse

	if h.pool != nil {
		tx, err := h.pool.Begin(r.Context())
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to start checkout transaction")
			return
		}
		defer tx.Rollback(r.Context())

		qtx := generated.New(tx)

		// Create booking record
		createdBooking, err = qtx.CreateBooking(r.Context(), generated.CreateBookingParams{
			BookingReference: bookingRef,
			CustomerID:       customerPgUUID,
			EventID:          eventPgUUID,
			TotalAmount:      utils.Float64ToPgtypeNumeric(totalAmount),
			Currency:         currency,
		})
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to create booking record")
			return
		}

		// Create ticket records
		for _, hs := range heldSeats {
			ticketUUID := uuid.New()
			seatIDStr := utils.PgtypeToUUID(hs.SeatID).String()
			qrPayload := qrcode.GenerateTicketPayload(bookingRef, seatIDStr, ticketUUID.String())

			tRecord, err := qtx.CreateTicket(r.Context(), generated.CreateTicketParams{
				BookingID:     createdBooking.ID,
				EventID:       eventPgUUID,
				SeatID:        hs.SeatID,
				UnitPrice:     hs.Price,
				QrCodePayload: qrPayload,
			})
			if err != nil {
				RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to generate ticket")
				return
			}

			qrDataURL, _ := qrcode.GenerateDataURL(qrPayload, 200)

			createdTickets = append(createdTickets, TicketDetailResponse{
				ID:             utils.PgtypeToUUID(tRecord.ID).String(),
				BookingID:      utils.PgtypeToUUID(createdBooking.ID).String(),
				SeatID:         seatIDStr,
				RowLabel:       hs.RowLabel,
				SeatNumber:     hs.SeatNumber,
				SeatCategoryID: utils.PgtypeToUUID(hs.SeatCategoryID).String(),
				CategoryName:   hs.CategoryName,
				CategoryColor:  hs.CategoryColor,
				UnitPrice:      utils.PgtypeNumericToFloat64(hs.Price),
				QRCodePayload:  qrPayload,
				QRCodeDataURL:  qrDataURL,
				Status:         string(tRecord.Status),
			})
		}

		// Confirm reservations
		_, err = qtx.ConfirmReservationToBooked(r.Context(), generated.ConfirmReservationToBookedParams{
			EventID:   eventPgUUID,
			HoldToken: holdTokenPgUUID,
			BookingID: createdBooking.ID,
		})
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to confirm reservations")
			return
		}

		if err := tx.Commit(r.Context()); err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to commit checkout")
			return
		}
	} else {
		// Mock Querier Execution
		createdBooking, err = h.queries.CreateBooking(r.Context(), generated.CreateBookingParams{
			BookingReference: bookingRef,
			CustomerID:       customerPgUUID,
			EventID:          eventPgUUID,
			TotalAmount:      utils.Float64ToPgtypeNumeric(totalAmount),
			Currency:         currency,
		})
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to create booking record")
			return
		}

		for _, hs := range heldSeats {
			ticketUUID := uuid.New()
			seatIDStr := utils.PgtypeToUUID(hs.SeatID).String()
			qrPayload := qrcode.GenerateTicketPayload(bookingRef, seatIDStr, ticketUUID.String())

			tRecord, err := h.queries.CreateTicket(r.Context(), generated.CreateTicketParams{
				BookingID:     createdBooking.ID,
				EventID:       eventPgUUID,
				SeatID:        hs.SeatID,
				UnitPrice:     hs.Price,
				QrCodePayload: qrPayload,
			})
			if err != nil {
				RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to generate ticket")
				return
			}

			qrDataURL, _ := qrcode.GenerateDataURL(qrPayload, 200)

			createdTickets = append(createdTickets, TicketDetailResponse{
				ID:             utils.PgtypeToUUID(tRecord.ID).String(),
				BookingID:      utils.PgtypeToUUID(createdBooking.ID).String(),
				SeatID:         seatIDStr,
				RowLabel:       hs.RowLabel,
				SeatNumber:     hs.SeatNumber,
				SeatCategoryID: utils.PgtypeToUUID(hs.SeatCategoryID).String(),
				CategoryName:   hs.CategoryName,
				CategoryColor:  hs.CategoryColor,
				UnitPrice:      utils.PgtypeNumericToFloat64(hs.Price),
				QRCodePayload:  qrPayload,
				QRCodeDataURL:  qrDataURL,
				Status:         string(tRecord.Status),
			})
		}

		_, _ = h.queries.ConfirmReservationToBooked(r.Context(), generated.ConfirmReservationToBookedParams{
			EventID:   eventPgUUID,
			HoldToken: holdTokenPgUUID,
			BookingID: createdBooking.ID,
		})
	}

	RespondSuccess(w, http.StatusCreated, BookingResponse{
		ID:               utils.PgtypeToUUID(createdBooking.ID).String(),
		BookingReference: createdBooking.BookingReference,
		CustomerID:       customerID.String(),
		EventID:          eventUUID.String(),
		TotalAmount:      totalAmount,
		Currency:         currency,
		Status:           string(createdBooking.Status),
		TicketCount:      len(createdTickets),
		Tickets:          createdTickets,
		CreatedAt:        utils.PgtypeToTime(createdBooking.CreatedAt),
	}, "Booking confirmed successfully")
}

// ListCustomerBookings retrieves the authenticated customer's booking history.
func (h *BookingHandler) ListCustomerBookings(w http.ResponseWriter, r *http.Request) {
	customerID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	bookings, err := h.queries.GetCustomerBookings(r.Context(), utils.UUIDToPgtype(customerID))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve bookings")
		return
	}

	res := make([]BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		var eStart *time.Time
		if b.EventStartTime.Valid {
			t := utils.PgtypeToTime(b.EventStartTime)
			eStart = &t
		}
		var reason *string
		if b.CancellationReason.Valid && b.CancellationReason.String != "" {
			reason = &b.CancellationReason.String
		}
		var cancelledAt *time.Time
		if b.CancelledAt.Valid {
			t := utils.PgtypeToTime(b.CancelledAt)
			cancelledAt = &t
		}

		res = append(res, BookingResponse{
			ID:                 utils.PgtypeToUUID(b.ID).String(),
			BookingReference:   b.BookingReference,
			CustomerID:         utils.PgtypeToUUID(b.CustomerID).String(),
			EventID:            utils.PgtypeToUUID(b.EventID).String(),
			EventTitle:         &b.EventTitle,
			EventStartTime:     eStart,
			VenueName:          &b.VenueName,
			VenueCity:          &b.VenueCity,
			TotalAmount:        utils.PgtypeNumericToFloat64(b.TotalAmount),
			Currency:           b.Currency,
			Status:             string(b.Status),
			CancellationReason: reason,
			CancelledAt:        cancelledAt,
			TicketCount:        int(b.TicketCount),
			CreatedAt:          utils.PgtypeToTime(b.CreatedAt),
		})
	}

	RespondSuccess(w, http.StatusOK, res, "Customer bookings retrieved")
}

// GetBooking retrieves single booking details with QR codes.
func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	bookingID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid booking ID format")
		return
	}

	customerID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	booking, err := h.queries.GetBookingByID(r.Context(), utils.UUIDToPgtype(bookingID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "BOOKING_NOT_FOUND", "booking not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve booking")
		return
	}

	userRole, _ := middleware.GetUserRole(r.Context())
	if utils.PgtypeToUUID(booking.CustomerID) != customerID && userRole != "ADMIN" {
		RespondError(w, http.StatusForbidden, "FORBIDDEN", "access denied to this booking")
		return
	}

	tickets, err := h.queries.GetTicketsByBookingID(r.Context(), utils.UUIDToPgtype(bookingID))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve tickets")
		return
	}

	ticketDetails := make([]TicketDetailResponse, 0, len(tickets))
	for _, t := range tickets {
		qrDataURL, _ := qrcode.GenerateDataURL(t.QrCodePayload, 200)

		ticketDetails = append(ticketDetails, TicketDetailResponse{
			ID:             utils.PgtypeToUUID(t.ID).String(),
			BookingID:      utils.PgtypeToUUID(t.BookingID).String(),
			SeatID:         utils.PgtypeToUUID(t.SeatID).String(),
			RowLabel:       t.RowLabel,
			SeatNumber:     t.SeatNumber,
			GridRow:        int(t.GridRow),
			GridCol:        int(t.GridCol),
			SeatCategoryID: utils.PgtypeToUUID(t.SeatCategoryID).String(),
			CategoryName:   t.CategoryName,
			CategoryColor:  t.CategoryColor,
			UnitPrice:      utils.PgtypeNumericToFloat64(t.UnitPrice),
			QRCodePayload:  t.QrCodePayload,
			QRCodeDataURL:  qrDataURL,
			Status:         string(t.Status),
		})
	}

	var eStart *time.Time
	if booking.EventStartTime.Valid {
		t := utils.PgtypeToTime(booking.EventStartTime)
		eStart = &t
	}
	var reason *string
	if booking.CancellationReason.Valid && booking.CancellationReason.String != "" {
		reason = &booking.CancellationReason.String
	}
	var cancelledAt *time.Time
	if booking.CancelledAt.Valid {
		t := utils.PgtypeToTime(booking.CancelledAt)
		cancelledAt = &t
	}

	RespondSuccess(w, http.StatusOK, BookingResponse{
		ID:                 utils.PgtypeToUUID(booking.ID).String(),
		BookingReference:   booking.BookingReference,
		CustomerID:         utils.PgtypeToUUID(booking.CustomerID).String(),
		CustomerName:       &booking.CustomerName,
		CustomerEmail:      &booking.CustomerEmail,
		EventID:            utils.PgtypeToUUID(booking.EventID).String(),
		EventTitle:         &booking.EventTitle,
		EventStartTime:     eStart,
		VenueName:          &booking.VenueName,
		VenueCity:          &booking.VenueCity,
		TotalAmount:        utils.PgtypeNumericToFloat64(booking.TotalAmount),
		Currency:           booking.Currency,
		Status:             string(booking.Status),
		CancellationReason: reason,
		CancelledAt:        cancelledAt,
		TicketCount:        len(ticketDetails),
		Tickets:            ticketDetails,
		CreatedAt:          utils.PgtypeToTime(booking.CreatedAt),
	}, "Booking details retrieved")
}

// CancelBooking handles customer cancellation of a confirmed booking.
func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	bookingID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid booking ID format")
		return
	}

	customerID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req CancelBookingRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	reason := "Customer requested cancellation"
	if req.Reason != nil && strings.TrimSpace(*req.Reason) != "" {
		reason = strings.TrimSpace(*req.Reason)
	}

	bookingPgUUID := utils.UUIDToPgtype(bookingID)
	customerPgUUID := utils.UUIDToPgtype(customerID)

	var cancelledBooking generated.Booking

	if h.pool != nil {
		tx, err := h.pool.Begin(r.Context())
		if err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to start cancellation transaction")
			return
		}
		defer tx.Rollback(r.Context())

		qtx := generated.New(tx)

		cancelledBooking, err = qtx.CancelBooking(r.Context(), generated.CancelBookingParams{
			ID:                 bookingPgUUID,
			CancellationReason: pgtype.Text{String: reason, Valid: true},
			CustomerID:         customerPgUUID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				RespondError(w, http.StatusNotFound, "BOOKING_NOT_FOUND", "confirmed booking not found or already cancelled")
				return
			}
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to cancel booking")
			return
		}

		// Void tickets and reservations atomically
		if _, err := qtx.CancelBookingTickets(r.Context(), bookingPgUUID); err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to cancel tickets")
			return
		}
		if _, err := qtx.CancelBookingReservations(r.Context(), bookingPgUUID); err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to cancel reservations")
			return
		}

		if err := tx.Commit(r.Context()); err != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to commit cancellation")
			return
		}
	} else {
		var err error
		cancelledBooking, err = h.queries.CancelBooking(r.Context(), generated.CancelBookingParams{
			ID:                 bookingPgUUID,
			CancellationReason: pgtype.Text{String: reason, Valid: true},
			CustomerID:         customerPgUUID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				RespondError(w, http.StatusNotFound, "BOOKING_NOT_FOUND", "confirmed booking not found or already cancelled")
				return
			}
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to cancel booking")
			return
		}

		// Void tickets and reservations
		_, _ = h.queries.CancelBookingTickets(r.Context(), bookingPgUUID)
		_, _ = h.queries.CancelBookingReservations(r.Context(), bookingPgUUID)
	}

	var cancelledAt *time.Time
	if cancelledBooking.CancelledAt.Valid {
		t := utils.PgtypeToTime(cancelledBooking.CancelledAt)
		cancelledAt = &t
	}

	RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"booking_id":        bookingID.String(),
		"booking_reference": cancelledBooking.BookingReference,
		"status":            string(cancelledBooking.Status),
		"cancelled_at":      cancelledAt,
	}, "Booking cancelled successfully")
}
