import {
  EventItem,
  SeatMapItem,
  HoldSeatsResponse,
  Booking,
  AuthResponse,
  EventAnalytics,
  User,
  TicketVerificationResult,
  EventTicketItem,
  EventCheckInOverview,
} from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

function getAuthHeader(): Record<string, string> {
  if (typeof window === "undefined") return {};
  const token = localStorage.getItem("velvet_auth_token");
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function handleResponse<T>(res: Response, fallbackErrorMessage: string): Promise<T> {
  let json: any = null;
  try {
    json = await res.json();
  } catch {
    if (!res.ok) {
      throw new Error(`${fallbackErrorMessage} (HTTP ${res.status}: ${res.statusText})`);
    }
    throw new Error(`Invalid JSON response from server (HTTP ${res.status})`);
  }

  if (!res.ok) {
    const errorMsg =
      json?.error?.message ||
      json?.error ||
      json?.message ||
      `${fallbackErrorMessage} (HTTP ${res.status})`;
    throw new Error(errorMsg);
  }

  return json.data !== undefined ? json.data : json;
}

export async function fetchEvents(params?: {
  event_type?: string;
  search?: string;
  limit?: number;
}): Promise<EventItem[]> {
  const url = new URL(`${API_BASE}/events`);
  if (params?.event_type && params.event_type !== "ALL") {
    url.searchParams.set("event_type", params.event_type);
  }
  if (params?.search) {
    url.searchParams.set("search", params.search);
  }
  if (params?.limit) {
    url.searchParams.set("limit", params.limit.toString());
  }

  const res = await fetch(url.toString(), {
    headers: { "Content-Type": "application/json" },
    signal: AbortSignal.timeout(5000),
  });

  const data = await handleResponse<EventItem[]>(res, "Failed to fetch events from backend");
  return Array.isArray(data) ? data : [];
}

export async function fetchEventById(id: string): Promise<EventItem> {
  const res = await fetch(`${API_BASE}/events/${id}`, {
    headers: { "Content-Type": "application/json" },
    signal: AbortSignal.timeout(5000),
  });

  return handleResponse<EventItem>(res, `Failed to fetch event details for ID ${id}`);
}

export async function fetchEventSeatMap(eventId: string): Promise<SeatMapItem[]> {
  const res = await fetch(`${API_BASE}/events/${eventId}/seats`, {
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    signal: AbortSignal.timeout(5000),
  });

  const data = await handleResponse<SeatMapItem[]>(res, `Failed to fetch seat map for event ${eventId}`);
  return Array.isArray(data) ? data : [];
}

export async function holdSeats(eventId: string, seatIds: string[]): Promise<HoldSeatsResponse> {
  const res = await fetch(`${API_BASE}/events/${eventId}/hold`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    body: JSON.stringify({ seat_ids: seatIds }),
    signal: AbortSignal.timeout(5000),
  });

  return handleResponse<HoldSeatsResponse>(res, "Failed to place seat hold");
}

export async function releaseHold(eventId: string, holdToken: string): Promise<boolean> {
  const res = await fetch(`${API_BASE}/events/${eventId}/release`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    body: JSON.stringify({ hold_token: holdToken }),
    signal: AbortSignal.timeout(5000),
  });

  await handleResponse<{ released: boolean }>(res, "Failed to release seat hold");
  return true;
}

export async function checkoutBooking(eventId: string, holdToken: string): Promise<Booking> {
  const res = await fetch(`${API_BASE}/bookings/checkout`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    body: JSON.stringify({ event_id: eventId, hold_token: holdToken }),
    signal: AbortSignal.timeout(6000),
  });

  return handleResponse<Booking>(res, "Checkout failed");
}

export async function fetchCustomerBookings(): Promise<Booking[]> {
  const res = await fetch(`${API_BASE}/customer/bookings`, {
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    signal: AbortSignal.timeout(5000),
  });

  const data = await handleResponse<Booking[]>(res, "Failed to fetch your bookings");
  return Array.isArray(data) ? data : [];
}

export async function cancelBooking(bookingId: string, reason?: string): Promise<boolean> {
  const res = await fetch(`${API_BASE}/customer/bookings/${bookingId}/cancel`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    body: JSON.stringify({ reason: reason || "Customer requested cancellation" }),
    signal: AbortSignal.timeout(5000),
  });

  await handleResponse<any>(res, `Failed to cancel booking ${bookingId}`);
  return true;
}

export async function fetchOrganiserEvents(): Promise<EventItem[]> {
  const res = await fetch(`${API_BASE}/organiser/events`, {
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    signal: AbortSignal.timeout(5000),
  });

  const data = await handleResponse<EventItem[]>(res, "Failed to fetch organiser events");
  return Array.isArray(data) ? data : [];
}

export async function fetchOrganiserAnalytics(eventId: string): Promise<EventAnalytics> {
  const res = await fetch(`${API_BASE}/organiser/events/${eventId}/analytics`, {
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    signal: AbortSignal.timeout(5000),
  });

  return handleResponse<EventAnalytics>(res, `Failed to fetch organiser analytics for event ${eventId}`);
}

export async function createOrganiserEvent(data: {
  venue_id: string;
  title: string;
  description?: string;
  event_type: string;
  poster_url?: string;
  start_time: string;
  end_time: string;
  hold_ttl_seconds?: number;
}): Promise<EventItem> {
  const res = await fetch(`${API_BASE}/organiser/events`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    body: JSON.stringify(data),
    signal: AbortSignal.timeout(6000),
  });

  return handleResponse<EventItem>(res, "Failed to create event");
}

export async function authLogin(email: string, pass: string): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password: pass }),
    signal: AbortSignal.timeout(5000),
  });

  return handleResponse<AuthResponse>(res, "Invalid email or password");
}

export async function authRegister(data: {
  email: string;
  password: string;
  full_name: string;
  phone?: string;
  role?: string;
}): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
    signal: AbortSignal.timeout(5000),
  });

  return handleResponse<AuthResponse>(res, "Registration failed");
}

export async function authMe(): Promise<User> {
  const res = await fetch(`${API_BASE}/auth/me`, {
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    signal: AbortSignal.timeout(5000),
  });

  return handleResponse<User>(res, "Failed to retrieve user profile");
}

export async function verifyTicket(payload: {
  qr_payload?: string;
  ticket_id?: string;
}): Promise<TicketVerificationResult> {
  const res = await fetch(`${API_BASE}/tickets/verify`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    body: JSON.stringify(payload),
    signal: AbortSignal.timeout(5000),
  });

  return handleResponse<TicketVerificationResult>(res, "Ticket verification failed");
}

export async function checkInTicket(payload: {
  qr_payload?: string;
  ticket_id?: string;
}): Promise<TicketVerificationResult> {
  const res = await fetch(`${API_BASE}/tickets/check-in`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    body: JSON.stringify(payload),
    signal: AbortSignal.timeout(5000),
  });

  return handleResponse<TicketVerificationResult>(res, "Ticket check-in failed");
}

export async function fetchEventTickets(
  eventId: string,
  status?: string
): Promise<EventCheckInOverview> {
  const url = new URL(`${API_BASE}/organiser/events/${eventId}/tickets`);
  if (status) {
    url.searchParams.set("status", status);
  }

  const res = await fetch(url.toString(), {
    headers: {
      "Content-Type": "application/json",
      ...getAuthHeader(),
    },
    signal: AbortSignal.timeout(5000),
  });

  return handleResponse<EventCheckInOverview>(res, "Failed to fetch event tickets");
}


