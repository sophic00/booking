"use client";

import React, { useState, useEffect, use } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Calendar,
  MapPin,
  Clock,
  Ticket,
  ShieldCheck,
  Zap,
  RefreshCw,
  QrCode,
  ArrowLeft,
  CheckCircle2,
  AlertTriangle,
  Info,
  X,
} from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import confetti from "canvas-confetti";
import { fetchEventById, fetchEventSeatMap, holdSeats, releaseHold, checkoutBooking } from "../../../lib/api";
import { EventItem, SeatMapItem, Booking, HoldSeatsResponse } from "../../../lib/types";
import { useAuth } from "../../../context/AuthContext";

export default function EventDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const resolvedParams = use(params);
  const eventId = resolvedParams.id;
  const router = useRouter();
  const { user } = useAuth();

  const [event, setEvent] = useState<EventItem | null>(null);
  const [seatMap, setSeatMap] = useState<SeatMapItem[]>([]);
  const [selectedSeatIds, setSelectedSeatIds] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);

  // Hold State
  const [holdData, setHoldData] = useState<HoldSeatsResponse | null>(null);
  const [holdSecondsRemaining, setHoldSecondsRemaining] = useState<number | null>(null);
  const [isHolding, setIsHolding] = useState(false);

  // Checkout State
  const [isCheckingOut, setIsCheckingOut] = useState(false);
  const [confirmedBooking, setConfirmedBooking] = useState<Booking | null>(null);
  const [showQRModal, setShowQRModal] = useState(false);

  // Waitlist State
  const [waitlistJoined, setWaitlistJoined] = useState(false);

  // Load Event & Seat Map
  useEffect(() => {
    async function loadData() {
      setLoading(true);
      const ev = await fetchEventById(eventId);
      const seats = await fetchEventSeatMap(eventId);
      setEvent(ev);
      setSeatMap(seats);
      setLoading(false);
    }
    loadData();
  }, [eventId]);

  // Hold TTL Countdown Timer
  useEffect(() => {
    if (holdSecondsRemaining === null || holdSecondsRemaining <= 0) return;

    const interval = setInterval(() => {
      setHoldSecondsRemaining((prev) => {
        if (prev === null || prev <= 1) {
          clearInterval(interval);
          // Auto release hold on TTL expiry
          if (holdData) {
            releaseHold(eventId, holdData.hold_token);
            setHoldData(null);
            alert("⏰ Your 10-minute seat hold has expired. The seats have been auto-released.");
          }
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [holdSecondsRemaining, holdData, eventId]);

  if (loading || !event) {
    return (
      <div className="min-h-[70vh] flex flex-col items-center justify-center space-y-4">
        <RefreshCw className="w-10 h-10 animate-spin text-indigo-500" />
        <p className="text-sm font-semibold text-slate-300">Loading interactive seat layout...</p>
      </div>
    );
  }

  const selectedSeats = seatMap.filter((s) => selectedSeatIds.includes(s.seat_id));
  const subtotal = selectedSeats.reduce((sum, s) => sum + s.price, 0);
  const bookingFee = selectedSeats.length > 0 ? selectedSeats.length * 3.5 : 0;
  const grandTotal = subtotal + bookingFee;
  const isSoldOut = event.seats_left === 0 || event.badge?.includes("Sold Out");

  const toggleSeat = (seat: SeatMapItem) => {
    if (seat.status !== "AVAILABLE" && !seat.is_my_hold) return;

    setSelectedSeatIds((prev) => {
      if (prev.includes(seat.seat_id)) {
        return prev.filter((id) => id !== seat.seat_id);
      } else {
        if (prev.length >= 8) {
          alert("Maximum 8 seats can be reserved in a single transaction.");
          return prev;
        }
        return [...prev, seat.seat_id];
      }
    });
  };

  const handleHoldSeats = async () => {
    if (selectedSeatIds.length === 0) return;
    setIsHolding(true);
    try {
      const res = await holdSeats(eventId, selectedSeatIds);
      setHoldData(res);
      setHoldSecondsRemaining(res.hold_ttl_seconds || 600);
    } catch (err: any) {
      alert(err.message || "Failed to hold seats");
    } finally {
      setIsHolding(false);
    }
  };

  const handleConfirmCheckout = async () => {
    if (selectedSeats.length === 0) return;
    setIsCheckingOut(true);

    try {
      const token = holdData ? holdData.hold_token : "demo-hold-" + Date.now();
      const booking = await checkoutBooking(eventId, token, selectedSeats, event);
      setConfirmedBooking(booking);
      setShowQRModal(true);

      // Launch Celebratory Confetti
      try {
        confetti({
          particleCount: 100,
          spread: 70,
          origin: { y: 0.6 },
        });
      } catch {
        // ignore
      }
    } catch (err: any) {
      alert(err.message || "Checkout failed");
    } finally {
      setIsCheckingOut(false);
    }
  };

  const formatTime = (secs: number) => {
    const m = Math.floor(secs / 60);
    const s = secs % 60;
    return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
  };

  // Group seats by row
  const rowsMap = seatMap.reduce((acc, seat) => {
    if (!acc[seat.row_label]) acc[seat.row_label] = [];
    acc[seat.row_label].push(seat);
    return acc;
  }, {} as Record<string, SeatMapItem[]>);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Top Header & Breadcrumb */}
      <div className="flex items-center justify-between">
        <Link
          href="/"
          className="inline-flex items-center text-xs font-semibold text-slate-400 hover:text-indigo-300 transition"
        >
          <ArrowLeft className="w-3.5 h-3.5 mr-1.5" />
          Back to Events
        </Link>
        <span className="text-xs font-mono text-slate-400 bg-slate-900/80 px-3 py-1 rounded-full border border-slate-800">
          EVENT ID: {event.id}
        </span>
      </div>

      {/* Event Overview Hero Banner */}
      <div className="relative overflow-hidden rounded-3xl border border-slate-800 bg-[#131A26] p-6 sm:p-8 shadow-2xl flex flex-col md:flex-row gap-6 items-start md:items-center justify-between">
        <div className="flex items-center space-x-6">
          <div className="w-24 h-32 sm:w-28 sm:h-36 rounded-2xl bg-slate-900 overflow-hidden shrink-0 border border-slate-700 shadow-lg">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={event.poster_url || ""}
              alt={event.title}
              className="w-full h-full object-cover"
            />
          </div>

          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <span className="px-2.5 py-0.5 rounded-md bg-indigo-500/20 text-indigo-300 text-[11px] font-bold uppercase tracking-wider border border-indigo-500/30">
                {event.event_type}
              </span>
              {event.badge && (
                <span className="px-2.5 py-0.5 rounded-md bg-rose-500/20 text-rose-300 text-[11px] font-bold border border-rose-500/30">
                  {event.badge}
                </span>
              )}
            </div>

            <h1 className="text-2xl sm:text-3xl font-black text-white tracking-tight">
              {event.title}
            </h1>

            <div className="flex flex-wrap items-center gap-4 text-xs text-slate-300">
              <span className="flex items-center">
                <Calendar className="w-3.5 h-3.5 mr-1 text-indigo-400" />
                {new Date(event.start_time).toLocaleDateString("en-US", {
                  weekday: "short",
                  month: "short",
                  day: "numeric",
                  year: "numeric",
                })}
              </span>
              <span className="flex items-center">
                <Clock className="w-3.5 h-3.5 mr-1 text-violet-400" />
                {new Date(event.start_time).toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit" })}
              </span>
              <span className="flex items-center">
                <MapPin className="w-3.5 h-3.5 mr-1 text-rose-400" />
                {event.venue_name}, {event.venue_city}
              </span>
            </div>
          </div>
        </div>

        {/* Pricing Quick Bar */}
        <div className="bg-[#0B0F17] p-4 rounded-2xl border border-slate-800 text-right shrink-0 w-full md:w-auto">
          <span className="text-[11px] text-slate-400 block uppercase tracking-wider font-mono">
            Pricing Tiers
          </span>
          <div className="flex md:flex-col gap-3 md:gap-1 mt-1 text-xs">
            <span className="text-purple-400 font-bold">VIP: $240+</span>
            <span className="text-indigo-400 font-bold">Premium: $140</span>
            <span className="text-blue-400 font-bold">Standard: $65</span>
          </div>
        </div>
      </div>

      {/* Active Hold Countdown Banner */}
      {holdSecondsRemaining !== null && holdSecondsRemaining > 0 && (
        <div className="glass-panel rounded-2xl p-4 border border-amber-500/40 bg-gradient-to-r from-amber-500/10 via-[#131A26] to-amber-500/10 shadow-lg flex flex-col sm:flex-row items-center justify-between gap-4 animate-in fade-in">
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 rounded-xl bg-amber-500/20 text-amber-400 flex items-center justify-center shrink-0">
              <Zap className="w-5 h-5 animate-pulse" />
            </div>
            <div>
              <h4 className="text-sm font-bold text-amber-300">
                10-Minute Fair Hold Active ({selectedSeats.length} seats locked)
              </h4>
              <p className="text-xs text-slate-400">
                Your seats are reserved and hidden from all other customers. Complete checkout before timer expires.
              </p>
            </div>
          </div>

          <div className="flex items-center space-x-3">
            <div className="px-4 py-2 rounded-xl bg-black/60 border border-amber-500/50 text-amber-300 font-mono font-black text-xl tracking-wider shadow-inner">
              {formatTime(holdSecondsRemaining)}
            </div>
            <button
              onClick={handleConfirmCheckout}
              disabled={isCheckingOut}
              className="px-5 py-2.5 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 shadow-lg shadow-emerald-600/30 transition flex items-center shrink-0"
            >
              {isCheckingOut ? <RefreshCw className="w-4 h-4 animate-spin" /> : "Confirm Now"}
            </button>
          </div>
        </div>
      )}

      {/* Main Seat Map & Checkout Columns */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* ========================================================
            LEFT COLUMN: INTERACTIVE VISUAL SEAT MAP
            ======================================================== */}
        <div className="lg:col-span-8 glass-panel rounded-3xl p-6 sm:p-8 border border-slate-800 bg-[#131A26] space-y-6">
          <div className="flex items-center justify-between border-b border-slate-800 pb-4">
            <div>
              <h3 className="text-lg font-bold text-white flex items-center">
                <Ticket className="w-5 h-5 mr-2 text-indigo-400" />
                Select Your Seats
              </h3>
              <p className="text-xs text-slate-400 mt-0.5">
                Click available seats to add to your reservation hold.
              </p>
            </div>

            {/* Seat Map Legend */}
            <div className="flex flex-wrap items-center gap-3 text-[11px] font-medium text-slate-300">
              <div className="flex items-center space-x-1.5">
                <span className="w-3.5 h-3.5 rounded bg-purple-500" />
                <span>VIP ($240)</span>
              </div>
              <div className="flex items-center space-x-1.5">
                <span className="w-3.5 h-3.5 rounded bg-indigo-500" />
                <span>Premium ($140)</span>
              </div>
              <div className="flex items-center space-x-1.5">
                <span className="w-3.5 h-3.5 rounded bg-blue-500" />
                <span>Standard ($65)</span>
              </div>
              <div className="flex items-center space-x-1.5">
                <span className="w-3.5 h-3.5 rounded bg-slate-700 opacity-40" />
                <span>Occupied</span>
              </div>
            </div>
          </div>

          {/* Screen / Stage Curved Visual Indicator */}
          <div className="py-2 text-center space-y-2">
            <div className="h-2 w-3/4 mx-auto rounded-full bg-gradient-to-r from-transparent via-indigo-500 to-transparent shadow-[0_0_20px_rgba(99,102,241,0.8)]" />
            <span className="text-[10px] font-mono uppercase tracking-widest text-slate-500">
              STAGE / SCREEN DIRECTION
            </span>
          </div>

          {/* Visual Seat Grid Container */}
          <div className="overflow-x-auto py-6 flex flex-col items-center gap-2.5">
            {Object.entries(rowsMap).map(([rowLabel, seats]) => (
              <div key={rowLabel} className="flex items-center gap-2">
                <span className="w-6 text-center text-xs font-mono font-bold text-slate-500">
                  {rowLabel}
                </span>

                <div className="flex items-center gap-1.5">
                  {seats.map((seat) => {
                    const isSelected = selectedSeatIds.includes(seat.seat_id);
                    const isBooked = seat.status === "BOOKED";
                    const isHeldByOther = seat.status === "HELD" && !seat.is_my_hold;

                    let bgClass = "bg-blue-600/70 hover:bg-blue-500";
                    if (seat.seat_category_id === "cat-vip") {
                      bgClass = "bg-purple-600/80 hover:bg-purple-500";
                    } else if (seat.seat_category_id === "cat-prem") {
                      bgClass = "bg-indigo-600/80 hover:bg-indigo-500";
                    }

                    if (isSelected) {
                      bgClass = "bg-emerald-500 ring-2 ring-white scale-110 shadow-lg shadow-emerald-500/50";
                    } else if (isBooked) {
                      bgClass = "bg-slate-800/40 text-slate-600 cursor-not-allowed border border-slate-800";
                    } else if (isHeldByOther) {
                      bgClass = "bg-amber-600/30 text-amber-500/50 cursor-not-allowed border border-amber-500/20";
                    }

                    return (
                      <button
                        key={seat.seat_id}
                        disabled={isBooked || isHeldByOther}
                        onClick={() => toggleSeat(seat)}
                        title={`${rowLabel}${seat.seat_number} - ${seat.category_name} ($${seat.price}) - ${seat.status}`}
                        className={`w-7 h-7 sm:w-8 sm:h-8 rounded-lg text-[10px] font-bold font-mono transition-all flex items-center justify-center ${bgClass}`}
                      >
                        {isSelected ? "✓" : seat.seat_number}
                      </button>
                    );
                  })}
                </div>

                <span className="w-6 text-center text-xs font-mono font-bold text-slate-500">
                  {rowLabel}
                </span>
              </div>
            ))}
          </div>

          {/* Sold Out & Waitlist Banner (if applicable) */}
          {isSoldOut && (
            <div className="p-4 rounded-2xl bg-rose-500/10 border border-rose-500/30 flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <AlertTriangle className="w-5 h-5 text-rose-400" />
                <div>
                  <h4 className="text-xs font-bold text-rose-200">
                    This show is currently Sold Out
                  </h4>
                  <p className="text-[11px] text-slate-400">
                    Join the automated reallocation waitlist to be notified first on cancellation.
                  </p>
                </div>
              </div>
              <button
                onClick={() => setWaitlistJoined(true)}
                disabled={waitlistJoined}
                className="px-4 py-2 rounded-xl text-xs font-bold text-white bg-rose-600 hover:bg-rose-500 transition disabled:opacity-50"
              >
                {waitlistJoined ? "Joined (#14 in Queue)" : "Join Waitlist"}
              </button>
            </div>
          )}
        </div>

        {/* ========================================================
            RIGHT COLUMN: ORDER SUMMARY & CHECKOUT DRAWER
            ======================================================== */}
        <div className="lg:col-span-4 glass-panel rounded-3xl p-6 border border-slate-800 bg-[#131A26] space-y-6">
          <div className="border-b border-slate-800 pb-4">
            <h3 className="text-base font-bold text-white">Booking Summary</h3>
            <p className="text-xs text-slate-400 mt-0.5">
              Review selected seats and confirm your reservation.
            </p>
          </div>

          {/* Selected Seats List */}
          <div className="space-y-2.5">
            <span className="text-xs font-semibold text-slate-400 uppercase font-mono tracking-wider block">
              Selected Seats ({selectedSeats.length})
            </span>

            {selectedSeats.length === 0 ? (
              <div className="py-8 text-center border border-dashed border-slate-800 rounded-2xl text-slate-500 text-xs">
                No seats selected yet. Click any available seat on the grid.
              </div>
            ) : (
              <div className="space-y-2 max-h-48 overflow-y-auto">
                {selectedSeats.map((seat) => (
                  <div
                    key={seat.seat_id}
                    className="flex items-center justify-between p-2.5 rounded-xl bg-[#0B0F17] border border-slate-800 text-xs"
                  >
                    <div className="flex items-center space-x-2">
                      <span className="w-2.5 h-2.5 rounded-full bg-indigo-400" />
                      <span className="font-bold text-slate-200 font-mono">
                        Row {seat.row_label} - Seat {seat.seat_number}
                      </span>
                      <span className="text-[10px] text-slate-400 font-mono">
                        ({seat.category_name})
                      </span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <span className="font-bold text-slate-100">${seat.price}</span>
                      <button
                        onClick={() => toggleSeat(seat)}
                        className="text-slate-500 hover:text-rose-400 p-0.5"
                      >
                        <X className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Price Breakdown */}
          {selectedSeats.length > 0 && (
            <div className="space-y-2 pt-4 border-t border-slate-800 text-xs text-slate-300">
              <div className="flex justify-between">
                <span className="text-slate-400">Seat Subtotal</span>
                <span className="font-medium text-slate-200">${subtotal.toFixed(2)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Service Fee ($3.50/tkt)</span>
                <span className="font-medium text-slate-200">${bookingFee.toFixed(2)}</span>
              </div>
              <div className="flex justify-between pt-2 border-t border-slate-800 text-sm font-bold text-white">
                <span>Grand Total</span>
                <span className="text-emerald-400 font-mono">${grandTotal.toFixed(2)}</span>
              </div>
            </div>
          )}

          {/* Action Buttons */}
          <div className="space-y-2.5 pt-2">
            {!holdData ? (
              <button
                onClick={handleHoldSeats}
                disabled={selectedSeats.length === 0 || isHolding}
                className="w-full py-3 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 shadow-xl shadow-indigo-600/30 transition flex items-center justify-center disabled:opacity-40"
              >
                {isHolding ? (
                  <RefreshCw className="w-4 h-4 animate-spin" />
                ) : (
                  <>
                    <Zap className="w-4 h-4 mr-1.5" />
                    Place 10-Min Hold & Checkout
                  </>
                )}
              </button>
            ) : (
              <button
                onClick={handleConfirmCheckout}
                disabled={isCheckingOut}
                className="w-full py-3.5 rounded-xl font-bold text-xs sm:text-sm text-white bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 shadow-xl shadow-emerald-600/30 transition flex items-center justify-center disabled:opacity-40"
              >
                {isCheckingOut ? (
                  <RefreshCw className="w-4 h-4 animate-spin" />
                ) : (
                  <>
                    <CheckCircle2 className="w-4 h-4 mr-2" />
                    Confirm Booking (${grandTotal.toFixed(2)})
                  </>
                )}
              </button>
            )}

            <div className="flex items-center justify-center space-x-1.5 text-[11px] text-slate-500 font-mono pt-1">
              <ShieldCheck className="w-3.5 h-3.5 text-emerald-400" />
              <span>Instant QR E-Ticket Dispatch</span>
            </div>
          </div>
        </div>
      </div>

      {/* ========================================================
          CONFIRMATION & QR E-TICKET MODAL DIALOG
          ======================================================== */}
      {showQRModal && confirmedBooking && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/85 backdrop-blur-md animate-in fade-in">
          <div className="w-full max-w-md bg-[#131A26] border border-indigo-500/40 rounded-3xl shadow-2xl overflow-hidden p-6 sm:p-8 space-y-6 text-center animate-in zoom-in-95">
            <div className="w-12 h-12 rounded-2xl bg-emerald-500/20 text-emerald-400 flex items-center justify-center mx-auto shadow-lg shadow-emerald-500/20">
              <CheckCircle2 className="w-6 h-6" />
            </div>

            <div className="space-y-1">
              <h2 className="text-xl font-black text-white">Booking Confirmed!</h2>
              <p className="text-xs text-slate-400">
                Your seats have been permanently locked and scannable QR tickets are generated.
              </p>
            </div>

            {/* Visual Scannable QR Ticket Card */}
            <div className="bg-[#0B0F17] rounded-2xl p-5 border border-slate-800 space-y-4">
              <div className="flex justify-between items-center text-xs font-mono border-b border-slate-800 pb-2">
                <span className="text-slate-400">REF:</span>
                <span className="text-indigo-400 font-bold">{confirmedBooking.booking_reference}</span>
              </div>

              {/* Scannable SVG QR Code */}
              <div className="p-3 bg-white rounded-xl inline-block shadow-md">
                <QRCodeSVG
                  value={confirmedBooking.tickets?.[0]?.qr_code_payload || confirmedBooking.booking_reference}
                  size={160}
                  level="H"
                />
              </div>

              <div className="space-y-1 text-xs">
                <h4 className="font-bold text-white">{event.title}</h4>
                <p className="text-slate-400">{event.venue_name}</p>
                <div className="pt-2 flex justify-center gap-2">
                  {confirmedBooking.tickets?.map((t) => (
                    <span
                      key={t.id}
                      className="px-2.5 py-1 rounded bg-indigo-950 text-indigo-300 border border-indigo-800 text-[11px] font-mono font-bold"
                    >
                      {t.row_label}{t.seat_number} ({t.category_name})
                    </span>
                  ))}
                </div>
              </div>
            </div>

            {/* Actions */}
            <div className="flex gap-3">
              <button
                onClick={() => {
                  setShowQRModal(false);
                  router.push("/my-bookings");
                }}
                className="flex-1 py-3 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 shadow-lg shadow-indigo-600/30 transition"
              >
                View in My Bookings
              </button>
              <button
                onClick={() => setShowQRModal(false)}
                className="px-4 py-3 rounded-xl font-semibold text-xs text-slate-300 hover:text-white bg-slate-900 border border-slate-700 transition"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
