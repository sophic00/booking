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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/utils"
)

type mockVenueQuerier struct {
	generated.Querier
	listSeatCategoriesFunc  func(ctx context.Context) ([]generated.SeatCategory, error)
	createSeatCategoryFunc  func(ctx context.Context, arg generated.CreateSeatCategoryParams) (generated.SeatCategory, error)
	listVenuesFunc          func(ctx context.Context) ([]generated.Venue, error)
	getVenueByIDFunc        func(ctx context.Context, id pgtype.UUID) (generated.Venue, error)
	createVenueFunc         func(ctx context.Context, arg generated.CreateVenueParams) (generated.Venue, error)
	updateVenueFunc         func(ctx context.Context, arg generated.UpdateVenueParams) (generated.Venue, error)
	deleteVenueFunc         func(ctx context.Context, id pgtype.UUID) error
	getSeatsByVenueIDFunc   func(ctx context.Context, venueID pgtype.UUID) ([]generated.GetSeatsByVenueIDRow, error)
	createSeatFunc          func(ctx context.Context, arg generated.CreateSeatParams) (generated.Seat, error)
	deleteSeatsByVenueIDFunc func(ctx context.Context, venueID pgtype.UUID) error
}

func (m *mockVenueQuerier) ListSeatCategories(ctx context.Context) ([]generated.SeatCategory, error) {
	if m.listSeatCategoriesFunc != nil {
		return m.listSeatCategoriesFunc(ctx)
	}
	return nil, nil
}

func (m *mockVenueQuerier) CreateSeatCategory(ctx context.Context, arg generated.CreateSeatCategoryParams) (generated.SeatCategory, error) {
	if m.createSeatCategoryFunc != nil {
		return m.createSeatCategoryFunc(ctx, arg)
	}
	return generated.SeatCategory{}, nil
}

func (m *mockVenueQuerier) ListVenues(ctx context.Context) ([]generated.Venue, error) {
	if m.listVenuesFunc != nil {
		return m.listVenuesFunc(ctx)
	}
	return nil, nil
}

func (m *mockVenueQuerier) GetVenueByID(ctx context.Context, id pgtype.UUID) (generated.Venue, error) {
	if m.getVenueByIDFunc != nil {
		return m.getVenueByIDFunc(ctx, id)
	}
	return generated.Venue{}, pgx.ErrNoRows
}

func (m *mockVenueQuerier) CreateVenue(ctx context.Context, arg generated.CreateVenueParams) (generated.Venue, error) {
	if m.createVenueFunc != nil {
		return m.createVenueFunc(ctx, arg)
	}
	return generated.Venue{}, nil
}

func (m *mockVenueQuerier) UpdateVenue(ctx context.Context, arg generated.UpdateVenueParams) (generated.Venue, error) {
	if m.updateVenueFunc != nil {
		return m.updateVenueFunc(ctx, arg)
	}
	return generated.Venue{}, nil
}

func (m *mockVenueQuerier) DeleteVenue(ctx context.Context, id pgtype.UUID) error {
	if m.deleteVenueFunc != nil {
		return m.deleteVenueFunc(ctx, id)
	}
	return nil
}

func (m *mockVenueQuerier) GetSeatsByVenueID(ctx context.Context, venueID pgtype.UUID) ([]generated.GetSeatsByVenueIDRow, error) {
	if m.getSeatsByVenueIDFunc != nil {
		return m.getSeatsByVenueIDFunc(ctx, venueID)
	}
	return nil, nil
}

func (m *mockVenueQuerier) CreateSeat(ctx context.Context, arg generated.CreateSeatParams) (generated.Seat, error) {
	if m.createSeatFunc != nil {
		return m.createSeatFunc(ctx, arg)
	}
	return generated.Seat{}, nil
}

func (m *mockVenueQuerier) DeleteSeatsByVenueID(ctx context.Context, venueID pgtype.UUID) error {
	if m.deleteSeatsByVenueIDFunc != nil {
		return m.deleteSeatsByVenueIDFunc(ctx, venueID)
	}
	return nil
}

