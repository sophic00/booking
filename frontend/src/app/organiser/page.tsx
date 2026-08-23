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
} from "lucide-react";
import { fetchOrganiserAnalytics, fetchEvents } from "../../lib/api";
import { EventAnalytics, EventItem } from "../../lib/types";

export default function OrganiserPage() {
  const [analytics, setAnalytics] = useState<EventAnalytics | null>(null);
  const [events, setEvents] = useState<EventItem[]>([]);
  const [selectedEventId, setSelectedEventId] = useState("evt-004");
  const [loading, setLoading] = useState(true);

  // New Event Modal
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newEventTitle, setNewEventTitle] = useState("");
  const [newEventType, setNewEventType] = useState("CONCERT");
  const [newEventDate, setNewEventDate] = useState("2026-11-20T20:00");
  const [newEventVenue, setNewEventVenue] = useState("Chase Center, San Francisco");
  const [holdTTL, setHoldTTL] = useState(600);

  useEffect(() => {
    async function load() {
      setLoading(true);
      const evs = await fetchEvents();
      setEvents(evs);
      const data = await fetchOrganiserAnalytics(selectedEventId);
      setAnalytics(data);
      setLoading(false);
    }
    load();
  }, [selectedEventId]);

  const handleCreateEvent = (e: React.FormEvent) => {
    e.preventDefault();
    alert(`🎉 New event listing "${newEventTitle}" created successfully in Draft mode.`);
    setShowCreateModal(false);
  };

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

      {/* Key Metric Stats Cards */}
      {loading || !analytics ? (
        <div className="py-20 text-center text-slate-400">
          <RefreshCw className="w-8 h-8 mx-auto animate-spin text-violet-400 mb-3" />
          <p className="font-semibold text-slate-300">Calculating revenue & occupancy metrics...</p>
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
                  className="w-full bg-[#0B0F17] border border-slate-700 rounded-xl px-3 py-2 text-slate-100 focus:outline-none focus:border-violet-500 transition"
                />
              </div>

              <div className="flex gap-3 pt-2">
                <button
                  type="submit"
                  className="flex-1 py-3 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 shadow-lg shadow-violet-600/30 transition"
                >
                  Create Listing (Draft)
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
