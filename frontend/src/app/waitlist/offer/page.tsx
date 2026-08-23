"use client";

import React, { useState, useEffect, Suspense } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import {
  Sparkles,
  Ticket,
  Clock,
  MapPin,
  Calendar,
  CheckCircle2,
  AlertTriangle,
  RefreshCw,
  ArrowRight,
  ShieldCheck,
  Zap,
  Lock,
} from "lucide-react";
import { QRCodeSVG } from "qrcode.react";
import confetti from "canvas-confetti";
import { acceptWaitlistOffer, fetchEventById } from "../../../lib/api";
import { Booking, EventItem } from "../../../lib/types";
import { useAuth } from "../../../context/AuthContext";

function WaitlistOfferContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token") || "";
  const eventId = searchParams.get("event") || "";

  const { user, isLoading: authLoading } = useAuth();
  const [event, setEvent] = useState<EventItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [confirmedBooking, setConfirmedBooking] = useState<Booking | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    async function loadEvent() {
      if (!eventId) {
        setLoading(false);
        return;
      }
      try {
        const ev = await fetchEventById(eventId);
        setEvent(ev);
      } catch {
        // Event metadata load is non-fatal
      } finally {
        setLoading(false);
      }
    }
    loadEvent();
  }, [eventId]);

  const handleAcceptOffer = async () => {
    if (!user) {
      router.push(`/login?redirect=${encodeURIComponent(`/waitlist/offer?token=${token}&event=${eventId}`)}`);
      return;
    }
    if (!token) {
      setErrorMessage("Missing offer token in URL.");
      return;
    }

    setSubmitting(true);
    setErrorMessage(null);

    try {
      const booking = await acceptWaitlistOffer(token);
      setConfirmedBooking(booking);

      try {
        confetti({
          particleCount: 120,
          spread: 80,
          origin: { y: 0.6 },
        });
      } catch {
        // Ignore confetti errors
      }
    } catch (err: any) {
      setErrorMessage(err.message || "Failed to accept waitlist offer. The offer may have expired.");
    } finally {
      setSubmitting(false);
    }
  };

  if (authLoading || loading) {
    return (
      <div className="min-h-[70vh] flex flex-col items-center justify-center space-y-4 text-center">
        <RefreshCw className="w-10 h-10 animate-spin text-emerald-400" />
        <p className="text-sm font-semibold text-slate-300">Loading your exclusive waitlist offer...</p>
      </div>
    );
  }

  if (!token) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-24 text-center space-y-6">
        <div className="glass-panel rounded-3xl border border-rose-500/40 p-8 sm:p-12 space-y-4 bg-[#131A26]">
          <AlertTriangle className="w-12 h-12 mx-auto text-rose-400" />
          <h2 className="text-xl font-bold text-rose-200">Invalid Offer Link</h2>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            No valid offer token was detected in your link. Please check the link from your notification email.
          </p>
          <div className="pt-2">
            <Link
              href="/"
              className="inline-flex items-center px-5 py-2.5 rounded-xl font-bold text-xs text-white bg-indigo-600 hover:bg-indigo-500 transition"
            >
              Browse Live Shows
            </Link>
          </div>
        </div>
      </div>
    );
  }

  // State 1: Booking Confirmed via Offer
  if (confirmedBooking) {
    return (
      <div className="max-w-xl mx-auto px-4 py-16 text-center space-y-8 animate-in fade-in zoom-in-95">
        <div className="w-16 h-16 rounded-3xl bg-emerald-500/20 text-emerald-400 flex items-center justify-center mx-auto shadow-xl shadow-emerald-500/20 border border-emerald-500/30">
          <CheckCircle2 className="w-8 h-8" />
        </div>

        <div className="space-y-2">
          <span className="text-xs font-mono font-bold text-emerald-400 uppercase tracking-widest bg-emerald-500/10 px-3 py-1 rounded-full border border-emerald-500/20">
            WAITLIST OFFER ACCEPTED
          </span>
          <h1 className="text-2xl sm:text-3xl font-black text-white tracking-tight">
            You&apos;re Going to the Show!
          </h1>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            Your waitlist offer was successfully confirmed. Your e-ticket pass is ready below.
          </p>
        </div>

        {/* E-Ticket Display */}
        <div className="glass-panel rounded-3xl p-6 sm:p-8 border border-slate-800 bg-[#131A26] shadow-2xl space-y-6">
          <div className="flex justify-between items-center text-xs font-mono border-b border-slate-800 pb-3">
            <span className="text-slate-400">BOOKING REF:</span>
            <span className="text-emerald-400 font-black">{confirmedBooking.booking_reference}</span>
          </div>

          <div className="p-4 bg-white rounded-2xl inline-block shadow-lg mx-auto">
            <QRCodeSVG
              value={confirmedBooking.tickets?.[0]?.qr_code_payload || confirmedBooking.booking_reference}
              size={180}
              level="H"
            />
          </div>

          <div className="space-y-2 text-xs">
            <h3 className="text-base font-bold text-white">
              {event?.title || "Live Performance"}
            </h3>
            {event?.venue_name && (
              <p className="text-slate-400">{event.venue_name}, {event.venue_city}</p>
            )}

            <div className="pt-2 flex flex-wrap justify-center gap-2">
              {confirmedBooking.tickets?.map((t) => (
                <span
                  key={t.id}
                  className="px-3 py-1.5 rounded-xl bg-emerald-950/80 text-emerald-300 border border-emerald-700/60 text-xs font-mono font-bold"
                >
                  Row {t.row_label} • Seat {t.seat_number} ({t.category_name})
                </span>
              ))}
            </div>

            <div className="pt-4 border-t border-slate-800 flex justify-between text-xs font-mono">
              <span className="text-slate-400">Total Charged:</span>
              <span className="text-white font-bold">${confirmedBooking.total_amount.toFixed(2)} {confirmedBooking.currency}</span>
            </div>
          </div>
        </div>

        <div className="flex gap-4 justify-center">
          <Link
            href="/my-bookings"
            className="px-6 py-3 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 shadow-lg shadow-emerald-600/30 transition flex items-center"
          >
            <Ticket className="w-4 h-4 mr-2" />
            View in My Passes
          </Link>
          <Link
            href="/"
            className="px-6 py-3 rounded-xl font-semibold text-xs text-slate-300 hover:text-white bg-slate-800 border border-slate-700 transition"
          >
            Back to Home
          </Link>
        </div>
      </div>
    );
  }

  // State 2: Time-Limited Offer Claim Card
  return (
    <div className="max-w-2xl mx-auto px-4 py-16 space-y-8">
      {/* Header Banner */}
      <div className="text-center space-y-2">
        <div className="inline-flex items-center space-x-2 text-xs font-mono font-bold text-emerald-400 uppercase tracking-widest bg-emerald-500/10 px-3 py-1 rounded-full border border-emerald-500/20">
          <Sparkles className="w-3.5 h-3.5 animate-pulse" />
          <span>EXCLUSIVE SEAT ALLOCATION</span>
        </div>
        <h1 className="text-2xl sm:text-3xl font-black text-white tracking-tight">
          A Seat Opened Up For You!
        </h1>
        <p className="text-xs text-slate-400 max-w-md mx-auto">
          A cancellation occurred and your position on the FIFO waitlist has matured. Complete your booking before the offer window closes.
        </p>
      </div>

      {/* Offer Countdown & Notice */}
      <div className="glass-panel rounded-2xl p-4 border border-amber-500/40 bg-gradient-to-r from-amber-500/10 via-[#131A26] to-amber-500/10 shadow-lg flex items-center justify-between gap-4">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 rounded-xl bg-amber-500/20 text-amber-400 flex items-center justify-center shrink-0">
            <Clock className="w-5 h-5 animate-spin" style={{ animationDuration: "12s" }} />
          </div>
          <div>
            <h4 className="text-xs font-bold text-amber-300">
              Time-Limited Seat Hold Active
            </h4>
            <p className="text-[11px] text-slate-400">
              If not confirmed within the offer window, this seat will automatically pass to the next customer in line.
            </p>
          </div>
        </div>
      </div>

      {/* Error Alert */}
      {errorMessage && (
        <div className="p-4 rounded-2xl bg-rose-500/10 border border-rose-500/40 text-rose-300 text-xs flex items-center justify-between animate-in fade-in">
          <div className="flex items-center space-x-2.5">
            <AlertTriangle className="w-4 h-4 text-rose-400 shrink-0" />
            <span className="font-semibold">{errorMessage}</span>
          </div>
          <button onClick={() => setErrorMessage(null)} className="text-rose-400 hover:text-rose-200 font-bold">
            ✕
          </button>
        </div>
      )}

      {/* Offer Details Card */}
      <div className="glass-panel rounded-3xl p-6 sm:p-8 border border-slate-800 bg-[#131A26] shadow-2xl space-y-6">
        <div className="flex items-start space-x-5 border-b border-slate-800 pb-6">
          {event?.poster_url && (
            <div className="w-20 h-28 rounded-xl bg-slate-900 overflow-hidden shrink-0 border border-slate-700 shadow-md">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src={event.poster_url} alt={event.title} className="w-full h-full object-cover" />
            </div>
          )}
          <div className="space-y-2">
            <span className="text-[10px] font-mono font-bold uppercase tracking-wider px-2.5 py-0.5 rounded bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
              {event?.event_type || "EVENT"}
            </span>
            <h2 className="text-xl font-bold text-white tracking-tight">
              {event?.title || "Reserved Performance"}
            </h2>
            <div className="flex flex-wrap items-center gap-3 text-xs text-slate-400">
              {event?.start_time && (
                <span className="flex items-center">
                  <Calendar className="w-3.5 h-3.5 mr-1 text-indigo-400" />
                  {new Date(event.start_time).toLocaleDateString("en-US", {
                    weekday: "short",
                    month: "short",
                    day: "numeric",
                  })}
                </span>
              )}
              {event?.venue_name && (
                <span className="flex items-center">
                  <MapPin className="w-3.5 h-3.5 mr-1 text-rose-400" />
                  {event.venue_name}, {event.venue_city}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Action Button */}
        <div className="space-y-3 pt-2">
          <button
            onClick={handleAcceptOffer}
            disabled={submitting}
            className="w-full py-4 rounded-2xl font-bold text-sm text-white bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 shadow-xl shadow-emerald-600/30 transition flex items-center justify-center disabled:opacity-40"
          >
            {submitting ? (
              <RefreshCw className="w-5 h-5 animate-spin" />
            ) : (
              <>
                <CheckCircle2 className="w-5 h-5 mr-2" />
                Accept Offer & Complete Booking
              </>
            )}
          </button>

          <div className="flex items-center justify-center space-x-2 text-[11px] text-slate-400 font-mono">
            <ShieldCheck className="w-3.5 h-3.5 text-emerald-400" />
            <span>Instant Digital Pass Generation & Confirmation Email</span>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function WaitlistOfferPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-[70vh] flex flex-col items-center justify-center space-y-4">
          <RefreshCw className="w-8 h-8 animate-spin text-indigo-400" />
          <p className="text-xs font-semibold text-slate-300">Loading offer details...</p>
        </div>
      }
    >
      <WaitlistOfferContent />
    </Suspense>
  );
}
