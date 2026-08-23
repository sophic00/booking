"use client";

import React, { useState, useEffect, useMemo } from "react";
import { HeroCarousel } from "../components/home/HeroCarousel";
import { QuickBookingBar } from "../components/home/QuickBookingBar";
import { CategoryFilters } from "../components/home/CategoryFilters";
import { EventCard } from "../components/home/EventCard";
import { FeatureHighlights } from "../components/home/FeatureHighlights";
import { OrganiserCallout } from "../components/home/OrganiserCallout";
import { fetchEvents } from "../lib/api";
import { EventItem } from "../lib/types";
import { Flame, Ticket, Search, RefreshCw } from "lucide-react";

export default function HomePage() {
  const [events, setEvents] = useState<EventItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState("ALL");
  const [searchKeyword, setSearchKeyword] = useState("");
  const [quickFilters, setQuickFilters] = useState<{
    eventType: string;
    city: string;
    priceRange: string;
  }>({
    eventType: "ALL",
    city: "ALL",
    priceRange: "ALL",
  });

  useEffect(() => {
    async function load() {
      setLoading(true);
      const data = await fetchEvents();
      setEvents(data);
      setLoading(false);
    }
    load();
  }, []);

  const filteredEvents = useMemo(() => {
    return events.filter((e) => {
      // Tab filter
      if (activeTab === "MOVIE" && e.event_type !== "MOVIE") return false;
      if (activeTab === "CONCERT" && e.event_type !== "CONCERT") return false;
      if (activeTab === "THEATRE" && e.event_type !== "THEATRE") return false;
      if (activeTab === "SELLING_FAST" && (!e.badge || !e.badge.toLowerCase().includes("fast") && !e.badge.toLowerCase().includes("hold"))) return false;
      if (activeTab === "WAITLIST" && (!e.badge || !e.badge.toLowerCase().includes("sold out") && !e.badge.toLowerCase().includes("waitlist"))) return false;

      // Quick bar filters
      if (quickFilters.eventType !== "ALL" && e.event_type !== quickFilters.eventType) {
        return false;
      }
      if (quickFilters.city !== "ALL" && e.venue_city && !e.venue_city.includes(quickFilters.city)) {
        return false;
      }
      if (quickFilters.priceRange === "BUDGET" && e.min_price && e.min_price > 50) {
        return false;
      }
      if (quickFilters.priceRange === "MID" && e.min_price && (e.min_price < 50 || e.min_price > 150)) {
        return false;
      }
      if (quickFilters.priceRange === "VIP" && e.max_price && e.max_price < 150) {
        return false;
      }

      // Keyword search
      if (searchKeyword.trim()) {
        const q = searchKeyword.toLowerCase();
        const matchTitle = e.title.toLowerCase().includes(q);
        const matchVenue = e.venue_name?.toLowerCase().includes(q) || false;
        const matchCity = e.venue_city?.toLowerCase().includes(q) || false;
        if (!matchTitle && !matchVenue && !matchCity) return false;
      }

      return true;
    });
  }, [events, activeTab, quickFilters, searchKeyword]);

  return (
    <div className="space-y-12 pb-16">
      {/* 1. Cinematic Hero Spotlight Carousel */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-6">
        <HeroCarousel />
      </section>

      {/* 2. Floating Quick Booking Search Bar */}
      <QuickBookingBar onFilterChange={(filters) => setQuickFilters(filters)} />

      {/* 3. Main Live Events Grid Section */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 space-y-6 pt-4">
        {/* Section Header & Tabs */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800/80 pb-4">
          <div>
            <div className="flex items-center space-x-2 text-xs font-mono font-bold text-indigo-400 uppercase tracking-wider">
              <Flame className="w-3.5 h-3.5" />
              <span>LIVE SHOWTIMES & HEADLINERS</span>
            </div>
            <h2 className="text-2xl sm:text-3xl font-black text-white tracking-tight mt-1">
              Now Showing & Upcoming Tours
            </h2>
          </div>

          <CategoryFilters activeTab={activeTab} onTabChange={(tab) => setActiveTab(tab)} />
        </div>

        {/* Live Filter Bar Status */}
        {(quickFilters.eventType !== "ALL" || quickFilters.city !== "ALL" || quickFilters.priceRange !== "ALL" || searchKeyword) && (
          <div className="flex items-center justify-between px-4 py-2.5 rounded-xl bg-slate-900/80 border border-slate-800 text-xs text-slate-300">
            <div className="flex items-center space-x-2">
              <span className="text-slate-500 font-medium">Active Filters:</span>
              {quickFilters.eventType !== "ALL" && (
                <span className="px-2 py-0.5 rounded bg-indigo-600/30 text-indigo-300 font-mono">
                  Type: {quickFilters.eventType}
                </span>
              )}
              {quickFilters.city !== "ALL" && (
                <span className="px-2 py-0.5 rounded bg-violet-600/30 text-violet-300 font-mono">
                  City: {quickFilters.city}
                </span>
              )}
              {quickFilters.priceRange !== "ALL" && (
                <span className="px-2 py-0.5 rounded bg-emerald-600/30 text-emerald-300 font-mono">
                  Tier: {quickFilters.priceRange}
                </span>
              )}
            </div>
            <button
              onClick={() => {
                setQuickFilters({ eventType: "ALL", city: "ALL", priceRange: "ALL" });
                setActiveTab("ALL");
                setSearchKeyword("");
              }}
              className="text-xs text-rose-400 hover:text-rose-300 transition underline cursor-pointer"
            >
              Reset Filters
            </button>
          </div>
        )}

        {/* Grid of Event Cards */}
        {loading ? (
          <div className="py-20 text-center text-slate-400">
            <RefreshCw className="w-8 h-8 mx-auto animate-spin text-indigo-400 mb-3" />
            <p className="font-semibold text-slate-300">Loading live shows & seat statuses...</p>
          </div>
        ) : filteredEvents.length === 0 ? (
          <div className="py-20 text-center glass-panel rounded-3xl border border-slate-800 p-8 space-y-3">
            <Ticket className="w-12 h-12 mx-auto text-slate-600" />
            <h3 className="text-lg font-bold text-slate-200">No events matched your filter criteria</h3>
            <p className="text-xs text-slate-400 max-w-md mx-auto">
              Try adjusting your category pills or price range to explore more available cinema and stadium performances.
            </p>
            <button
              onClick={() => {
                setActiveTab("ALL");
                setQuickFilters({ eventType: "ALL", city: "ALL", priceRange: "ALL" });
              }}
              className="mt-2 px-4 py-2 rounded-xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-500 transition"
            >
              Show All Events
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredEvents.map((event) => (
              <EventCard key={event.id} event={event} />
            ))}
          </div>
        )}
      </section>

      {/* 4. Core System Guarantees & Features Showcase */}
      <FeatureHighlights />

      {/* 5. Organiser & Venue Manager Banner */}
      <OrganiserCallout />
    </div>
  );
}
