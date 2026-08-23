# VelvetSeats REST API Reference

This document outlines the API specifications for the VelvetSeats platform backend service.

- **Base URL**: `http://localhost:8080/api/v1`
- **Protocol**: HTTP/1.1 or HTTP/2
- **Data Format**: JSON (`Content-Type: application/json`)
- **Authentication**: Bearer Token in `Authorization: Bearer <JWT>` header

---

## Standard Error Response Format

All failure responses conform to the following schema:

```json
{
  "error": "Error description message",
  "code": "ERROR_CODE_STRING",
  "status": 400
}
```

### Common HTTP Status Codes
- `200 OK`: Request succeeded with payload.
- `201 Created`: Resource successfully created.
- `204 No Content`: Action succeeded with no response body.
- `400 Bad Request`: Invalid payload, malformed parameters, or validation failure.
- `401 Unauthorized`: Missing or invalid JWT credentials.
- `403 Forbidden`: Insufficient role permissions for the endpoint.
- `404 Not Found`: Target entity does not exist.
- `409 Conflict`: Concurrency conflict (e.g. seat already held/booked, duplicate entry).
- `500 Internal Server Error`: Unhandled server exception.

---

## 1. Authentication & Identity

### 1.1 Register Account
Creates a new user account.

- **Endpoint**: `POST /api/v1/auth/register`
- **Auth Required**: No
- **Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123",
  "full_name": "Jane Doe",
  "phone": "+14155552671",
  "role": "CUSTOMER"
}
```
*Note: `role` must be one of `CUSTOMER`, `ORGANISER`, or `ADMIN`.*

- **Response `201 Created`**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "email": "user@example.com",
    "full_name": "Jane Doe",
    "phone": "+14155552671",
    "role": "CUSTOMER",
    "created_at": "2026-08-23T10:00:00Z"
  }
}
```

---

### 1.2 User Login
Authenticates an existing user and returns a signed JWT.

- **Endpoint**: `POST /api/v1/auth/login`
- **Auth Required**: No
- **Request Body**:
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123"
}
```
- **Response `200 OK`**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "email": "user@example.com",
    "full_name": "Jane Doe",
    "role": "CUSTOMER"
  }
}
```

---

### 1.3 Get Current Profile
Fetches identity metadata for the authenticated caller.

- **Endpoint**: `GET /api/v1/auth/me`
- **Auth Required**: Yes (`CUSTOMER`, `ORGANISER`, `ADMIN`)
- **Response `200 OK`**: User profile object.

---

## 2. Venues, Auditoriums & Seat Layouts

### 2.1 List Venues
Retrieves all registered venues.

- **Endpoint**: `GET /api/v1/venues`
- **Auth Required**: No
- **Response `200 OK`**:
```json
[
  {
    "id": "00000000-0000-0000-0000-000000000001",
    "name": "Metropolitan Grand Arena",
    "address": "100 Stadium Way",
    "city": "San Francisco",
    "state": "CA",
    "country": "US",
    "total_capacity": 60,
    "created_at": "2026-08-23T00:00:00Z"
  }
]
```

---

### 2.2 Create Venue
Registers a new venue or auditorium.

- **Endpoint**: `POST /api/v1/venues`
- **Auth Required**: Yes (`ADMIN`)
- **Request Body**:
```json
{
  "name": "Dolby Atmos Cinema Hall 1",
  "address": "450 Market St",
  "city": "San Francisco",
  "state": "CA",
  "country": "US"
}
```
- **Response `201 Created`**: Created venue entity.

---

### 2.3 Batch Configure Venue Seats
Configures the 2D layout and seat categories for a venue.

- **Endpoint**: `POST /api/v1/venues/{id}/seats/batch`
- **Auth Required**: Yes (`ADMIN`)
- **Request Body**:
```json
{
  "replace": true,
  "seats": [
    {
      "seat_category_id": "00000000-0000-0000-0000-000000000001",
      "row_label": "A",
      "seat_number": "1",
      "grid_row": 1,
      "grid_col": 1,
      "is_active": true
    },
    {
      "seat_category_id": "00000000-0000-0000-0000-000000000001",
      "row_label": "A",
      "seat_number": "2",
      "grid_row": 1,
      "grid_col": 2,
      "is_active": true
    }
  ]
}
```
- **Response `200 OK`**:
```json
{
  "venue_id": "00000000-0000-0000-0000-000000000001",
  "total_configured": 60,
  "active_capacity": 56
}
```

---

## 3. Seat Categories

### 3.1 List Categories
- **Endpoint**: `GET /api/v1/categories`
- **Auth Required**: No
- **Response `200 OK`**: Array of category objects.

### 3.2 Create Category
- **Endpoint**: `POST /api/v1/categories`
- **Auth Required**: Yes (`ADMIN`)
- **Request Body**:
```json
{
  "name": "VIP Lounge",
  "description": "Front row seats with complimentary lounge access",
  "color_code": "#8B5CF6"
}
```

---

## 4. Events & Ticketing

### 4.1 List Published Events
Retrieves all public event listings.

- **Endpoint**: `GET /api/v1/events`
- **Auth Required**: No
- **Query Parameters**:
  - `city` (string, optional): Filter by venue city.
  - `category` (string, optional): Filter by event category (`CONCERT`, `MOVIE`, `THEATRE`, `SPORTS`).
- **Response `200 OK`**: Array of published events with venue metadata and seat availability counts.

---

### 4.2 Get Event Seat Map
Retrieves the real-time seating grid and hold status for an event.