func TestListCategories(t *testing.T) {
	mock := &mockVenueQuerier{
		listSeatCategoriesFunc: func(ctx context.Context) ([]generated.SeatCategory, error) {
			return []generated.SeatCategory{
				{
					ID:        utils.UUIDToPgtype(uuid.New()),
					Name:      "Premium",
					ColorCode: "#F59E0B",
					CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				},
				{
					ID:        utils.UUIDToPgtype(uuid.New()),
					Name:      "Standard",
					ColorCode: "#3B82F6",
					CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				},
			}, nil
		},
	}

	handler := NewVenueHandler(mock, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	rr := httptest.NewRecorder()

	handler.ListCategories(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool               `json:"success"`
		Data    []CategoryResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, "Premium", resp.Data[0].Name)
}

func TestCreateCategory_Success(t *testing.T) {
	catID := uuid.New()
	desc := "Front row seats"
	color := "#10B981"

	mock := &mockVenueQuerier{
		createSeatCategoryFunc: func(ctx context.Context, arg generated.CreateSeatCategoryParams) (generated.SeatCategory, error) {
			assert.Equal(t, "VIP", arg.Name)
			assert.Equal(t, color, arg.ColorCode)
			return generated.SeatCategory{
				ID:          utils.UUIDToPgtype(catID),
				Name:        arg.Name,
				Description: arg.Description,
				ColorCode:   arg.ColorCode,
				CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	handler := NewVenueHandler(mock, nil)

	body := CreateCategoryRequest{
		Name:        "VIP",
		Description: &desc,
		ColorCode:   &color,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/categories", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	handler.CreateCategory(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp struct {
		Success bool             `json:"success"`
		Data    CategoryResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "VIP", resp.Data.Name)
	assert.Equal(t, "#10B981", resp.Data.ColorCode)
}

func TestCreateCategory_InvalidColor(t *testing.T) {
	handler := NewVenueHandler(&mockVenueQuerier{}, nil)

	badColor := "not-a-color"
	body := CreateCategoryRequest{
		Name:      "VIP",
		ColorCode: &badColor,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/categories", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	handler.CreateCategory(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "INVALID_COLOR")
}

func TestCreateCategory_Duplicate(t *testing.T) {
	mock := &mockVenueQuerier{
		createSeatCategoryFunc: func(ctx context.Context, arg generated.CreateSeatCategoryParams) (generated.SeatCategory, error) {
			return generated.SeatCategory{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation}
		},
	}

	handler := NewVenueHandler(mock, nil)

	body := CreateCategoryRequest{
		Name: "Standard",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/categories", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	handler.CreateCategory(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "CATEGORY_EXISTS")
}

func TestCreateVenue_Success(t *testing.T) {
	venueID := uuid.New()
	mock := &mockVenueQuerier{
		createVenueFunc: func(ctx context.Context, arg generated.CreateVenueParams) (generated.Venue, error) {
			assert.Equal(t, "Grand Arena", arg.Name)
			assert.Equal(t, "123 Main St", arg.Address)
			assert.Equal(t, "Mumbai", arg.City)
			return generated.Venue{
				ID:            utils.UUIDToPgtype(venueID),
				Name:          arg.Name,
				Address:       arg.Address,
				City:          arg.City,
				Country:       arg.Country,
				TotalCapacity: arg.TotalCapacity,
				CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	handler := NewVenueHandler(mock, nil)

	body := CreateVenueRequest{
		Name:    "Grand Arena",
		Address: "123 Main St",
		City:    "Mumbai",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/venues", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	handler.CreateVenue(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp struct {
		Success bool          `json:"success"`
		Data    VenueResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Grand Arena", resp.Data.Name)
	assert.Equal(t, venueID.String(), resp.Data.ID)
}

func TestGetVenue_Success(t *testing.T) {
	venueID := uuid.New()
	mock := &mockVenueQuerier{
		getVenueByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.Venue, error) {
			return generated.Venue{
				ID:            id,
				Name:          "Royal Theatre",
				Address:       "456 Park Ave",
				City:          "London",
				Country:       "UK",
				TotalCapacity: 500,
				CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	handler := NewVenueHandler(mock, nil)

	r := chi.NewRouter()
	r.Get("/api/v1/venues/{id}", handler.GetVenue)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/venues/"+venueID.String(), nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool          `json:"success"`
		Data    VenueResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Royal Theatre", resp.Data.Name)
	assert.Equal(t, 500, resp.Data.TotalCapacity)
}

func TestBatchCreateSeats_Success(t *testing.T) {
	venueID := uuid.New()
	catID := uuid.New()

	deletedCalled := false
	createdSeatsCount := 0

	mock := &mockVenueQuerier{
		deleteSeatsByVenueIDFunc: func(ctx context.Context, vID pgtype.UUID) error {
			deletedCalled = true
			return nil
		},
		createSeatFunc: func(ctx context.Context, arg generated.CreateSeatParams) (generated.Seat, error) {
			createdSeatsCount++
			return generated.Seat{
				ID:             utils.UUIDToPgtype(uuid.New()),
				VenueID:        arg.VenueID,
				SeatCategoryID: arg.SeatCategoryID,
				RowLabel:       arg.RowLabel,
				SeatNumber:     arg.SeatNumber,
				GridRow:        arg.GridRow,
				GridCol:        arg.GridCol,
				IsActive:       arg.IsActive,
				CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	handler := NewVenueHandler(mock, nil)

	r := chi.NewRouter()
	r.Post("/api/v1/admin/venues/{id}/seats/batch", handler.BatchCreateSeats)

	body := BatchCreateSeatsRequest{
		Replace: true,
		Seats: []CreateSeatRequest{
			{
				SeatCategoryID: catID.String(),
				RowLabel:       "A",
				SeatNumber:     "1",
				GridRow:        1,
				GridCol:        1,
			},
			{
				SeatCategoryID: catID.String(),
				RowLabel:       "A",
				SeatNumber:     "2",
				GridRow:        1,
				GridCol:        2,
			},
			{
				SeatCategoryID: catID.String(),
				RowLabel:       "B",
				SeatNumber:     "1",
				GridRow:        2,
				GridCol:        1,
			},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/venues/"+venueID.String()+"/seats/batch", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.True(t, deletedCalled)
	assert.Equal(t, 3, createdSeatsCount)
}

func TestBatchCreateSeats_DuplicateCoordinates(t *testing.T) {
	venueID := uuid.New()
	catID := uuid.New()

	handler := NewVenueHandler(&mockVenueQuerier{}, nil)

	r := chi.NewRouter()
	r.Post("/api/v1/admin/venues/{id}/seats/batch", handler.BatchCreateSeats)

	body := BatchCreateSeatsRequest{
		Replace: true,
		Seats: []CreateSeatRequest{
			{
				SeatCategoryID: catID.String(),
				RowLabel:       "A",
				SeatNumber:     "1",
				GridRow:        1,
				GridCol:        1,
			},
			{
				SeatCategoryID: catID.String(),
				RowLabel:       "A",
				SeatNumber:     "2",
				GridRow:        1,
				GridCol:        1, // Duplicate grid coordinate!
			},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/venues/"+venueID.String()+"/seats/batch", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "DUPLICATE_COORDINATES")
}

func TestBatchCreateSeats_DuplicateLabels(t *testing.T) {
	venueID := uuid.New()
	catID := uuid.New()

	handler := NewVenueHandler(&mockVenueQuerier{}, nil)

	r := chi.NewRouter()
	r.Post("/api/v1/admin/venues/{id}/seats/batch", handler.BatchCreateSeats)

	body := BatchCreateSeatsRequest{
		Replace: true,
		Seats: []CreateSeatRequest{
			{
				SeatCategoryID: catID.String(),
				RowLabel:       "A",
				SeatNumber:     "1",
				GridRow:        1,
				GridCol:        1,
			},
			{
				SeatCategoryID: catID.String(),
				RowLabel:       "A",
				SeatNumber:     "1", // Duplicate seat label!
				GridRow:        1,
				GridCol:        2,
			},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/venues/"+venueID.String()+"/seats/batch", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "DUPLICATE_SEAT_LABEL")
}
