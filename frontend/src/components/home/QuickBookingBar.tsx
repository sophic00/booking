"use client";

import React, { useState } from "react";
import { Film, Music2, Calendar, MapPin, DollarSign, Search } from "lucide-react";

interface QuickBookingBarProps {
  onFilterChange: (filters: {
    eventType: string;
    city: string;
    priceRange: string;
  }) => void;
}

export function QuickBookingBar({ onFilterChange }: QuickBookingBarProps) {
  const [eventType, setEventType] = useState("ALL");
  const [dateFilter, setDateFilter] = useState("ALL");
  const [city, setCity] = useState("ALL");
  const [priceRange, setPriceRange] = useState("ALL");

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    onFilterChange({
      eventType,
      city,
      priceRange,
    });
  };

  return (
    <div className="relative -mt-8 z-30 max-w-6xl mx-auto px-4">
      <form
        onSubmit={handleSearch}
        className="glass-panel rounded-2xl p-4 sm:p-5 shadow-2xl border border-slate-700/80 bg-[#131A26]/90 backdrop-blur-xl"
      >
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {/* Event Type Filter */}
          <div className="space-y-1">
            <label className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider flex items-center">
              <Film className="w-3.5 h-3.5 mr-1.5 text-indigo-400" />
              Event Category
            </label>
            <select
              value={eventType}
              onChange={(e) => setEventType(e.target.value)}
              className="w-full bg-[#0B0F17] border border-slate-700/80 rounded-xl px-3 py-2 text-xs font-medium text-slate-200 focus:outline-none focus:border-indigo-500 transition cursor-pointer"
            >
              <option value="ALL">All Entertainment</option>
              <option value="MOVIE">🎬 Movies & IMAX 70mm</option>
              <option value="CONCERT">🎸 Live Concerts & Tours</option>
              <option value="THEATRE">🎭 Theater & Musicals</option>
              <option value="SPORTS">🏟️ Stadium Sports</option>
            </select>
          </div>

          {/* Date Picker Range */}
          <div className="space-y-1">
            <label className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider flex items-center">
              <Calendar className="w-3.5 h-3.5 mr-1.5 text-violet-400" />
              Date & Schedule
            </label>
            <select
              value={dateFilter}
              onChange={(e) => setDateFilter(e.target.value)}
              className="w-full bg-[#0B0F17] border border-slate-700/80 rounded-xl px-3 py-2 text-xs font-medium text-slate-200 focus:outline-none focus:border-indigo-500 transition cursor-pointer"
            >
              <option value="ALL">Any Upcoming Date</option>
              <option value="TODAY">Tonight / Today</option>
              <option value="WEEKEND">This Weekend</option>
              <option value="MONTH">Next 30 Days</option>
            </select>
          </div>

          {/* City / Location */}
          <div className="space-y-1">
            <label className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider flex items-center">
              <MapPin className="w-3.5 h-3.5 mr-1.5 text-rose-400" />
              Venue Location
            </label>
            <select
              value={city}
              onChange={(e) => setCity(e.target.value)}
              className="w-full bg-[#0B0F17] border border-slate-700/80 rounded-xl px-3 py-2 text-xs font-medium text-slate-200 focus:outline-none focus:border-indigo-500 transition cursor-pointer"
            >
              <option value="ALL">All Cities</option>
              <option value="San Francisco">San Francisco, CA</option>
              <option value="New York">New York, NY</option>
              <option value="London">London, UK</option>
              <option value="East Rutherford">East Rutherford, NJ</option>
            </select>
          </div>

          {/* Price Range & CTA Button */}
          <div className="space-y-1 flex flex-col justify-end">
            <label className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider flex items-center">
              <DollarSign className="w-3.5 h-3.5 mr-1 text-emerald-400" />
              Price Tier
            </label>
            <div className="flex gap-2">
              <select
                value={priceRange}
                onChange={(e) => setPriceRange(e.target.value)}
                className="flex-1 bg-[#0B0F17] border border-slate-700/80 rounded-xl px-2.5 py-2 text-xs font-medium text-slate-200 focus:outline-none focus:border-indigo-500 transition cursor-pointer"
              >
                <option value="ALL">Any Price</option>
                <option value="BUDGET">Standard (&lt;$50)</option>
                <option value="MID">Premium ($50-$150)</option>
                <option value="VIP">VIP Lounge ($150+)</option>
              </select>

              <button
                type="submit"
                className="px-4 py-2 rounded-xl bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white font-bold text-xs shadow-lg shadow-indigo-600/25 transition shrink-0 flex items-center"
              >
                <Search className="w-3.5 h-3.5 mr-1" />
                Find Seats
              </button>
            </div>
          </div>
        </div>

        {/* Live Concurrency Info Strip */}
        <div className="mt-3 pt-3 border-t border-slate-800/80 flex flex-wrap items-center justify-between gap-2 text-[11px] text-slate-400">
          <div className="flex items-center space-x-2">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
            <span>
              <b>Live Anti-Scalping Engine:</b> 10-Minute Lock TTL active on all seat selections.
            </span>
          </div>
          <div className="flex items-center space-x-4 text-slate-500">
            <span>⚡ Zero-Double-Booking Guarantee</span>
            <span>•</span>
            <span>🔄 Instant Waitlist Auto-Pass</span>
          </div>
        </div>
      </form>
    </div>
  );
}
