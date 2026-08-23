package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"ticket-booking/backend/internal/config"
	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/email"
	"ticket-booking/backend/internal/middleware"
	"ticket-booking/backend/internal/qrcode"
	"ticket-booking/backend/internal/utils"
)

// WaitlistHandler implements the sold-out waitlist and time-limited offer flow.
type WaitlistHandler struct {
	queries generated.Querier
	pool    *pgxpool.Pool
	cfg     *config.Config
	mailer  *email.Service
}

func NewWaitlistHandler(queries generated.Querier, pool *pgxpool.Pool, cfg *config.Config, mailer *email.Service) *WaitlistHandler {
	return &WaitlistHandler{queries: queries, pool: pool, cfg: cfg, mailer: mailer}
}

// ============================================================================
// DTOs
// ============================================================================

type JoinWaitlistRequest struct {
	SeatCategoryID string `json:"seat_category_id"`
}

type WaitlistEntryResponse struct {
	ID             string     `json:"id"`
	EventID        string     `json:"event_id"`
	EventTitle     *string    `json:"event_title,omitempty"`
	SeatCategoryID string     `json:"seat_category_id"`
	CategoryName   *string    `json:"category_name,omitempty"`
	CategoryColor  *string    `json:"category_color,omitempty"`
	Status         string     `json:"status"`
	QueuePosition  int        `json:"queue_position"`
	OfferToken     *string    `json:"offer_token,omitempty"`
	OfferExpiresAt *time.Time `json:"offer_expires_at,omitempty"`
	EventStartTime *time.Time `json:"event_start_time,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type WaitlistOfferDetailResponse struct {
	ID              string     `json:"id"`
	WaitlistEntryID string     `json:"waitlist_entry_id"`
	EventID         string     `json:"event_id"`
	EventTitle      string     `json:"event_title"`
	EventStartTime  *time.Time `json:"event_start_time,omitempty"`
	EventEndTime    *time.Time `json:"event_end_time,omitempty"`
	SeatID          string     `json:"seat_id"`
	RowLabel        string     `json:"row_label"`
	SeatNumber      string     `json:"seat_number"`
	SeatCategoryID  string     `json:"seat_category_id"`
	CategoryName    string     `json:"category_name"`
	Price           float64    `json:"price"`
	Currency        string     `json:"currency"`
	OfferToken      string     `json:"offer_token"`
	OfferedAt       time.Time  `json:"offered_at"`
	ExpiresAt       time.Time  `json:"expires_at"`
	Status          string     `json:"status"`
}

// ============================================================================
// HANDLERS
// ============================================================================

// JoinWaitlist adds the authenticated customer to the FIFO waitlist for a
// specific seat category of a sold-out event.
func (h *WaitlistHandler) JoinWaitlist(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid event ID format")
		return
	}

	customerID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	var req JoinWaitlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SeatCategoryID == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "seat_category_id is required")
		return
	}
	categoryID, err := uuid.Parse(req.SeatCategoryID)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_CATEGORY_ID", "invalid seat category ID format")
		return
	}

	eventPgUUID := utils.UUIDToPgtype(eventID)
	categoryPgUUID := utils.UUIDToPgtype(categoryID)

	event, err := h.queries.GetEventByID(r.Context(), eventPgUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "event not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve event")
		return
	}
	if event.Status != generated.EventStatusPUBLISHED {
		RespondError(w, http.StatusBadRequest, "EVENT_NOT_AVAILABLE", "waitlist is only open for published events")
		return
	}

	// The waitlist is only for sold-out categories. If seats are still freely
	// available, point the customer at the normal booking flow instead.
	available, err := h.queries.CountAvailableSeatsInCategory(r.Context(), generated.CountAvailableSeatsInCategoryParams{
		ID:             eventPgUUID,
		SeatCategoryID: categoryPgUUID,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to check seat availability")
		return
	}
	if available > 0 {
		RespondError(w, http.StatusConflict, "CATEGORY_AVAILABLE", "seats are still available in this category; book directly from the seat map")
		return
	}

	entry, err := h.queries.JoinWaitlist(r.Context(), generated.JoinWaitlistParams{
		EventID:        eventPgUUID,
		SeatCategoryID: categoryPgUUID,
		CustomerID:     utils.UUIDToPgtype(customerID),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			RespondError(w, http.StatusConflict, "ALREADY_ON_WAITLIST", "you are already on the waitlist for this category")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to join waitlist")
		return
	}

	position, err := h.queries.GetWaitlistQueuePosition(r.Context(), generated.GetWaitlistQueuePositionParams{
		EventID:        eventPgUUID,
		SeatCategoryID: categoryPgUUID,
		CreatedAt:      entry.CreatedAt,
	})
	if err != nil {
		position = 0
	}

	RespondSuccess(w, http.StatusCreated, WaitlistEntryResponse{
		ID:             utils.PgtypeToUUID(entry.ID).String(),
		EventID:        utils.PgtypeToUUID(entry.EventID).String(),
		SeatCategoryID: utils.PgtypeToUUID(entry.SeatCategoryID).String(),
		Status:         string(entry.Status),
		QueuePosition:  int(position),
		CreatedAt:      utils.PgtypeToTime(entry.CreatedAt),
	}, "Joined waitlist successfully")
}

// ListMyWaitlist returns the authenticated customer's active waitlist entries with queue positions.
func (h *WaitlistHandler) ListMyWaitlist(w http.ResponseWriter, r *http.Request) {
	customerID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	rows, err := h.queries.GetCustomerWaitlists(r.Context(), utils.UUIDToPgtype(customerID))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve waitlists")
		return
	}

	res := make([]WaitlistEntryResponse, 0, len(rows))
	for _, row := range rows {
		entry := WaitlistEntryResponse{
			ID:             utils.PgtypeToUUID(row.ID).String(),
			EventID:        utils.PgtypeToUUID(row.EventID).String(),
			SeatCategoryID: utils.PgtypeToUUID(row.SeatCategoryID).String(),
			Status:         string(row.Status),
			QueuePosition:  int(row.QueuePosition),
			CreatedAt:      utils.PgtypeToTime(row.CreatedAt),
		}
		if row.EventTitle != "" {
			title := row.EventTitle
			entry.EventTitle = &title
		}
		if row.CategoryName != "" {
			name := row.CategoryName
			entry.CategoryName = &name
		}
		if row.CategoryColor != "" {
			color := row.CategoryColor
			entry.CategoryColor = &color
		}
		if row.OfferToken.Valid {
			tok := utils.PgtypeToUUID(row.OfferToken).String()
			entry.OfferToken = &tok
		}
		if row.OfferExpiresAt.Valid {
			t := utils.PgtypeToTime(row.OfferExpiresAt)
			entry.OfferExpiresAt = &t
		}
		if row.EventStartTime.Valid {
			t := utils.PgtypeToTime(row.EventStartTime)
			entry.EventStartTime = &t
		}
		res = append(res, entry)
	}

	RespondSuccess(w, http.StatusOK, res, "Waitlists retrieved")
}

// GetOfferDetails returns metadata and seat details for a waitlist offer token.
func (h *WaitlistHandler) GetOfferDetails(w http.ResponseWriter, r *http.Request) {
	offerToken, err := uuid.Parse(chi.URLParam(r, "token"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_TOKEN", "invalid offer token format")
		return
	}

	offer, err := h.queries.GetWaitlistOfferByToken(r.Context(), utils.UUIDToPgtype(offerToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "OFFER_NOT_FOUND", "offer not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve offer")
		return
	}

	var startTime *time.Time
	if offer.EventStartTime.Valid {
		t := utils.PgtypeToTime(offer.EventStartTime)
		startTime = &t
	}
	var endTime *time.Time
	if offer.EventEndTime.Valid {
		t := utils.PgtypeToTime(offer.EventEndTime)
		endTime = &t
	}

	RespondSuccess(w, http.StatusOK, WaitlistOfferDetailResponse{
		ID:              utils.PgtypeToUUID(offer.ID).String(),
		WaitlistEntryID: utils.PgtypeToUUID(offer.WaitlistEntryID).String(),
		EventID:         utils.PgtypeToUUID(offer.EventID).String(),
		EventTitle:      offer.EventTitle,
		EventStartTime:  startTime,
		EventEndTime:    endTime,
		SeatID:          utils.PgtypeToUUID(offer.SeatID).String(),
		RowLabel:        offer.RowLabel,
		SeatNumber:      offer.SeatNumber,
		SeatCategoryID:  utils.PgtypeToUUID(offer.SeatCategoryID).String(),
		CategoryName:    offer.CategoryName,
		Price:           utils.PgtypeNumericToFloat64(offer.Price),
		Currency:        offer.Currency,
		OfferToken:      utils.PgtypeToUUID(offer.OfferToken).String(),
		OfferedAt:       utils.PgtypeToTime(offer.OfferedAt),
		ExpiresAt:       utils.PgtypeToTime(offer.ExpiresAt),
		Status:          string(offer.Status),
	}, "Offer details retrieved")
}

// AcceptOffer converts a time-limited waitlist offer into a confirmed booking
// with tickets and QR codes. The offer must still be PENDING and unexpired.
func (h *WaitlistHandler) AcceptOffer(w http.ResponseWriter, r *http.Request) {
	offerToken, err := uuid.Parse(chi.URLParam(r, "token"))
	if err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_TOKEN", "invalid offer token format")
		return
	}

	customerID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	offer, err := h.queries.GetWaitlistOfferByToken(r.Context(), utils.UUIDToPgtype(offerToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "OFFER_NOT_FOUND", "waitlist offer not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve waitlist offer")
		return
	}

	// Ownership check: only the offered customer may accept the offer.
	if utils.PgtypeToUUID(offer.CustomerID) != customerID {
		RespondError(w, http.StatusForbidden, "FORBIDDEN", "this offer belongs to another account")
		return
	}
	if offer.Status != generated.OfferStatusPENDING {
		RespondError(w, http.StatusConflict, "OFFER_NOT_PENDING", "this offer has already been accepted, expired or revoked")
		return
	}
	if !offer.ExpiresAt.Valid || !time.Now().Before(offer.ExpiresAt.Time) {
		RespondError(w, http.StatusGone, "OFFER_EXPIRED", "this offer has expired; the seat has been passed to the next customer in line")
		return
	}

	// Fetch the OFFERED seat reservation created for this offer.
	// AssignSeat sets its hold_token equal to the offer token.
	heldSeats, err := h.queries.GetActiveHoldOrOfferByToken(r.Context(), generated.GetActiveHoldOrOfferByTokenParams{
		EventID:   offer.EventID,
		HoldToken: offer.OfferToken,
	})
	if err != nil || len(heldSeats) == 0 {
		RespondError(w, http.StatusConflict, "OFFER_UNAVAILABLE", "the offered seat is no longer reserved for you")
		return
	}
	for _, hs := range heldSeats {
		if hs.UserID.Valid && utils.PgtypeToUUID(hs.UserID) != customerID {
			RespondError(w, http.StatusForbidden, "FORBIDDEN", "this offer belongs to another account")
			return
		}
	}

	totalAmount := 0.0
	currency := "USD"
	for _, hs := range heldSeats {
		totalAmount += utils.PgtypeNumericToFloat64(hs.Price)
		if hs.Currency != "" {
			currency = hs.Currency
		}
	}

	var booking generated.Booking
	var tickets []TicketDetailResponse
	if h.pool != nil {
		tx, txErr := h.pool.Begin(r.Context())
		if txErr != nil {
			RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to start transaction")
			return
		}
		defer tx.Rollback(r.Context())

		booking, tickets, err = confirmOfferTx(r.Context(), generated.New(tx), offer, heldSeats, customerID, totalAmount, currency)
		if err == nil {
			err = tx.Commit(r.Context())
		}
	} else {
		// Non-pool execution for mock tests.
		booking, tickets, err = confirmOfferTx(r.Context(), h.queries, offer, heldSeats, customerID, totalAmount, currency)
	}
	if err != nil {
		respondWaitlistConfirmError(w, err)
		return
	}

	h.sendOfferConfirmationEmail(booking.BookingReference, offer, tickets, totalAmount, currency)

	RespondSuccess(w, http.StatusCreated, BookingResponse{
		ID:               utils.PgtypeToUUID(booking.ID).String(),
		BookingReference: booking.BookingReference,
		CustomerID:       customerID.String(),
		EventID:          utils.PgtypeToUUID(offer.EventID).String(),
		TotalAmount:      totalAmount,
		Currency:         currency,
		Status:           string(booking.Status),
		TicketCount:      len(tickets),
		Tickets:          tickets,
		CreatedAt:        utils.PgtypeToTime(booking.CreatedAt),
	}, "Waitlist offer accepted and booking confirmed")
}

// ============================================================================
// OFFER CONFIRMATION HELPERS
// ============================================================================

type confirmErrKind int

const (
	errKindDB confirmErrKind = iota
	errKindExpired
)

type confirmedError struct {
	kind confirmErrKind
	err  error
}

func (e *confirmedError) Error() string { return e.err.Error() }
func (e *confirmedError) Unwrap() error { return e.err }

func wrapConfirmErr(kind confirmErrKind, err error) *confirmedError {
	return &confirmedError{kind: kind, err: err}
}

func respondWaitlistConfirmError(w http.ResponseWriter, err error) {
	var ce *confirmedError
	if errors.As(err, &ce) && ce.kind == errKindExpired {
		RespondError(w, http.StatusConflict, "OFFER_EXPIRED", "the offer expired while confirming; the seat was released")
		return
	}
	log.Printf("⚠️  Waitlist offer confirmation failed: %v", err)
	RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to confirm booking from waitlist offer")
}

// confirmOfferTx performs the transactional part of offer acceptance:
// booking creation + tickets + reservation confirmation + waitlist entry acceptance.
func confirmOfferTx(
	ctx context.Context,
	q generated.Querier,
	offer generated.GetWaitlistOfferByTokenRow,
	heldSeats []generated.GetActiveHoldOrOfferByTokenRow,
	customerID uuid.UUID,
	totalAmount float64,
	currency string,
) (generated.Booking, []TicketDetailResponse, error) {
	eventPgUUID := offer.EventID
	holdTokenPgUUID := offer.OfferToken

	booking, err := q.CreateBooking(ctx, generated.CreateBookingParams{
		BookingReference: generateBookingReference(),
		CustomerID:       utils.UUIDToPgtype(customerID),
		EventID:          eventPgUUID,
		TotalAmount:      utils.Float64ToPgtypeNumeric(totalAmount),
		Currency:         currency,
	})
	if err != nil {
		return generated.Booking{}, nil, wrapConfirmErr(errKindDB, err)
	}

	tickets := make([]TicketDetailResponse, 0, len(heldSeats))
	for _, hs := range heldSeats {
		ticketUUID := uuid.New()
		seatIDStr := utils.PgtypeToUUID(hs.SeatID).String()
		qrPayload := qrcode.GenerateTicketPayload(booking.BookingReference,
			seatIDStr, ticketUUID.String())
		tRecord, err := q.CreateTicket(ctx, generated.CreateTicketParams{
			BookingID:     booking.ID,
			EventID:       eventPgUUID,
			SeatID:        hs.SeatID,
			UnitPrice:     hs.Price,
			QrCodePayload: qrPayload,
		})
		if err != nil {
			return generated.Booking{}, nil, wrapConfirmErr(errKindDB, err)
		}

		qrDataURL, _ := qrcode.GenerateDataURL(qrPayload, 200)
		tickets = append(tickets, TicketDetailResponse{
			ID:             utils.PgtypeToUUID(tRecord.ID).String(),
			BookingID:      utils.PgtypeToUUID(booking.ID).String(),
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

	confirmedRows, err := q.ConfirmReservationToBooked(ctx, generated.ConfirmReservationToBookedParams{
		EventID:   eventPgUUID,
		HoldToken: holdTokenPgUUID,
		BookingID: booking.ID,
	})
	if err != nil {
		return generated.Booking{}, nil, wrapConfirmErr(errKindDB, err)
	}
	if confirmedRows != int64(len(heldSeats)) {
		// The offer expired concurrently and its reservation was released.
		return generated.Booking{}, nil, wrapConfirmErr(errKindExpired, errors.New("reservation no longer active"))
	}

	if _, err := q.UpdateWaitlistEntryStatus(ctx, generated.UpdateWaitlistEntryStatusParams{
		ID:     offer.WaitlistEntryID,
		Status: generated.WaitlistStatusACCEPTED,
	}); err != nil {
		return generated.Booking{}, nil, wrapConfirmErr(errKindDB, err)
	}

	if _, err := q.AcceptWaitlistOffer(ctx, offer.ID); err != nil {
		return generated.Booking{}, nil, wrapConfirmErr(errKindDB, err)
	}
	return booking, tickets, nil
}

// sendOfferConfirmationEmail best-effort sends the ticket confirmation email for an accepted offer.
func (h *WaitlistHandler) sendOfferConfirmationEmail(bookingRef string, offer generated.GetWaitlistOfferByTokenRow, tickets []TicketDetailResponse, totalAmount float64, currency string) {
	if h.mailer == nil || len(tickets) == 0 {
		return
	}
	seats := make([]email.SeatInfo, 0, len(tickets))
	var firstQR []byte
	qrName := "ticket.png"
	for i, t := range tickets {
		seats = append(seats, email.SeatInfo{
			RowLabel:     t.RowLabel,
			SeatNumber:   t.SeatNumber,
			CategoryName: t.CategoryName,
			UnitPrice:    t.UnitPrice,
		})
		if png, err := qrcode.GeneratePNG(t.QRCodePayload, 300); err == nil && i == 0 {
			firstQR = png
			qrName = fmt.Sprintf("ticket-%s%s.png", t.RowLabel, t.SeatNumber)
		}
	}

	var start *time.Time
	if offer.EventStartTime.Valid {
		t := utils.PgtypeToTime(offer.EventStartTime)
		start = &t
	}

	go func() {
		info := email.TicketConfirmation{
			CustomerName:     offer.CustomerName,
			EventTitle:       offer.EventTitle,
			EventStartTime:   start,
			VenueName:        "",
			VenueCity:        "",
			BookingRef:       bookingRef,
			Seats:            seats,
			TotalAmount:      totalAmount,
			Currency:         currency,
			QRPNG:            firstQR,
			QRAttachmentName: qrName,
		}
		if err := h.mailer.SendTicketConfirmation(offer.CustomerEmail, info); err != nil {
			log.Printf("⚠️  Failed to send waitlist-offer confirmation email to %s: %v", offer.CustomerEmail, err)
		}
	}()
}

// ============================================================================
// WAITLIST AUTO-ASSIGNMENT ENGINE
// ============================================================================

// WaitlistAssigner implements automatic FIFO seat assignment on cancellation
// and time-limited offer expiry. It is shared by the cancellation handler and
// the background scheduler.
type WaitlistAssigner struct {
	queries generated.Querier
	pool    *pgxpool.Pool
	cfg     *config.Config
	mailer  *email.Service
}

func NewWaitlistAssigner(queries generated.Querier, pool *pgxpool.Pool, cfg *config.Config, mailer *email.Service) *WaitlistAssigner {
	return &WaitlistAssigner{queries: queries, pool: pool, cfg: cfg, mailer: mailer}
}

// AssignSeat offers a free seat to the next customer on the category waitlist.
// Within one transaction it atomically:
//  1. re-verifies the seat is actually still available,
//  2. picks the oldest WAITING candidate (SELECT ... FOR UPDATE SKIP LOCKED),
//  3. marks the entry OFFERED,
//  4. creates an OFFERED seat reservation (blocking other customers via the
//     partial unique index) whose hold_token equals the offer token,
//  5. creates the PENDING waitlist offer with cfg.WaitlistOfferTTL expiry.
//
// After commit it sends the notification email with the time-limited link.
func (a *WaitlistAssigner) AssignSeat(ctx context.Context, eventID, seatID uuid.UUID) bool {
	eventPgUUID := utils.UUIDToPgtype(eventID)
	seatPgUUID := utils.UUIDToPgtype(seatID)

	exec := func(q generated.Querier) (generated.GetNextWaitlistCandidateRow, pgtype.UUID, pgtype.Timestamptz, error) {
		avail, err := q.CheckSeatAvailable(ctx, generated.CheckSeatAvailableParams{
			ID:   eventPgUUID,
			ID_2: seatPgUUID,
		})
		if err != nil {
			return generated.GetNextWaitlistCandidateRow{}, pgtype.UUID{}, pgtype.Timestamptz{}, err
		}

		candidate, err := q.GetNextWaitlistCandidate(ctx, generated.GetNextWaitlistCandidateParams{
			EventID:        eventPgUUID,
			SeatCategoryID: avail.SeatCategoryID,
		})
		if err != nil {
			return generated.GetNextWaitlistCandidateRow{}, pgtype.UUID{}, pgtype.Timestamptz{}, err
		}

		offerToken := uuid.New()
		expiresAt := time.Now().Add(a.cfg.WaitlistOfferTTL)

		if _, err := q.UpdateWaitlistEntryStatus(ctx, generated.UpdateWaitlistEntryStatusParams{
			ID:     candidate.ID,
			Status: generated.WaitlistStatusOFFERED,
		}); err != nil {
			return generated.GetNextWaitlistCandidateRow{}, pgtype.UUID{}, pgtype.Timestamptz{}, err
		}

		if _, err := q.CreateSeatReservation(ctx, generated.CreateSeatReservationParams{
			EventID:   eventPgUUID,
			SeatID:    seatPgUUID,
			UserID:    candidate.CustomerID,
			Status:    generated.ReservationStatusOFFERED,
			HoldToken: utils.UUIDToPgtype(offerToken),
			ExpiresAt: utils.TimeToPgtype(expiresAt),
		}); err != nil {
			return generated.GetNextWaitlistCandidateRow{}, pgtype.UUID{}, pgtype.Timestamptz{}, err
		}

		if _, err := q.CreateWaitlistOffer(ctx, generated.CreateWaitlistOfferParams{
			WaitlistEntryID: candidate.ID,
			EventID:         eventPgUUID,
			SeatID:          seatPgUUID,
			OfferToken:      utils.UUIDToPgtype(offerToken),
			ExpiresAt:       utils.TimeToPgtype(expiresAt),
		}); err != nil {
			return generated.GetNextWaitlistCandidateRow{}, pgtype.UUID{}, pgtype.Timestamptz{}, err
		}
		return candidate, utils.UUIDToPgtype(offerToken), utils.TimeToPgtype(expiresAt), nil
	}

	var candidate generated.GetNextWaitlistCandidateRow
	var offerToken pgtype.UUID
	var expiresAt pgtype.Timestamptz

	if a.pool != nil {
		tx, err := a.pool.Begin(ctx)
		if err != nil {
			log.Printf("⚠️  Waitlist assigner: failed to start transaction: %v", err)
			return false
		}
		defer tx.Rollback(ctx)

		candidate, offerToken, expiresAt, err = exec(generated.New(tx))
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				log.Printf("⚠️  Waitlist assigner: %v", err)
			}
			return false
		}
		if err := tx.Commit(ctx); err != nil {
			log.Printf("⚠️  Waitlist assigner: commit failed: %v", err)
			return false
		}
	} else {
		var err error
		candidate, offerToken, expiresAt, err = exec(a.queries)
		if err != nil || !offerToken.Valid {
			return false
		}
	}

	a.notifyOffer(ctx, eventID, candidate, offerToken, expiresAt)
	return true
}

// ExpireStaleOffers processes every PENDING offer past its deadline:
// the offer is marked EXPIRED, its waitlist entry EXPIRED, and the OFFERED
// seat reservation is released so the seat can be offered to the next in line.
func (a *WaitlistAssigner) ExpireStaleOffers(ctx context.Context) int {
	expiredOffers, err := a.queries.GetPendingExpiredOffers(ctx)
	if err != nil {
		log.Printf("⚠️  Offer expiry worker: failed to list expired offers: %v", err)
		return 0
	}

	processed := 0
	for _, offer := range expiredOffers {
		eventID := utils.PgtypeToUUID(offer.EventID)
		seatID := utils.PgtypeToUUID(offer.SeatID)

		release := func(q generated.Querier) error {
			if _, err := q.ExpireWaitlistOffer(ctx, offer.ID); err != nil {
				return err
			}
			if _, err := q.UpdateWaitlistEntryStatus(ctx, generated.UpdateWaitlistEntryStatusParams{
				ID:     offer.WaitlistEntryID,
				Status: generated.WaitlistStatusEXPIRED,
			}); err != nil {
				return err
			}
			_, err := q.RevokeOfferedSeatReservation(ctx, generated.RevokeOfferedSeatReservationParams{
				EventID: offer.EventID,
				SeatID:  offer.SeatID,
			})
			return err
		}

		if a.pool != nil {
			tx, txErr := a.pool.Begin(ctx)
			if txErr != nil {
				log.Printf("⚠️  Offer expiry worker: failed to start transaction: %v", txErr)
				continue
			}
			if err := release(generated.New(tx)); err == nil {
				err = tx.Commit(ctx)
			} else {
				tx.Rollback(ctx)
			}
			if err != nil {
				log.Printf("⚠️  Offer expiry worker: failed to expire offer %s: %v", utils.PgtypeToUUID(offer.ID), err)
				continue
			}
		} else if err := release(a.queries); err != nil {
			continue
		}

		processed++
		// Pass the freed seat to the next customer in line.
		a.AssignSeat(ctx, eventID, seatID)
	}
	return processed
}

// notifyOffer best-effort emails the offered customer their time-limited link.
func (a *WaitlistAssigner) notifyOffer(ctx context.Context, eventID uuid.UUID, candidate generated.GetNextWaitlistCandidateRow, offerToken pgtype.UUID, expiresAt pgtype.Timestamptz) {
	if a.mailer == nil || !offerToken.Valid || !expiresAt.Valid {
		return
	}

	detail, err := a.queries.GetWaitlistOfferByToken(ctx, offerToken)
	if err != nil {
		log.Printf("⚠️  Waitlist assigner: could not load offer details for email: %v", err)
		return
	}

	var start *time.Time
	if detail.EventStartTime.Valid {
		t := utils.PgtypeToTime(detail.EventStartTime)
		start = &t
	}

	info := email.WaitlistOfferNotification{
		CustomerName:   candidate.CustomerName,
		EventTitle:     detail.EventTitle,
		EventStartTime: start,
		CategoryName:   detail.CategoryName,
		RowLabel:       detail.RowLabel,
		SeatNumber:     detail.SeatNumber,
		Price:          utils.PgtypeNumericToFloat64(detail.Price),
		Currency:       detail.Currency,
		OfferURL: fmt.Sprintf("%s/waitlist/offer?token=%s&event=%s",
			strings.TrimRight(a.cfg.FrontendBaseURL, "/"), offerToken.String(), eventID.String()),
		ExpiresAt: utils.PgtypeToTime(expiresAt),
	}

	go func() {
		if err := a.mailer.SendWaitlistOffer(candidate.CustomerEmail, info); err != nil {
			log.Printf("⚠️  Failed to send waitlist offer email to %s: %v", candidate.CustomerEmail, err)
		}
	}()
}
