"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import {
  Calendar,
  MapPin,
  Ticket,
  QrCode,
  CheckCircle2,
  RefreshCw,
  XCircle,
  Clock,
  ArrowRight,
  ShieldCheck,
  AlertCircle,
  Lock,
  Sparkles,
  Users,
} from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { fetchCustomerBookings, cancelBooking, fetchCustomerWaitlists } from "../../lib/api";
import { Booking, WaitlistEntry } from "../../lib/types";
import { useAuth } from "../../context/AuthContext";

export default function MyBookingsPage() {
  const { user, isLoading: authLoading } = useAuth();
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [waitlists, setWaitlists] = useState<WaitlistEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedBookingForQR, setSelectedBookingForQR] = useState<Booking | null>(null);
  const [cancellingId, setCancellingId] = useState<string | null>(null);

  const loadBookings = async () => {
    if (!user) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const [bData, wData] = await Promise.all([
        fetchCustomerBookings(),
        fetchCustomerWaitlists().catch(() => []),
      ]);
      setBookings(bData);
      setWaitlists(wData);
    } catch (err: any) {
      setError(err.message || "Failed to load bookings from backend API");
      setBookings([]);
      setWaitlists([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!authLoading) {
      loadBookings();
    }
  }, [user, authLoading]);

  const handleCancelBooking = async (booking: Booking) => {
    const confirmCancel = window.confirm(
      `Are you sure you want to cancel booking ${booking.booking_reference}? A refund will be initiated according to venue policy.`
    );
    if (!confirmCancel) return;

    setCancellingId(booking.id);
    try {
      const ok = await cancelBooking(booking.id, "Customer requested cancellation");
      if (ok) {
        setBookings((prev) =>
          prev.map((b) =>
            b.id === booking.id
              ? { ...b, status: "CANCELLED", cancelled_at: new Date().toISOString() }
              : b
          )
        );
        alert(`✅ Booking ${booking.booking_reference} has been cancelled successfully.`);
      }
    } catch (err: any) {
      alert(`Failed to cancel booking: ${err.message || "Unknown error"}`);
    } finally {
      setCancellingId(null);
    }
  };

  if (authLoading) {
    return (
      <div className="py-24 text-center text-slate-400">
        <RefreshCw className="w-8 h-8 mx-auto animate-spin text-indigo-400 mb-3" />
        <p className="font-semibold text-slate-300">Checking authentication...</p>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-24 text-center space-y-6">
        <div className="glass-panel rounded-3xl border border-slate-800 p-8 sm:p-12 space-y-4 bg-[#131A26]">
          <div className="w-12 h-12 rounded-2xl bg-indigo-500/20 text-indigo-400 flex items-center justify-center mx-auto">
            <Lock className="w-6 h-6" />
          </div>
          <h2 className="text-xl font-bold text-white">Sign In Required</h2>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            Please sign in with your customer account to view your confirmed tickets, mobile passes, and reservation history.
          </p>
          <div className="pt-2">
            <Link
              href="/login"
              className="inline-flex items-center px-5 py-2.5 rounded-xl font-bold text-xs text-white bg-indigo-600 hover:bg-indigo-500 transition shadow-lg shadow-indigo-600/30"
            >
              Sign In to Your Account
              <ArrowRight className="w-3.5 h-3.5 ml-1.5" />
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-6">
        <div>
          <div className="flex items-center space-x-2 text-xs font-mono font-bold text-indigo-400 uppercase tracking-wider">
            <Ticket className="w-3.5 h-3.5" />
            <span>MY PASSES & RESERVATIONS</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-black text-white tracking-tight mt-1">
            My Tickets & Passes
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Manage your upcoming events, access digital passes, and view receipts.
          </p>
        </div>

        <Link
          href="/"
          className="px-5 py-2.5 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 shadow-md shadow-indigo-600/20 transition self-start sm:self-auto flex items-center"
        >
          Explore More Shows
          <ArrowRight className="w-3.5 h-3.5 ml-1.5" />
        </Link>
      </div>

      {/* Active Waitlist Entries & Seat Reallocation Offers Section */}
      {waitlists.length > 0 && (
        <div className="glass-panel rounded-3xl p-6 sm:p-8 border border-slate-800 bg-[#131A26] shadow-xl space-y-4">
          <div className="flex items-center justify-between border-b border-slate-800 pb-3">
            <div className="flex items-center space-x-2">
              <Users className="w-5 h-5 text-indigo-400" />
              <h3 className="text-base font-bold text-white">
                My Waitlist Entries & Reallocation Offers ({waitlists.length})
              </h3>
            </div>
            <span className="text-[11px] font-mono text-slate-400">
              FIFO Auto-Reallocation Active
            </span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {waitlists.map((w) => {
              const isOffered = w.status === "OFFERED";
              const isWaiting = w.status === "WAITING";

              return (
                <div
                  key={w.id}
                  className={`p-4 rounded-2xl border ${
                    isOffered
                      ? "bg-emerald-950/20 border-emerald-500/40 shadow-lg shadow-emerald-500/10 animate-pulse"
                      : "bg-[#0B0F17] border-slate-800"
                  } space-y-3 flex flex-col justify-between`}
                >
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between">
                      <span
                        className="text-[11px] font-bold font-mono px-2 py-0.5 rounded flex items-center"
                        style={{
                          backgroundColor: `${w.category_color || "#3B82F6"}25`,
                          color: w.category_color || "#3B82F6",
                          border: `1px solid ${w.category_color || "#3B82F6"}50`,
                        }}
                      >
                        {w.category_name || "Tier"} Category
                      </span>

                      <span
                        className={`text-[10px] font-mono font-bold px-2 py-0.5 rounded-full uppercase ${
                          isOffered
                            ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/40"
                            : isWaiting
                            ? "bg-indigo-500/20 text-indigo-300 border border-indigo-500/30"
                            : "bg-slate-800 text-slate-400"
                        }`}
                      >
                        {isOffered ? "🎉 SEAT OFFERED" : isWaiting ? `QUEUE POSITION #${w.queue_position}` : w.status}
                      </span>
                    </div>

                    <h4 className="text-sm font-bold text-white">
                      {w.event_title || "Sold-Out Performance"}
                    </h4>

                    {w.event_start_time && (
                      <p className="text-[11px] text-slate-400 flex items-center">
                        <Calendar className="w-3 h-3 mr-1 text-slate-500" />
                        {new Date(w.event_start_time).toLocaleDateString("en-US", {
                          weekday: "short",
                          month: "short",
                          day: "numeric",
                        })}
                      </p>
                    )}
                  </div>

                  {isOffered && (
                    <div className="pt-2 border-t border-emerald-500/20 flex items-center justify-between">
                      <span className="text-[10px] font-mono text-emerald-400 font-bold">
                        Time-limited offer active!
                      </span>
                      <Link
                        href={`/waitlist/offer?token=${w.offer_token || ""}&event=${w.event_id}`}
                        className="px-3 py-1.5 rounded-xl font-bold text-xs text-white bg-emerald-600 hover:bg-emerald-500 transition shadow-md shadow-emerald-600/30 flex items-center"
                      >
                        Claim Seat Now
                        <ArrowRight className="w-3 h-3 ml-1" />
                      </Link>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Bookings List / Error Banner */}
      {loading ? (
        <div className="py-20 text-center text-slate-400">
          <RefreshCw className="w-8 h-8 mx-auto animate-spin text-indigo-400 mb-3" />
          <p className="font-semibold text-slate-300">Loading your tickets & mobile passes from backend...</p>
        </div>
      ) : error ? (
        <div className="py-16 text-center glass-panel rounded-3xl border border-rose-500/40 p-8 space-y-4 bg-gradient-to-b from-rose-950/20 to-[#131A26]">
          <AlertCircle className="w-12 h-12 mx-auto text-rose-400" />
          <h3 className="text-lg font-bold text-rose-200">Failed to Load Your Bookings</h3>
          <p className="text-xs text-slate-300 max-w-md mx-auto font-mono bg-black/40 p-3 rounded-xl border border-rose-500/20 text-rose-300">
            {error}
          </p>
          <button
            onClick={loadBookings}
            className="mt-2 px-5 py-2.5 rounded-xl text-xs font-bold text-white bg-rose-600 hover:bg-rose-500 transition shadow-lg shadow-rose-600/30"
          >
            Retry Loading Bookings
          </button>
        </div>
      ) : bookings.length === 0 ? (
        <div className="py-20 text-center glass-panel rounded-3xl border border-slate-800 p-8 space-y-4">
          <Ticket className="w-12 h-12 mx-auto text-slate-600" />
          <h3 className="text-lg font-bold text-slate-200">No tickets found yet</h3>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            You haven&apos;t booked any shows yet. Explore our live concerts and cinema screenings to reserve seats with instant mobile passes.
          </p>
          <Link
            href="/"
            className="inline-block px-5 py-2.5 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-500 transition"
          >
            Explore Live Shows
          </Link>
        </div>
      ) : (
        <div className="space-y-6">
          {bookings.map((booking) => {
            const isConfirmed = booking.status === "CONFIRMED";

            return (
              <div
                key={booking.id}
                className="glass-panel rounded-3xl p-6 sm:p-8 border border-slate-800 bg-[#131A26] shadow-xl hover:border-slate-700 transition space-y-6"
              >
                {/* Header Row */}
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-4">
                  <div>
                    <div className="flex items-center gap-3">
                      <span className="text-xs font-mono font-bold text-indigo-400">
                        ORDER #{booking.booking_reference}
                      </span>
                      <span
                        className={`px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold uppercase tracking-wider flex items-center ${
                          isConfirmed
                            ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/30"
                            : "bg-rose-500/20 text-rose-300 border border-rose-500/30"
                        }`}
                      >
                        {isConfirmed ? (
                          <>
                            <CheckCircle2 className="w-3 h-3 mr-1" />
                            Active & Confirmed
                          </>
                        ) : (
                          <>
                            <XCircle className="w-3 h-3 mr-1" />
                            Cancelled
                          </>
                        )}
                      </span>
                    </div>

                    <h3 className="text-xl font-black text-white mt-1">
                      {booking.event_title || "VelvetSeats Event"}
                    </h3>
                  </div>

                  <div className="text-left sm:text-right">
                    <span className="text-xs text-slate-400 block font-mono">Total Paid</span>
                    <span className="text-lg font-black text-slate-100">
                      ${booking.total_amount.toFixed(2)}{" "}
                      <span className="text-xs text-slate-400 font-normal">{booking.currency}</span>
                    </span>
                  </div>
                </div>

                {/* Event Schedule & Location */}
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs text-slate-300 bg-[#0B0F17] p-4 rounded-2xl border border-slate-800">
                  <div className="flex items-center space-x-2.5">
                    <Calendar className="w-4 h-4 text-indigo-400 shrink-0" />
                    <div>
                      <span className="text-[10px] text-slate-500 block">EVENT DATE</span>
                      <span className="font-semibold text-slate-200">
                        {booking.event_start_time
                          ? new Date(booking.event_start_time).toLocaleDateString("en-US", {
                              weekday: "short",
                              month: "short",
                              day: "numeric",
                              year: "numeric",
                            })
                          : "Scheduled"}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center space-x-2.5">
                    <Clock className="w-4 h-4 text-violet-400 shrink-0" />
                    <div>
                      <span className="text-[10px] text-slate-500 block">SHOWTIME</span>
                      <span className="font-semibold text-slate-200">
                        {booking.event_start_time
                          ? new Date(booking.event_start_time).toLocaleTimeString("en-US", {
                              hour: "numeric",
                              minute: "2-digit",
                            })
                          : "TBA"}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center space-x-2.5">
                    <MapPin className="w-4 h-4 text-rose-400 shrink-0" />
                    <div>
                      <span className="text-[10px] text-slate-500 block">VENUE</span>
                      <span className="font-semibold text-slate-200">
                        {booking.venue_name || "Venue"}, {booking.venue_city || ""}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Ticket Details & Seats */}
                <div className="space-y-3">
                  <span className="text-xs font-semibold text-slate-400 uppercase font-mono tracking-wider block">
                    Reserved Seats ({booking.ticket_count || booking.tickets?.length || 0} Passes)
                  </span>

                  <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
                    {(booking.tickets || []).map((ticket) => (
                      <div
                        key={ticket.id}
                        className="p-3 rounded-xl bg-slate-900/80 border border-slate-800 flex items-center justify-between text-xs"
                      >
                        <div>
                          <span className="font-bold text-white font-mono block">
                            Row {ticket.row_label} • Seat {ticket.seat_number}
                          </span>
                          <span className="text-[10px] text-indigo-400 font-mono">
                            {ticket.category_name} (${ticket.unit_price})
                          </span>
                        </div>

                        <span className="text-[10px] px-2 py-0.5 rounded bg-slate-800 text-emerald-400 font-mono font-bold">
                          {isConfirmed ? "VALID" : "CANCELLED"}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="flex flex-wrap items-center justify-between gap-4 pt-2 border-t border-slate-800">
                  <div className="flex items-center space-x-2 text-[11px] text-slate-400 font-mono">
                    <ShieldCheck className="w-3.5 h-3.5 text-emerald-400" />
                    <span>Official Mobile Pass • Ready for Gate Admission</span>
                  </div>

                  <div className="flex items-center gap-3">
                    {isConfirmed && (
                      <>
                        <button
                          onClick={() => handleCancelBooking(booking)}
                          disabled={cancellingId === booking.id}
                          className="px-4 py-2 rounded-xl text-xs font-semibold text-rose-400 hover:text-rose-300 hover:bg-rose-500/10 border border-rose-500/30 transition flex items-center"
                        >
                          {cancellingId === booking.id ? (
                            <RefreshCw className="w-3.5 h-3.5 animate-spin mr-1.5" />
                          ) : (
                            <XCircle className="w-3.5 h-3.5 mr-1.5" />
                          )}
                          Cancel Booking
                        </button>

                        <button
                          onClick={() => setSelectedBookingForQR(booking)}
                          className="px-5 py-2 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-500 shadow-md shadow-indigo-600/25 transition flex items-center"
                        >
                          <QrCode className="w-3.5 h-3.5 mr-1.5" />
                          View Mobile Pass
                        </button>
                      </>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* QR Ticket Modal Dialog */}
      {selectedBookingForQR && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/85 backdrop-blur-md animate-in fade-in"
          onClick={() => setSelectedBookingForQR(null)}
        >
          <div
            className="w-full max-w-sm bg-[#131A26] border border-indigo-500/40 rounded-3xl shadow-2xl overflow-hidden p-6 space-y-5 text-center animate-in zoom-in-95"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex justify-between items-center border-b border-slate-800 pb-3">
              <span className="text-xs font-mono text-indigo-400 font-bold">
                ORDER #{selectedBookingForQR.booking_reference}
              </span>
              <button
                onClick={() => setSelectedBookingForQR(null)}
                className="text-slate-400 hover:text-white text-xs font-bold"
              >
                ✕
              </button>
            </div>

            <div className="p-4 bg-white rounded-2xl inline-block shadow-xl">
              <QRCodeSVG
                value={
                  selectedBookingForQR.tickets?.[0]?.qr_code_payload ||
                  selectedBookingForQR.booking_reference
                }
                size={180}
                level="H"
              />
            </div>

            <div className="space-y-1 text-xs">
              <h3 className="font-black text-white text-base">
                {selectedBookingForQR.event_title}
              </h3>
              <p className="text-slate-400">{selectedBookingForQR.venue_name}</p>
              <div className="pt-2 flex justify-center gap-1.5">
                {selectedBookingForQR.tickets?.map((t) => (
                  <span
                    key={t.id}
                    className="px-2 py-0.5 rounded bg-indigo-950 text-indigo-300 border border-indigo-800 text-[10px] font-mono font-bold"
                  >
                    {t.row_label}{t.seat_number}
                  </span>
                ))}
              </div>
            </div>

            <p className="text-[10px] text-slate-400 font-mono">
              Display this barcode on your phone at venue admission turnstiles for fast entry.
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
