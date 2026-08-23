"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { ChevronLeft, ChevronRight, Calendar, MapPin, Ticket, Flame, Sparkles } from "lucide-react";
import { MOCK_EVENTS } from "../../data/mockData";

export function HeroCarousel() {
  const featured = MOCK_EVENTS.slice(0, 3);
  const [currentIndex, setCurrentIndex] = useState(0);

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentIndex((prev) => (prev + 1) % featured.length);
    }, 6000);
    return () => clearInterval(timer);
  }, [featured.length]);

  const current = featured[currentIndex];

  return (
    <div className="relative w-full h-[520px] lg:h-[600px] overflow-hidden rounded-3xl border border-slate-800/80 shadow-2xl bg-[#0B0F17] group">
      {/* Background Image Layer with Atmospheric Overlays */}
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={current.banner_url || current.poster_url || ""}
        alt={current.title}
        className="absolute inset-0 w-full h-full object-cover object-center scale-105 transition-all duration-1000 ease-out brightness-75"
      />

      {/* Cinematic Vignette & Gradient Mask */}
      <div className="absolute inset-0 bg-gradient-to-t from-[#0B0F17] via-[#0B0F17]/60 to-transparent" />
      <div className="absolute inset-0 bg-gradient-to-r from-[#0B0F17] via-[#0B0F17]/80 to-transparent" />
      <div className="absolute inset-0 bg-radial-at-c from-transparent via-black/20 to-black/80" />

      {/* Content Container */}
      <div className="relative h-full max-w-7xl mx-auto px-6 sm:px-10 lg:px-12 flex flex-col justify-end pb-12 z-10">
        <div className="max-w-2xl space-y-4">
          {/* Badges Row */}
          <div className="flex flex-wrap items-center gap-2.5">
            <span className="flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-indigo-500/20 text-indigo-300 border border-indigo-500/40 backdrop-blur-md">
              <Sparkles className="w-3.5 h-3.5 mr-1.5 text-indigo-400" />
              SPOTLIGHT HEADLINER
            </span>
            <span className="flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-rose-500/20 text-rose-300 border border-rose-500/40 backdrop-blur-md">
              <Flame className="w-3.5 h-3.5 mr-1.5 text-rose-400" />
              {current.badge || "High Demand Event"}
            </span>
          </div>

          {/* Title */}
          <h1 className="text-3xl sm:text-4xl lg:text-5xl font-black text-white tracking-tight leading-tight drop-shadow-md">
            {current.title}
          </h1>

          {/* Description */}
          <p className="text-slate-300 text-sm sm:text-base line-clamp-2 max-w-xl font-normal leading-relaxed drop-shadow">
            {current.description}
          </p>

          {/* Event Metadata (Date, Venue, Price) */}
          <div className="flex flex-wrap items-center gap-4 text-xs sm:text-sm text-slate-200 py-1">
            <div className="flex items-center px-3 py-1.5 rounded-lg bg-slate-900/60 backdrop-blur-md border border-slate-700/60">
              <Calendar className="w-4 h-4 mr-2 text-indigo-400" />
              <span>
                {new Date(current.start_time).toLocaleDateString("en-US", {
                  weekday: "short",
                  month: "short",
                  day: "numeric",
                  hour: "numeric",
                  minute: "2-digit",
                })}
              </span>
            </div>

            <div className="flex items-center px-3 py-1.5 rounded-lg bg-slate-900/60 backdrop-blur-md border border-slate-700/60">
              <MapPin className="w-4 h-4 mr-2 text-violet-400" />
              <span>{current.venue_name}, {current.venue_city}</span>
            </div>

            <div className="flex items-center px-3 py-1.5 rounded-lg bg-slate-900/60 backdrop-blur-md border border-slate-700/60 font-medium">
              <span className="text-slate-400 mr-1.5">Tiers:</span>
              <span className="text-emerald-400 font-bold">${current.min_price} - ${current.max_price}</span>
            </div>
          </div>

          {/* Call to Action Buttons */}
          <div className="flex items-center gap-3 pt-2">
            <Link
              href={`/events/${current.id}`}
              className="px-6 py-3 rounded-xl font-bold text-sm text-white bg-gradient-to-r from-indigo-600 via-indigo-500 to-violet-600 hover:from-indigo-500 hover:to-violet-500 shadow-xl shadow-indigo-600/30 transition transform hover:-translate-y-0.5 flex items-center"
            >
              <Ticket className="w-4 h-4 mr-2" />
              Select Seats from Map
            </Link>

            <Link
              href={`/events/${current.id}`}
              className="px-5 py-3 rounded-xl font-semibold text-sm text-slate-200 hover:text-white bg-slate-900/80 hover:bg-slate-800 border border-slate-700/80 backdrop-blur-md transition"
            >
              Show Details
            </Link>
          </div>
        </div>
      </div>

      {/* Navigation Controls (Arrows) */}
      <button
        onClick={() => setCurrentIndex((prev) => (prev - 1 + featured.length) % featured.length)}
        aria-label="Previous Slide"
        className="absolute left-4 top-1/2 -translate-y-1/2 p-2.5 rounded-full bg-black/40 hover:bg-black/70 border border-white/10 text-white backdrop-blur-md opacity-0 group-hover:opacity-100 transition z-20"
      >
        <ChevronLeft className="w-5 h-5" />
      </button>

      <button
        onClick={() => setCurrentIndex((prev) => (prev + 1) % featured.length)}
        aria-label="Next Slide"
        className="absolute right-4 top-1/2 -translate-y-1/2 p-2.5 rounded-full bg-black/40 hover:bg-black/70 border border-white/10 text-white backdrop-blur-md opacity-0 group-hover:opacity-100 transition z-20"
      >
        <ChevronRight className="w-5 h-5" />
      </button>

      {/* Slide Indicators */}
      <div className="absolute bottom-4 right-8 flex items-center space-x-2 z-20">
        {featured.map((_, idx) => (
          <button
            key={idx}
            onClick={() => setCurrentIndex(idx)}
            aria-label={`Go to slide ${idx + 1}`}
            className={`h-1.5 rounded-full transition-all duration-300 ${
              currentIndex === idx ? "w-8 bg-indigo-500" : "w-2 bg-slate-600 hover:bg-slate-400"
            }`}
          />
        ))}
      </div>
    </div>
  );
}
