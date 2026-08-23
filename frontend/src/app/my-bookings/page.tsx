"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import {
  Calendar,
  MapPin,
  Ticket,
  QrCode,
  AlertTriangle,
  CheckCircle2,
  RefreshCw,
  XCircle,
  Clock,
  ArrowRight,
  ShieldCheck,
} from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import { fetchCustomerBookings, cancelBooking } from "../../lib/api";
import { Booking } from "../../lib/types";

export default function MyBookingsPage() {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedBookingForQR, setSelectedBookingForQR] = useState<Booking | null>(null);
  const [cancellingId, setCancellingId] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      setLoading(true);
      const data = await fetchCustomerBookings();
      setBookings(data);
      setLoading(false);
    }
    load();
  }, []);

  const handleCancelBooking = async (booking: Booking) => {
    const confirmCancel = window.confirm(
      `Are you sure you want to cancel booking ${booking.booking_reference}? Your seats will be automatically reallocated to the next fan on the waitlist.`
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
        alert(
          `✅ Booking ${booking.booking_reference} has been cancelled. Automated waitlist engine has been notified to reallocate the seats.`
        );
      }
    } catch {
      alert("Failed to cancel booking.");
    } finally {
      setCancellingId(null);
    }
  };

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
            My Booking History & QR Tickets
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Manage your confirmed shows, view encrypted scannable QR passes, or cancel bookings.
          </p>
        </div>

        <Link
          href="/"
          className="px-5 py-2.5 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 shadow-md shadow-indigo-600/20 transition self-start sm:self-auto flex items-center"
        >
          Book Another Show
          <ArrowRight className="w-3.5 h-3.5 ml-1.5" />
        </Link>
      </div>

      {/* Bookings List */}
      {loading ? (
        <div className="py-20 text-center text-slate-400">
          <RefreshCw className="w-8 h-8 mx-auto animate-spin text-indigo-400 mb-3" />
          <p className="font-semibold text-slate-300">Loading your passes & ticket codes...</p>
        </div>
      ) : bookings.length === 0 ? (
        <div className="py-20 text-center glass-panel rounded-3xl border border-slate-800 p-8 space-y-4">
          <Ticket className="w-12 h-12 mx-auto text-slate-600" />
          <h3 className="text-lg font-bold text-slate-200">No bookings found yet</h3>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            You haven&apos;t reserved any shows yet. Explore our live concerts and cinema screenings to reserve seats with a 10-minute hold guarantee.
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
                        REF: {booking.booking_reference}
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
                            Confirmed & Active
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
                      {booking.event_title || "VelvetSeats Special Show"}
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
                          : "Sep 26, 2026"}
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
                          : "7:00 PM"}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center space-x-2.5">
                    <MapPin className="w-4 h-4 text-rose-400 shrink-0" />
                    <div>
                      <span className="text-[10px] text-slate-500 block">VENUE</span>
                      <span className="font-semibold text-slate-200">
                        {booking.venue_name || "Royal Albert Hall"}, {booking.venue_city || "London"}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Ticket Details & Seats */}
                <div className="space-y-3">
                  <span className="text-xs font-semibold text-slate-400 uppercase font-mono tracking-wider block">
                    Reserved Seats ({booking.ticket_count || booking.tickets?.length || 2} Tickets)
                  </span>

                  <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
                    {(booking.tickets || [
                      { id: "1", row_label: "A", seat_number: "14", category_name: "VIP Diamond", unit_price: 240, qr_code_payload: "TB:MOCK" },
                      { id: "2", row_label: "A", seat_number: "15", category_name: "VIP Diamond", unit_price: 240, qr_code_payload: "TB:MOCK" },
                    ]).map((ticket: any) => (
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
                          {isConfirmed ? "VALID" : "VOID"}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="flex flex-wrap items-center justify-between gap-4 pt-2 border-t border-slate-800">
                  <div className="flex items-center space-x-2 text-[11px] text-slate-500 font-mono">
                    <ShieldCheck className="w-3.5 h-3.5 text-indigo-400" />
                    <span>Cryptographic QR Verification Active</span>
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
                          Cancel Booking & Auto-Reallocate
                        </button>

                        <button
                          onClick={() => setSelectedBookingForQR(booking)}
                          className="px-5 py-2 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-500 shadow-md shadow-indigo-600/25 transition flex items-center"
                        >
                          <QrCode className="w-3.5 h-3.5 mr-1.5" />
                          View QR E-Pass
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
                REF: {selectedBookingForQR.booking_reference}
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

            <p className="text-[10px] text-slate-500 font-mono">
              Scan this barcode at venue admission turnstiles for instant entrance verification.
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