- **Endpoint**: `GET /api/v1/events/{id}/seats`
- **Auth Required**: Optional (If authenticated, `is_my_hold` indicates holds owned by caller).
- **Response `200 OK`**:
```json
[
  {
    "seat_id": "e1f2a3b4-...",
    "venue_seat_id": "v1v2v3v4-...",
    "row_label": "A",
    "seat_number": "1",
    "grid_row": 1,
    "grid_col": 1,
    "category_name": "VIP",
    "category_color": "#8B5CF6",
    "price": 240.00,
    "status": "AVAILABLE",
    "is_my_hold": false
  },
  {
    "seat_id": "e1f2a3b4-...",
    "row_label": "A",
    "seat_number": "2",
    "status": "HELD",
    "is_my_hold": true
  }
]
```

---

### 4.3 Place Seat Hold (10-Minute TTL)
Places a temporary hold on specified seats.

- **Endpoint**: `POST /api/v1/events/{id}/holds`
- **Auth Required**: Yes (`CUSTOMER`)
- **Request Body**:
```json
{
  "seat_ids": [
    "e1f2a3b4-...",
    "e1f2a3b5-..."
  ]
}
```
- **Response `200 OK`**:
```json
{
  "hold_token": "987fcb21-4567-8901-abcd-ef1234567890",
  "hold_ttl_seconds": 600,
  "expires_at": "2026-08-23T10:10:00Z",
  "seats": [ ... ]
}
```
- **Error Response `409 Conflict`**: Returned if any selected seat is already held or booked.

---

### 4.4 Release Seat Hold
Manually releases a hold before TTL expiration.

- **Endpoint**: `DELETE /api/v1/events/{id}/holds/{token}`
- **Auth Required**: Yes (`CUSTOMER`)
- **Response `200 OK`**: `{ "message": "Hold successfully released" }`

---

### 4.5 Checkout Booking
Confirms reservation of held seats and issues e-tickets.

- **Endpoint**: `POST /api/v1/events/{id}/bookings/checkout`
- **Auth Required**: Yes (`CUSTOMER`)
- **Request Body**:
```json
{
  "hold_token": "987fcb21-4567-8901-abcd-ef1234567890",
  "payment_method": "CREDIT_CARD",
  "billing_address": "123 Main St, City"
}
```
- **Response `201 Created`**:
```json
{
  "id": "b1b2b3b4-...",
  "booking_reference": "BK-20260823-7891",
  "status": "CONFIRMED",
  "total_amount": 487.00,
  "currency": "USD",
  "tickets": [
    {
      "id": "t1t2t3t4-...",
      "seat_number": "1",
      "row_label": "A",
      "category_name": "VIP",
      "qr_code_payload": "TICKET|t1t2t3t4-...|BK-20260823-7891|sig..."
    }
  ]
}
```

---

## 5. Waitlist & Automated Reallocation

### 5.1 Join Category Waitlist
Enters the caller into the FIFO queue for a specific category when an event is sold out.

- **Endpoint**: `POST /api/v1/events/{id}/waitlist`
- **Auth Required**: Yes (`CUSTOMER`)
- **Request Body**:
```json
{
  "seat_category_id": "00000000-0000-0000-0000-000000000001"
}
```
- **Response `201 Created`**:
```json
{
  "id": "w1w2w3w4-...",
  "event_id": "...",
  "seat_category_id": "...",
  "status": "WAITING",
  "queue_position": 1,
  "created_at": "2026-08-23T10:05:00Z"
}
```

---

### 5.2 List Customer Waitlist Entries
- **Endpoint**: `GET /api/v1/customers/waitlists`
- **Auth Required**: Yes (`CUSTOMER`)
- **Response `200 OK`**: Array of caller's waitlist entries, including current queue positions and any active time-limited offers.

---

### 5.3 Accept Waitlist Offer
Accepts a time-limited seat reallocation offer.

- **Endpoint**: `POST /api/v1/waitlist/offers/{token}/accept`
- **Auth Required**: Yes (`CUSTOMER`)
- **Response `200 OK`**: Booking confirmation and generated ticket passes.

---

## 6. Gate Ticket Verification & Check-In Scanner

### 6.1 Verify Ticket
Verifies ticket pass validity without checking in the attendee.

- **Endpoint**: `POST /api/v1/tickets/verify`
- **Auth Required**: Yes (`ORGANISER`, `ADMIN`)
- **Request Body**:
```json
{
  "qr_payload": "TICKET|t1t2t3t4-...|BK-20260823-7891|sig...",
  "ticket_id": ""
}
```
- **Response `200 OK`**: Verification status, customer details, seat assignment, and check-in history.

---

### 6.2 Check In Ticket
Performs atomic check-in at gate turnstiles.

- **Endpoint**: `POST /api/v1/tickets/check-in`
- **Auth Required**: Yes (`ORGANISER`, `ADMIN`)
- **Request Body**: Same as verify.
- **Response `200 OK`**: Confirmed check-in timestamp.
- **Error Response `409 Conflict`**: Returned if the ticket has already been checked in (anti-replay violation).

---

## 7. Organiser Analytics

### 7.1 Get Event Analytics
Retrieves live financial and operational metrics for an event.

- **Endpoint**: `GET /api/v1/organisers/events/{id}/analytics`
- **Auth Required**: Yes (`ORGANISER`, `ADMIN`)
- **Response `200 OK`**:
```json
{
  "event_id": "...",
  "total_capacity": 60,
  "booked_seats": 42,
  "held_seats": 8,
  "available_seats": 10,
  "occupancy_rate_percentage": 70.0,
  "gross_revenue": 5840.00,
  "currency": "USD",
  "category_breakdown": [
    {
      "category_name": "VIP",
      "capacity": 10,
      "booked": 10,
      "revenue": 2400.00
    }
  ],
  "waitlist_summary": {
    "total_waiting": 5,
    "active_offers": 1,
    "accepted_offers": 3,
    "expired_offers": 1
  }
}
```
