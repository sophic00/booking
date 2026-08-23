export type UserRole = "CUSTOMER" | "ORGANISER" | "ADMIN";

export interface User {
  id: string;
  email: string;
  full_name: string;
  phone?: string | null;
  role: UserRole;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  expires_at: string;
  user: User;
}

export type EventType = "MOVIE" | "CONCERT" | "THEATRE" | "SPORTS" | "OTHER";
export type EventStatus = "DRAFT" | "PUBLISHED" | "CANCELLED" | "COMPLETED";

export interface EventPricing {
  id: string;
  event_id: string;
  seat_category_id: string;
  category_name?: string;
  category_description?: string | null;
  category_color?: string;
  price: number;
  currency: string;
}

export interface EventItem {
  id: string;
  organiser_id: string;
  organiser_name?: string;
  organiser_email?: string;
  venue_id: string;
  venue_name?: string;
  venue_address?: string;
  venue_city?: string;
  venue_capacity?: number;
  title: string;
  description?: string | null;
  event_type: EventType;
  poster_url?: string | null;
  banner_url?: string | null;
  start_time: string;
  end_time: string;
  hold_ttl_seconds: number;
  status: EventStatus;
  pricing?: EventPricing[];
  created_at: string;
  updated_at: string;
  // UI helper fields
  badge?: string;
  badge_type?: "hot" | "urgent" | "sold_out" | "available" | "warning";
  min_price?: number;
  max_price?: number;
  seats_left?: number;
  waitlist_count?: number;
}

export type SeatStatus = "AVAILABLE" | "HELD" | "OFFERED" | "BOOKED";

export interface SeatMapItem {
  seat_id: string;
  venue_id: string;
  seat_category_id: string;
  category_name: string;
  category_color: string;
  row_label: string;
  seat_number: string;
  grid_row: number;
  grid_col: number;
  price: number;
  currency: string;
  status: SeatStatus;
  held_by_user_id?: string | null;
  hold_expires_at?: string | null;
  is_my_hold: boolean;
}

export interface HeldSeatDetail {
  seat_id: string;
  seat_category_id: string;
  price: number;
  currency: string;
}

export interface HoldSeatsResponse {
  hold_token: string;
  event_id: string;
  expires_at: string;
  hold_ttl_seconds: number;
  seat_count: number;
  total_price: number;
  currency: string;
  seats: HeldSeatDetail[];
}

export interface TicketDetail {
  id: string;
  booking_id: string;
  seat_id: string;
  row_label: string;
  seat_number: string;
  grid_row: number;
  grid_col: number;
  seat_category_id: string;
  category_name: string;
  category_color: string;
  unit_price: number;
  qr_code_payload: string;
  qr_code_data_url?: string;
  status: string;
}

export interface Booking {
  id: string;
  booking_reference: string;
  customer_id: string;
  customer_name?: string;
  customer_email?: string;
  event_id: string;
  event_title?: string;
  event_start_time?: string;
  venue_name?: string;
  venue_city?: string;
  total_amount: number;
  currency: string;
  status: "CONFIRMED" | "CANCELLED" | "REFUNDED";
  cancellation_reason?: string | null;
  cancelled_at?: string | null;
  ticket_count?: number;
  tickets?: TicketDetail[];
  created_at: string;
}

export interface Venue {
  id: string;
  name: string;
  address: string;
  city: string;
  state?: string | null;
  country: string;
  total_capacity: number;
  created_by?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateVenuePayload {
  name: string;
  address: string;
  city: string;
  state?: string;
  country?: string;
  total_capacity?: number;
}

export interface UpdateVenuePayload {
  name: string;
  address: string;
  city: string;
  state?: string;
  country: string;
  total_capacity: number;
}

export interface SeatCategory {
  id: string;
  name: string;
  description?: string | null;
  color_code: string;
  created_at: string;
}

export interface CreateCategoryPayload {
  name: string;
  description?: string;
  color_code?: string;
}

export interface VenueSeat {
  id: string;
  venue_id: string;
  seat_category_id: string;
  category_name?: string;
  category_color?: string;
  row_label: string;
  seat_number: string;
  grid_row: number;
  grid_col: number;
  is_active: boolean;
  created_at: string;
}

export interface CreateSeatPayload {
  seat_category_id: string;
  row_label: string;
  seat_number: string;
  grid_row: number;
  grid_col: number;
  is_active?: boolean;
}

export interface BatchCreateSeatsPayload {
  replace: boolean;
  seats: CreateSeatPayload[];
}

export interface CategoryPricingPayloadItem {
  seat_category_id: string;
  price: number;
  currency?: string;
}

export interface WaitlistEntry {
  id: string;
  event_id: string;
  event_title?: string;
  seat_category_id: string;
  category_name?: string;
  category_color?: string;
  status: "WAITING" | "OFFERED" | "ACCEPTED" | "EXPIRED" | "CANCELLED";
  queue_position: number;
  offer_token?: string;
  offer_expires_at?: string;
  event_start_time?: string;
  created_at: string;
}

export interface WaitlistOfferDetail {
  id: string;
  waitlist_entry_id: string;
  event_id: string;
  event_title: string;
  event_start_time?: string;
  event_end_time?: string;
  seat_id: string;
  row_label: string;
  seat_number: string;
  seat_category_id: string;
  category_name: string;
  price: number;
  currency: string;
  offer_token: string;
  offered_at: string;
  expires_at: string;
  status: string;
}

export interface CategoryBreakdown {
  seat_category_id: string;
  category_name: string;
  color_code: string;
  total_seats: number;
  booked_seats: number;
  revenue: number;
  waitlist_count: number;
}

export interface EventAnalytics {
  event_id: string;
  event_title: string;
  event_status: string;
  start_time: string;
  total_capacity: number;
  confirmed_bookings_count: number;
  cancelled_bookings_count: number;
  valid_tickets_count: number;
  checked_in_tickets_count?: number;
  total_revenue: number;
  occupancy_percentage: number;
  waitlist_waiting_count: number;
  category_breakdown: CategoryBreakdown[];
}

export interface TicketVerificationResult {
  is_valid: boolean;
  can_check_in: boolean;
  validation_message: string;
  ticket: {
    id: string;
    status: "VALID" | "CHECKED_IN" | "CANCELLED" | string;
    unit_price: number;
    qr_code_payload: string;
    created_at: string;
    checked_in_at?: string | null;
  };
  booking: {
    id: string;
    booking_reference: string;
    status: string;
    customer_id: string;
    customer_name: string;
    customer_email: string;
  };
  event: {
    id: string;
    organiser_id: string;
    title: string;
    start_time: string;
    end_time: string;
    status: string;
    poster_url?: string | null;
  };
  venue: {
    id: string;
    name: string;
    address: string;
    city: string;
  };
  seat: {
    id: string;
    row_label: string;
    seat_number: string;
    grid_row: number;
    grid_col: number;
    seat_category_id: string;
    category_name: string;
    category_color: string;
  };
}

export interface EventTicketItem {
  id: string;
  booking_id: string;
  booking_reference: string;
  customer_id: string;
  customer_name: string;
  customer_email: string;
  seat_id: string;
  row_label: string;
  seat_number: string;
  category_name: string;
  category_color: string;
  unit_price: number;
  qr_code_payload: string;
  status: "VALID" | "CHECKED_IN" | "CANCELLED" | string;
  created_at: string;
  checked_in_at?: string | null;
}

export interface EventCheckInOverview {
  event_id: string;
  total_tickets: number;
  valid_count: number;
  checked_in_count: number;
  cancelled_count: number;
  check_in_rate: number;
  tickets?: EventTicketItem[];
}

