"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { ChevronLeft, ChevronRight, Calendar, MapPin, Ticket, Flame, Sparkles, AlertCircle, RefreshCw } from "lucide-react";
import { EventItem } from "../../lib/types";

interface HeroCarouselProps {
  events?: EventItem[];
  loading?: boolean;
  error?: string | null;
  onRetry?: () => void;
}

export function HeroCarousel({ events = [], loading = false, error = null, onRetry }: HeroCarouselProps) {
  const featured = events.slice(0, 5);
  const [currentIndex, setCurrentIndex] = useState(0);

  useEffect(() => {
    if (featured.length <= 1) return;
    const timer = setInterval(() => {
      setCurrentIndex((prev) => (prev + 1) % featured.length);
    }, 6000);
    return () => clearInterval(timer);
  }, [featured.length]);

  if (loading) {
    return (
      <div className="relative w-full h-[380px] sm:h-[480px] lg:h-[540px] rounded-3xl border border-slate-800 bg-[#0B0F17] flex flex-col items-center justify-center p-8 text-center space-y-4">
        <RefreshCw className="w-10 h-10 animate-spin text-indigo-500" />
        <p className="text-sm font-semibold text-slate-300">Loading spotlight headliners...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="relative w-full min-h-[300px] rounded-3xl border border-rose-500/40 bg-gradient-to-r from-rose-950/20 via-[#131A26] to-rose-950/20 p-8 flex flex-col items-center justify-center text-center space-y-4 shadow-xl">
        <div className="w-12 h-12 rounded-2xl bg-rose-500/20 text-rose-400 flex items-center justify-center">
          <AlertCircle className="w-6 h-6" />
        </div>
        <div className="space-y-1 max-w-md">
          <h3 className="text-base font-bold text-rose-200">Unable to load spotlight events</h3>
          <p className="text-xs text-slate-400">{error}</p>
        </div>
        {onRetry && (
          <button
            onClick={onRetry}
            className="px-4 py-2 rounded-xl text-xs font-bold text-white bg-rose-600 hover:bg-rose-500 transition"
          >
            Retry Connection
          </button>
        )}
      </div>
    );
  }

  if (featured.length === 0) {
    return (
      <div className="relative w-full h-[300px] rounded-3xl border border-slate-800 bg-[#131A26] p-8 flex flex-col items-center justify-center text-center space-y-3">
        <Ticket className="w-10 h-10 text-slate-600" />
        <h3 className="text-sm font-bold text-slate-300">No spotlight events published yet</h3>
        <p className="text-xs text-slate-500 max-w-md">
          Events published by organisers in the backend will automatically appear here.
        </p>
      </div>
    );
  }

  const current = featured[currentIndex] || featured[0];

  return (
    <div className="relative w-full h-[480px] lg:h-[540px] overflow-hidden rounded-3xl border border-slate-800/80 shadow-2xl bg-[#0B0F17] group">
      {/* Background Image Layer */}
      {current.poster_url || current.banner_url ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={current.banner_url || current.poster_url || ""}
          alt={current.title}
          className="absolute inset-0 w-full h-full object-cover object-center scale-105 transition-all duration-1000 ease-out brightness-75"
        />
      ) : (
        <div className="absolute inset-0 bg-gradient-to-tr from-indigo-950 via-slate-900 to-[#0B0F17]" />
      )}

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
            {current.event_type && (
              <span className="flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-violet-500/20 text-violet-300 border border-violet-500/40 backdrop-blur-md uppercase">
                {current.event_type}
              </span>
            )}
            {current.badge && (
              <span className="flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-rose-500/20 text-rose-300 border border-rose-500/40 backdrop-blur-md">
                <Flame className="w-3.5 h-3.5 mr-1.5 text-rose-400" />
                {current.badge}
              </span>
            )}
          </div>

          {/* Title */}
          <h1 className="text-3xl sm:text-4xl lg:text-5xl font-black text-white tracking-tight leading-tight drop-shadow-md">
            {current.title}
          </h1>

          {/* Description */}
          {current.description && (
            <p className="text-slate-300 text-sm sm:text-base line-clamp-2 max-w-xl font-normal leading-relaxed drop-shadow">
              {current.description}
            </p>
          )}

          {/* Event Metadata (Date, Venue) */}
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

            {current.venue_name && (
              <div className="flex items-center px-3 py-1.5 rounded-lg bg-slate-900/60 backdrop-blur-md border border-slate-700/60">
                <MapPin className="w-4 h-4 mr-2 text-violet-400" />
                <span>
                  {current.venue_name}
                  {current.venue_city ? `, ${current.venue_city}` : ""}
                </span>
              </div>
            )}
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
      {featured.length > 1 && (
        <>
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
        </>
      )}
    </div>
  );
}

