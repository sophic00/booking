"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import {
  Building2,
  DollarSign,
  Users,
  Ticket,
  BarChart3,
  Percent,
  Plus,
  ArrowRight,
  RefreshCw,
  Clock,
  Sparkles,
  Calendar,
  MapPin,
  TrendingUp,
  AlertCircle,
  Lock,
  QrCode,
  ScanLine,
  CheckCircle2,
  XCircle,
  ShieldCheck,
  UserCheck,
  Search,
} from "lucide-react";
import {
  fetchOrganiserAnalytics,
  fetchEvents,
  fetchOrganiserEvents,
  createOrganiserEvent,
  verifyTicket,
  checkInTicket,
  fetchEventTickets,
} from "../../lib/api";
import {
  EventAnalytics,
  EventItem,
  TicketVerificationResult,
  EventTicketItem,
  EventCheckInOverview,
} from "../../lib/types";
import { useAuth } from "../../context/AuthContext";

export default function OrganiserPage() {
  const { user, isLoading: authLoading } = useAuth();
  const [analytics, setAnalytics] = useState<EventAnalytics | null>(null);
  const [events, setEvents] = useState<EventItem[]>([]);
  const [selectedEventId, setSelectedEventId] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Ticket Gate Scanner & Check-In State
  const [scannerInput, setScannerInput] = useState<string>("");
  const [verifying, setVerifying] = useState<boolean>(false);
  const [checkingIn, setCheckingIn] = useState<boolean>(false);
  const [verificationResult, setVerificationResult] = useState<TicketVerificationResult | null>(null);
  const [verifyError, setVerifyError] = useState<string | null>(null);
  const [actionSuccessMessage, setActionSuccessMessage] = useState<string | null>(null);
  const [ticketOverview, setTicketOverview] = useState<EventCheckInOverview | null>(null);
  const [loadingTickets, setLoadingTickets] = useState<boolean>(false);
  const [ticketStatusFilter, setTicketStatusFilter] = useState<string>("");

  // New Event Modal
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newEventTitle, setNewEventTitle] = useState("");
  const [newEventType, setNewEventType] = useState("CONCERT");
  const [newEventDate, setNewEventDate] = useState("2026-11-20T20:00");
  const [newEventVenue, setNewEventVenue] = useState("");
  const [holdTTL, setHoldTTL] = useState(600);
  const [isCreating, setIsCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);

  const loadTickets = async (eventId: string, filter?: string) => {
    if (!eventId) return;
    setLoadingTickets(true);
    try {
      const data = await fetchEventTickets(eventId, filter || undefined);
      setTicketOverview(data);
    } catch {
      setTicketOverview(null);
    } finally {
      setLoadingTickets(false);
    }
  };

  const loadEvents = async () => {
    if (!user) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      // Try fetching organiser-specific events first, fallback to published events
      let evs: EventItem[] = [];
      try {
        evs = await fetchOrganiserEvents();
      } catch {
        evs = await fetchEvents();
      }
      setEvents(evs);
      if (evs.length > 0) {
        const targetId = selectedEventId && evs.some((e) => e.id === selectedEventId)
          ? selectedEventId
          : evs[0].id;
        setSelectedEventId(targetId);
        await Promise.all([loadAnalytics(targetId), loadTickets(targetId)]);
      } else {
        setAnalytics(null);
        setTicketOverview(null);
      }
    } catch (err: any) {
      setError(err.message || "Failed to load events from backend API");
      setEvents([]);
      setAnalytics(null);
    } finally {
      setLoading(false);
    }
  };

  const loadAnalytics = async (eventId: string) => {
    if (!eventId) return;
    try {
      const data = await fetchOrganiserAnalytics(eventId);
      setAnalytics(data);
    } catch (err: any) {
      setError(err.message || `Failed to load analytics for event #${eventId}`);
      setAnalytics(null);
    }
  };

  useEffect(() => {
    if (!authLoading) {
      loadEvents();
    }
  }, [user, authLoading]);

  const handleEventSelect = (id: string) => {
    setSelectedEventId(id);
    loadAnalytics(id);
    loadTickets(id, ticketStatusFilter);
    setVerificationResult(null);
    setVerifyError(null);
    setActionSuccessMessage(null);
  };

  const handleVerify = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!scannerInput.trim()) return;

    setVerifying(true);
    setVerifyError(null);
    setActionSuccessMessage(null);

    try {
      const val = scannerInput.trim();
      const payload = val.startsWith("TICKET|") ? { qr_payload: val } : { ticket_id: val };
      const res = await verifyTicket(payload);
      setVerificationResult(res);
    } catch (err: any) {
      setVerifyError(err.message || "Ticket verification failed");
      setVerificationResult(null);
    } finally {
      setVerifying(false);
    }
  };

  const handleCheckIn = async (payloadOrId?: string) => {
    const val = (payloadOrId || scannerInput).trim();
    if (!val) return;

    setCheckingIn(true);
    setVerifyError(null);
    setActionSuccessMessage(null);

    try {
      const payload = val.startsWith("TICKET|") ? { qr_payload: val } : { ticket_id: val };
      const res = await checkInTicket(payload);
      setVerificationResult(res);
      setActionSuccessMessage(`✓ ${res.booking.customer_name} successfully checked in for seat ${res.seat.row_label}-${res.seat.seat_number}!`);
      if (selectedEventId) {
        await Promise.all([
          loadTickets(selectedEventId, ticketStatusFilter),
          loadAnalytics(selectedEventId),
        ]);
      }
    } catch (err: any) {
      setVerifyError(err.message || "Check-in failed");
    } finally {
      setCheckingIn(false);
    }
  };

  const handleCreateEvent = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateError(null);
    setIsCreating(true);

    try {
      await createOrganiserEvent({
        venue_id: newEventVenue || "00000000-0000-0000-0000-000000000001",
        title: newEventTitle,
        event_type: newEventType,
        start_time: new Date(newEventDate).toISOString(),
        end_time: new Date(new Date(newEventDate).getTime() + 3 * 3600 * 1000).toISOString(),
        hold_ttl_seconds: holdTTL,
      });
      alert(`🎉 New event listing "${newEventTitle}" created successfully in Draft mode.`);
      setShowCreateModal(false);
      setNewEventTitle("");
      await loadEvents();
    } catch (err: any) {
      setCreateError(err.message || "Failed to create event on backend API");
    } finally {
      setIsCreating(false);
    }
  };

  if (authLoading) {
    return (
      <div className="py-24 text-center text-slate-400">
        <RefreshCw className="w-8 h-8 mx-auto animate-spin text-violet-400 mb-3" />
        <p className="font-semibold text-slate-300">Checking organiser authorization...</p>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="max-w-2xl mx-auto px-4 py-24 text-center space-y-6">
        <div className="glass-panel rounded-3xl border border-slate-800 p-8 sm:p-12 space-y-4 bg-[#131A26]">
          <div className="w-12 h-12 rounded-2xl bg-violet-500/20 text-violet-400 flex items-center justify-center mx-auto">
            <Lock className="w-6 h-6" />
          </div>
          <h2 className="text-xl font-bold text-white">Organiser Portal Access Required</h2>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            Please sign in with an Organiser or Admin account to access real-time event analytics, manage seat pricing, and view live occupancy queues.
          </p>
          <div className="pt-2">
            <Link
              href="/login"
              className="inline-flex items-center px-5 py-2.5 rounded-xl font-bold text-xs text-white bg-violet-600 hover:bg-violet-500 transition shadow-lg shadow-violet-600/30"
            >
              Sign In as Organiser
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
          <div className="flex items-center space-x-2 text-xs font-mono font-bold text-violet-400 uppercase tracking-wider">
            <Building2 className="w-3.5 h-3.5" />
            <span>ORGANISER & VENUE PORTAL</span>
          </div>
          <h1 className="text-2xl sm:text-3xl font-black text-white tracking-tight mt-1">
            Event Management & Real-Time Analytics
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Monitor seat holds, live revenue by category, and automated waitlist reallocation queues.
          </p>
        </div>

        <button
          onClick={() => setShowCreateModal(true)}
          className="px-5 py-2.5 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 shadow-lg shadow-violet-600/25 transition flex items-center self-start sm:self-auto"
        >
          <Plus className="w-4 h-4 mr-1.5" />
          Create New Event Listing
        </button>
      </div>

      {/* Select Event Filter */}
      <div className="flex flex-wrap items-center justify-between gap-4 bg-[#131A26] p-4 rounded-2xl border border-slate-800">
        <div className="flex items-center space-x-3">
          <span className="text-xs font-semibold text-slate-400 uppercase font-mono">
            Active Event:
          </span>
          <select
            value={selectedEventId}
            onChange={(e) => setSelectedEventId(e.target.value)}
            className="bg-[#0B0F17] border border-slate-700/80 rounded-xl px-3 py-1.5 text-xs font-bold text-slate-100 focus:outline-none focus:border-violet-500 transition cursor-pointer"
          >
            {events.map((e) => (
              <option key={e.id} value={e.id}>
                {e.title} ({e.venue_city})
              </option>
            ))}
          </select>
        </div>

        <div className="flex items-center space-x-3 text-xs text-slate-400 font-mono">
          <span className="flex items-center text-emerald-400 font-bold">
            <span className="w-2 h-2 rounded-full bg-emerald-400 mr-1.5 animate-pulse" />
            Live Sales & Ticketing Active
          </span>
        </div>
      </div>

      {/* Key Metric Stats Cards / Error Banner */}
      {loading ? (
        <div className="py-20 text-center text-slate-400">
          <RefreshCw className="w-8 h-8 mx-auto animate-spin text-violet-400 mb-3" />
          <p className="font-semibold text-slate-300">Calculating revenue & occupancy metrics from backend...</p>
        </div>
      ) : error ? (
        <div className="py-16 text-center glass-panel rounded-3xl border border-rose-500/40 p-8 space-y-4 bg-gradient-to-b from-rose-950/20 to-[#131A26]">
          <AlertCircle className="w-12 h-12 mx-auto text-rose-400" />
          <h3 className="text-lg font-bold text-rose-200">Failed to Load Organiser Analytics</h3>
          <p className="text-xs text-slate-300 max-w-md mx-auto font-mono bg-black/40 p-3 rounded-xl border border-rose-500/20 text-rose-300">
            {error}
          </p>
          <button
            onClick={loadEvents}
            className="mt-2 px-5 py-2.5 rounded-xl text-xs font-bold text-white bg-rose-600 hover:bg-rose-500 transition shadow-lg shadow-rose-600/30"
          >
            Retry Connection
          </button>
        </div>
      ) : !analytics ? (
        <div className="py-16 text-center glass-panel rounded-3xl border border-slate-800 p-8 space-y-3">
          <Ticket className="w-12 h-12 mx-auto text-slate-600" />
          <h3 className="text-lg font-bold text-slate-200">No event analytics available</h3>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            Create an event listing to start monitoring real-time seat holds and tier revenues.
          </p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {/* Total Revenue */}
            <div className="glass-panel rounded-2xl p-6 border border-slate-800 bg-gradient-to-b from-indigo-950/20 to-transparent space-y-3">
              <div className="flex items-center justify-between text-slate-400">
                <span className="text-xs font-semibold uppercase font-mono">Total Revenue</span>
                <DollarSign className="w-4 h-4 text-emerald-400" />
              </div>
              <div className="text-2xl font-black text-white font-mono">
                ${analytics.total_revenue.toLocaleString()}{" "}
                <span className="text-xs text-slate-400 font-normal">USD</span>
              </div>
              <div className="text-[11px] text-emerald-400 flex items-center">
                <TrendingUp className="w-3.5 h-3.5 mr-1" />
                <span>+18.4% vs initial projections</span>
              </div>
            </div>

            {/* Occupancy Rate */}
            <div className="glass-panel rounded-2xl p-6 border border-slate-800 bg-gradient-to-b from-violet-950/20 to-transparent space-y-3">
              <div className="flex items-center justify-between text-slate-400">
                <span className="text-xs font-semibold uppercase font-mono">Occupancy Rate</span>
                <Percent className="w-4 h-4 text-violet-400" />
              </div>
              <div className="text-2xl font-black text-white font-mono">
                {analytics.occupancy_percentage.toFixed(1)}%
              </div>
              <div className="w-full bg-slate-800 h-1.5 rounded-full overflow-hidden">
                <div
                  className="bg-violet-500 h-full rounded-full"
                  style={{ width: `${analytics.occupancy_percentage}%` }}
                />
              </div>
            </div>

            {/* Confirmed Tickets */}
            <div className="glass-panel rounded-2xl p-6 border border-slate-800 bg-gradient-to-b from-purple-950/20 to-transparent space-y-3">
              <div className="flex items-center justify-between text-slate-400">
                <span className="text-xs font-semibold uppercase font-mono">Valid Tickets</span>
                <Ticket className="w-4 h-4 text-purple-400" />
              </div>
              <div className="text-2xl font-black text-white font-mono">
                {analytics.valid_tickets_count.toLocaleString()}{" "}
                <span className="text-xs text-slate-400 font-normal">
                  / {analytics.total_capacity}
                </span>
              </div>
              <div className="text-[11px] text-slate-400 font-mono">
                {analytics.confirmed_bookings_count} confirmed orders
              </div>
            </div>

            {/* Waitlist Queue */}
            <div className="glass-panel rounded-2xl p-6 border border-slate-800 bg-gradient-to-b from-rose-950/20 to-transparent space-y-3">
              <div className="flex items-center justify-between text-slate-400">
                <span className="text-xs font-semibold uppercase font-mono">Waitlist Queue</span>
                <Users className="w-4 h-4 text-rose-400" />
              </div>
              <div className="text-2xl font-black text-white font-mono">
                {analytics.waitlist_waiting_count}{" "}
                <span className="text-xs text-slate-400 font-normal">waiting</span>
              </div>
              <div className="text-[11px] text-rose-400 font-mono">
                Auto-assigned on cancellations
              </div>
            </div>
          </div>

          {/* Category Breakdown Table */}
          <div className="glass-panel rounded-3xl p-6 sm:p-8 border border-slate-800 bg-[#131A26] space-y-6">
            <div className="flex items-center justify-between border-b border-slate-800 pb-4">
              <div>
                <h3 className="text-lg font-bold text-white flex items-center">
                  <BarChart3 className="w-5 h-5 mr-2 text-violet-400" />
                  Per-Category Pricing & Revenue Breakdown
                </h3>
                <p className="text-xs text-slate-400 mt-0.5">
                  Detailed seat fill rates, gross revenue, and waitlist demand per seating tier.
                </p>
              </div>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-slate-800 text-slate-400 font-mono uppercase text-[11px]">
                    <th className="pb-3 font-semibold">Seat Category</th>
                    <th className="pb-3 font-semibold">Total Seats</th>
                    <th className="pb-3 font-semibold">Booked Seats</th>
                    <th className="pb-3 font-semibold">Fill Rate</th>
                    <th className="pb-3 font-semibold">Gross Revenue</th>
                    <th className="pb-3 font-semibold">Waitlist Queue</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60 font-mono">
                  {analytics.category_breakdown.map((cat) => {
                    const fillPercent = ((cat.booked_seats / cat.total_seats) * 100).toFixed(1);
                    return (
                      <tr key={cat.seat_category_id} className="hover:bg-slate-800/40 transition">
                        <td className="py-3.5 flex items-center space-x-2 font-bold text-slate-100">
                          <span
                            className="w-3 h-3 rounded-full"
                            style={{ backgroundColor: cat.color_code }}
                          />
                          <span>{cat.category_name}</span>
                        </td>
                        <td className="py-3.5 text-slate-300">{cat.total_seats}</td>
                        <td className="py-3.5 text-slate-300">{cat.booked_seats}</td>
                        <td className="py-3.5">
                          <div className="flex items-center space-x-2">
                            <span className="text-indigo-400 font-bold">{fillPercent}%</span>
                            <div className="w-16 bg-slate-800 h-1.5 rounded-full overflow-hidden hidden sm:block">
                              <div
                                className="bg-indigo-500 h-full rounded-full"
                                style={{ width: `${fillPercent}%` }}
                              />
                            </div>
                          </div>
                        </td>
                        <td className="py-3.5 font-bold text-emerald-400">
                          ${cat.revenue.toLocaleString()}
                        </td>
                        <td className="py-3.5 text-rose-400 font-bold">{cat.waitlist_count} fans</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>

          {/* Ticket Verification & Gate Check-In Scanner Section */}
          <div className="glass-panel rounded-3xl p-6 sm:p-8 border border-slate-800 bg-[#131A26] space-y-6">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-4">
              <div>
                <div className="flex items-center space-x-2 text-xs font-mono font-bold text-violet-400 uppercase tracking-wider">
                  <ScanLine className="w-4 h-4 text-violet-400" />
                  <span>GATE CONTROL & ADMISSION</span>
                </div>
                <h3 className="text-lg font-bold text-white flex items-center mt-0.5">
                  Live Ticket Scanner & Attendee Check-In
                </h3>
                <p className="text-xs text-slate-400 mt-0.5">
                  Verify QR code payloads, validate attendee seat assignments, and record check-ins with anti-duplicate protection.
                </p>
              </div>

              {ticketOverview && (
                <div className="flex items-center space-x-3 bg-black/40 px-4 py-2 rounded-2xl border border-slate-800 self-start sm:self-auto font-mono text-xs">
                  <div className="flex items-center space-x-1.5">
                    <UserCheck className="w-4 h-4 text-emerald-400" />
                    <span className="text-slate-400">Gate Checked-In:</span>
                    <span className="font-black text-white">
                      {ticketOverview.checked_in_count} / {ticketOverview.valid_count + ticketOverview.checked_in_count}
                    </span>
                  </div>
                  <div className="h-4 w-px bg-slate-700" />
                  <span className="font-bold text-emerald-400">{ticketOverview.check_in_rate.toFixed(1)}%</span>
                </div>
              )}
            </div>

            {/* Notification Messages */}
            {actionSuccessMessage && (
              <div className="p-3.5 rounded-2xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 text-xs flex items-center justify-between animate-in fade-in">
                <div className="flex items-center space-x-2">
                  <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
                  <span>{actionSuccessMessage}</span>
                </div>
                <button
                  onClick={() => setActionSuccessMessage(null)}
                  className="text-emerald-400/60 hover:text-emerald-300 text-xs font-bold"
                >
                  ✕
                </button>
              </div>
            )}

            {verifyError && (
              <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs flex items-center justify-between animate-in fade-in">
                <div className="flex items-center space-x-2">
                  <XCircle className="w-4 h-4 text-rose-400 shrink-0" />
                  <span>{verifyError}</span>
                </div>
                <button
                  onClick={() => setVerifyError(null)}
                  className="text-rose-400/60 hover:text-rose-300 text-xs font-bold"
                >
                  ✕
                </button>
              </div>
            )}

            {/* Scanner Input Form */}
            <form onSubmit={handleVerify} className="space-y-3">
              <div className="flex flex-col sm:flex-row gap-3">
                <div className="relative flex-1">
                  <QrCode className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                  <input
                    type="text"
                    value={scannerInput}
                    onChange={(e) => setScannerInput(e.target.value)}
                    placeholder="Scan QR code payload (e.g. TICKET|REF:TB-...|SEAT:...|ID:...) or Ticket UUID"
                    className="w-full pl-10 pr-4 py-2.5 bg-[#0B0F17] border border-slate-700/80 rounded-xl text-xs text-slate-100 placeholder-slate-500 focus:outline-none focus:border-violet-500 font-mono transition"
                  />
                </div>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() => handleVerify()}
                    disabled={verifying || !scannerInput.trim()}
                    className="px-4 py-2.5 rounded-xl font-bold text-xs text-white bg-slate-800 hover:bg-slate-700 border border-slate-600 transition flex items-center disabled:opacity-50"
                  >
                    {verifying ? <RefreshCw className="w-3.5 h-3.5 animate-spin mr-1.5" /> : <Search className="w-3.5 h-3.5 mr-1.5" />}
                    Verify Ticket
                  </button>
                  <button
                    type="button"
                    onClick={() => handleCheckIn()}
                    disabled={checkingIn || !scannerInput.trim()}
                    className="px-5 py-2.5 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 shadow-lg shadow-emerald-600/20 transition flex items-center disabled:opacity-50"
                  >
                    {checkingIn ? <RefreshCw className="w-3.5 h-3.5 animate-spin mr-1.5" /> : <UserCheck className="w-3.5 h-3.5 mr-1.5" />}
                    Direct Check-In
                  </button>
                </div>
              </div>
            </form>

            {/* Verification Result Card */}
            {verificationResult && (
              <div
                className={`p-5 rounded-2xl border ${
                  verificationResult.ticket.status === "VALID"
                    ? "bg-emerald-950/15 border-emerald-500/40"
                    : verificationResult.ticket.status === "CHECKED_IN"
                    ? "bg-amber-950/15 border-amber-500/40"
                    : "bg-rose-950/15 border-rose-500/40"
                } space-y-4 animate-in fade-in zoom-in-95`}
              >
                <div className="flex flex-wrap items-center justify-between gap-2 border-b border-slate-800/80 pb-3">
                  <div className="flex items-center space-x-2">
                    <span
                      className={`px-2.5 py-1 rounded-full text-[11px] font-black font-mono uppercase tracking-wider flex items-center ${
                        verificationResult.ticket.status === "VALID"
                          ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/40"
                          : verificationResult.ticket.status === "CHECKED_IN"
                          ? "bg-amber-500/20 text-amber-400 border border-amber-500/40"
                          : "bg-rose-500/20 text-rose-400 border border-rose-500/40"
                      }`}
                    >
                      {verificationResult.ticket.status === "VALID" && <CheckCircle2 className="w-3 h-3 mr-1" />}
                      {verificationResult.ticket.status === "CHECKED_IN" && <ShieldCheck className="w-3 h-3 mr-1" />}
                      {verificationResult.ticket.status === "CANCELLED" && <XCircle className="w-3 h-3 mr-1" />}
                      {verificationResult.ticket.status}
                    </span>
                    <span className="text-xs text-slate-300 font-medium">
                      {verificationResult.validation_message}
                    </span>
                  </div>

                  {verificationResult.can_check_in && (
                    <button
                      onClick={() => handleCheckIn(verificationResult.ticket.id)}
                      disabled={checkingIn}
                      className="px-4 py-1.5 rounded-xl text-xs font-bold text-white bg-emerald-600 hover:bg-emerald-500 shadow-md shadow-emerald-600/30 transition flex items-center"
                    >
                      {checkingIn ? <RefreshCw className="w-3 h-3 animate-spin mr-1.5" /> : <UserCheck className="w-3 h-3 mr-1.5" />}
                      Confirm Gate Check-In
                    </button>
                  )}
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4 text-xs font-mono">
                  <div className="bg-black/30 p-3 rounded-xl border border-slate-800">
                    <span className="text-slate-400 text-[10px] uppercase block">Attendee</span>
                    <span className="font-bold text-white block mt-0.5">{verificationResult.booking.customer_name}</span>
                    <span className="text-slate-400 text-[11px] block">{verificationResult.booking.customer_email}</span>
                  </div>

                  <div className="bg-black/30 p-3 rounded-xl border border-slate-800">
                    <span className="text-slate-400 text-[10px] uppercase block">Assigned Seat</span>
                    <div className="flex items-center space-x-1.5 mt-0.5">
                      <span className="font-bold text-white">
                        Row {verificationResult.seat.row_label} • Seat {verificationResult.seat.seat_number}
                      </span>
                    </div>
                    <span
                      className="text-[11px] font-bold inline-block mt-0.5"
                      style={{ color: verificationResult.seat.category_color }}
                    >
                      {verificationResult.seat.category_name} (${verificationResult.ticket.unit_price})
                    </span>
                  </div>

                  <div className="bg-black/30 p-3 rounded-xl border border-slate-800">
                    <span className="text-slate-400 text-[10px] uppercase block">Order Details</span>
                    <span className="font-bold text-slate-200 block mt-0.5">{verificationResult.booking.booking_reference}</span>
                    <span className="text-slate-400 text-[11px] block">Status: {verificationResult.booking.status}</span>
                  </div>

                  <div className="bg-black/30 p-3 rounded-xl border border-slate-800">
                    <span className="text-slate-400 text-[10px] uppercase block">Check-In Status</span>
                    {verificationResult.ticket.checked_in_at ? (
                      <span className="font-bold text-amber-400 block mt-0.5">
                        Checked in at {new Date(verificationResult.ticket.checked_in_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                      </span>
                    ) : (
                      <span className="font-bold text-emerald-400 block mt-0.5">Ready for Entry</span>
                    )}
                    <span className="text-slate-500 text-[10px] block truncate">{verificationResult.ticket.id}</span>
                  </div>
                </div>
              </div>
            )}

            {/* Event Attendee Tickets Roster Table */}
            <div className="space-y-3 pt-2">
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-800 pb-3">
                <h4 className="text-xs font-bold text-slate-200 uppercase font-mono flex items-center">
                  <Ticket className="w-3.5 h-3.5 mr-1.5 text-violet-400" />
                  Event Tickets & Gate Roster ({ticketOverview?.tickets?.length || 0})
                </h4>

                <div className="flex items-center space-x-2">
                  <span className="text-[11px] text-slate-400 font-mono">Filter:</span>
                  <select
                    value={ticketStatusFilter}
                    onChange={(e) => {
                      setTicketStatusFilter(e.target.value);
                      if (selectedEventId) loadTickets(selectedEventId, e.target.value);
                    }}
                    className="bg-[#0B0F17] border border-slate-700 rounded-lg px-2.5 py-1 text-xs text-slate-200 focus:outline-none focus:border-violet-500"
                  >
                    <option value="">All Tickets</option>
                    <option value="VALID">Valid (Not Checked-In)</option>
                    <option value="CHECKED_IN">Checked-In</option>
                    <option value="CANCELLED">Cancelled</option>
                  </select>
                </div>
              </div>

              {loadingTickets ? (
                <div className="py-8 text-center text-slate-400 text-xs flex items-center justify-center space-x-2">
                  <RefreshCw className="w-3.5 h-3.5 animate-spin text-violet-400" />
                  <span>Loading attendee tickets...</span>
                </div>
              ) : !ticketOverview?.tickets || ticketOverview.tickets.length === 0 ? (
                <div className="py-8 text-center text-slate-500 text-xs">
                  No tickets found for this event matching the filter.
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-xs font-mono">
                    <thead>
                      <tr className="border-b border-slate-800 text-slate-400 uppercase text-[11px]">
                        <th className="pb-2.5 font-semibold">Attendee</th>
                        <th className="pb-2.5 font-semibold">Seat</th>
                        <th className="pb-2.5 font-semibold">Category</th>
                        <th className="pb-2.5 font-semibold">Status</th>
                        <th className="pb-2.5 font-semibold">Checked In</th>
                        <th className="pb-2.5 font-semibold text-right">Quick Action</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/60">
                      {ticketOverview.tickets.map((t) => (
                        <tr key={t.id} className="hover:bg-slate-800/30 transition">
                          <td className="py-2.5 text-slate-200 font-medium">
                            <div>{t.customer_name}</div>
                            <div className="text-[10px] text-slate-400">{t.booking_reference}</div>
                          </td>
                          <td className="py-2.5 text-slate-200">
                            Row {t.row_label} • Seat {t.seat_number}
                          </td>
                          <td className="py-2.5">
                            <span
                              className="px-2 py-0.5 rounded text-[10px] font-bold"
                              style={{
                                backgroundColor: `${t.category_color}20`,
                                color: t.category_color,
                                border: `1px solid ${t.category_color}40`,
                              }}
                            >
                              {t.category_name}
                            </span>
                          </td>
                          <td className="py-2.5">
                            <span
                              className={`px-2 py-0.5 rounded-full text-[10px] font-bold uppercase ${
                                t.status === "VALID"
                                  ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/30"
                                  : t.status === "CHECKED_IN"
                                  ? "bg-amber-500/10 text-amber-400 border border-amber-500/30"
                                  : "bg-rose-500/10 text-rose-400 border border-rose-500/30"
                              }`}
                            >
                              {t.status}
                            </span>
                          </td>
                          <td className="py-2.5 text-slate-400 text-[11px]">
                            {t.checked_in_at
                              ? new Date(t.checked_in_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
                              : "—"}
                          </td>
                          <td className="py-2.5 text-right">
                            {t.status === "VALID" ? (
                              <button
                                onClick={() => {
                                  setScannerInput(t.qr_code_payload || t.id);
                                  handleCheckIn(t.id);
                                }}
                                disabled={checkingIn}
                                className="px-3 py-1 rounded-lg text-[11px] font-bold text-white bg-emerald-600 hover:bg-emerald-500 transition shadow-sm"
                              >
                                Check In
                              </button>
                            ) : (
                              <button
                                onClick={() => {
                                  setScannerInput(t.qr_code_payload || t.id);
                                  handleVerify();
                                }}
                                className="px-2.5 py-1 rounded-lg text-[11px] font-semibold text-slate-400 hover:text-white bg-slate-800 hover:bg-slate-700 transition"
                              >
                                Inspect
                              </button>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        </>
      )}

      {/* Create Event Modal Dialog */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/85 backdrop-blur-md animate-in fade-in">
          <div className="w-full max-w-lg bg-[#131A26] border border-violet-500/40 rounded-3xl shadow-2xl overflow-hidden p-6 sm:p-8 space-y-5 animate-in zoom-in-95">
            <div className="flex justify-between items-center border-b border-slate-800 pb-3">
              <h3 className="text-base font-bold text-white flex items-center">
                <Plus className="w-4 h-4 mr-2 text-violet-400" />
                Create New Show Listing
              </h3>
              <button
                onClick={() => setShowCreateModal(false)}
                className="text-slate-400 hover:text-white"
              >
                ✕
              </button>
            </div>

            {createError && (
              <div className="p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs flex items-center space-x-2">
                <AlertCircle className="w-4 h-4 shrink-0 text-rose-400" />
                <span>{createError}</span>
              </div>
            )}

            <form onSubmit={handleCreateEvent} className="space-y-4 text-xs">
              <div>
                <label className="text-slate-400 font-semibold block mb-1">Event Title</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Hans Zimmer Live Symphony 2026"
                  value={newEventTitle}
                  onChange={(e) => setNewEventTitle(e.target.value)}
                  className="w-full bg-[#0B0F17] border border-slate-700 rounded-xl px-3 py-2 text-slate-100 focus:outline-none focus:border-violet-500 transition"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-slate-400 font-semibold block mb-1">Category</label>
                  <select
                    value={newEventType}
                    onChange={(e) => setNewEventType(e.target.value)}
                    className="w-full bg-[#0B0F17] border border-slate-700 rounded-xl px-3 py-2 text-slate-100 focus:outline-none focus:border-violet-500 transition cursor-pointer"
                  >
                    <option value="CONCERT">Concert & Tour</option>
                    <option value="MOVIE">Movie / Cinema</option>
                    <option value="THEATRE">Theater & Musical</option>
                    <option value="SPORTS">Sports Event</option>
                  </select>
                </div>

                <div>
                  <label className="text-slate-400 font-semibold block mb-1">Hold TTL (Seconds)</label>
                  <input
                    type="number"
                    value={holdTTL}
                    onChange={(e) => setHoldTTL(Number(e.target.value))}
                    className="w-full bg-[#0B0F17] border border-slate-700 rounded-xl px-3 py-2 text-slate-100 focus:outline-none focus:border-violet-500 transition"
                  />
                </div>
              </div>

              <div>
                <label className="text-slate-400 font-semibold block mb-1">Date & Showtime</label>
                <input
                  type="datetime-local"
                  required
                  value={newEventDate}
                  onChange={(e) => setNewEventDate(e.target.value)}
                  className="w-full bg-[#0B0F17] border border-slate-700 rounded-xl px-3 py-2 text-slate-100 focus:outline-none focus:border-violet-500 transition"
                />
              </div>

              <div>
                <label className="text-slate-400 font-semibold block mb-1">Venue & Location</label>
                <input
                  type="text"
                  required
                  value={newEventVenue}
                  onChange={(e) => setNewEventVenue(e.target.value)}
                  placeholder="e.g. Royal Albert Hall, London"
                  className="w-full bg-[#0B0F17] border border-slate-700 rounded-xl px-3 py-2 text-slate-100 focus:outline-none focus:border-violet-500 transition"
                />
              </div>

              <div className="flex gap-3 pt-2">
                <button
                  type="submit"
                  disabled={isCreating}
                  className="flex-1 py-3 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 shadow-lg shadow-violet-600/30 transition flex items-center justify-center disabled:opacity-50"
                >
                  {isCreating ? (
                    <RefreshCw className="w-4 h-4 animate-spin" />
                  ) : (
                    "Create Listing (Draft)"
                  )}
                </button>
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-3 rounded-xl font-semibold text-xs text-slate-300 hover:text-white bg-slate-900 border border-slate-700 transition"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
