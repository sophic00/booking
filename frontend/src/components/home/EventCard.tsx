"use client";

import React from "react";
import Link from "next/link";
import { Calendar, MapPin, Ticket, Flame, Users, Sparkles } from "lucide-react";
import { EventItem } from "../../lib/types";

interface EventCardProps {
  event: EventItem;
}

export function EventCard({ event }: EventCardProps) {
  const isSoldOut = event.seats_left === 0 || event.badge?.includes("Sold Out");

  return (
    <div className="glass-panel glass-panel-hover rounded-2xl overflow-hidden flex flex-col justify-between group border border-slate-800 hover:border-indigo-500/40 transition-all duration-300">
      {/* Poster Image Container */}
      <div className="relative aspect-[16/10] w-full overflow-hidden bg-slate-900">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          src={event.poster_url || ""}
          alt={event.title}
          className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
        />

        {/* Gradient Mask for readable text */}
        <div className="absolute inset-0 bg-gradient-to-t from-[#131A26] via-transparent to-black/30" />

        {/* Top Badges (Date & Category) */}
        <div className="absolute top-3 left-3 flex items-center gap-2">
          <span className="px-2.5 py-1 rounded-lg bg-black/60 backdrop-blur-md border border-white/10 text-[11px] font-bold text-white uppercase tracking-wider">
            {new Date(event.start_time).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
          </span>
          <span className="px-2.5 py-1 rounded-lg bg-indigo-600/80 backdrop-blur-md text-[11px] font-semibold text-white uppercase tracking-wider">
            {event.event_type}
          </span>
        </div>

        {/* Status Badge overlay */}
        {event.badge && (
          <div className="absolute bottom-3 left-3 right-3">
            <span
              className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold backdrop-blur-md shadow-md ${
                isSoldOut
                  ? "bg-rose-500/20 text-rose-300 border border-rose-500/30"
                  : event.badge_type === "warning"
                  ? "bg-amber-500/20 text-amber-300 border border-amber-500/30"
                  : "bg-indigo-500/20 text-indigo-300 border border-indigo-500/30"
              }`}
            >
              {isSoldOut ? (
                <Users className="w-3 h-3 mr-1.5 text-rose-400" />
              ) : (
                <Flame className="w-3 h-3 mr-1.5 text-amber-400" />
              )}
              {event.badge}
            </span>
          </div>
        )}
      </div>

      {/* Card Body */}
      <div className="p-5 flex-1 flex flex-col justify-between space-y-4">
        <div>
          {/* Title */}
          <Link href={`/events/${event.id}`}>
            <h3 className="text-base font-bold text-slate-100 group-hover:text-indigo-300 transition-colors line-clamp-1">
              {event.title}
            </h3>
          </Link>

          {/* Venue & Time info */}
          <div className="mt-2 space-y-1.5 text-xs text-slate-400">
            <div className="flex items-center">
              <MapPin className="w-3.5 h-3.5 mr-1.5 text-slate-500 shrink-0" />
              <span className="truncate">{event.venue_name}, {event.venue_city}</span>
            </div>
            <div className="flex items-center">
              <Calendar className="w-3.5 h-3.5 mr-1.5 text-slate-500 shrink-0" />
              <span>
                {new Date(event.start_time).toLocaleDateString("en-US", {
                  weekday: "short",
                  hour: "numeric",
                  minute: "2-digit",
                })}
              </span>
            </div>
          </div>
        </div>

        {/* Pricing & CTA Row */}
        <div className="pt-3 border-t border-slate-800/80 flex items-center justify-between gap-3">
          <div>
            <span className="text-[11px] text-slate-400 block">Tier Pricing</span>
            <span className="text-sm font-black text-slate-100">
              ${event.min_price} <span className="text-xs font-normal text-slate-400">- ${event.max_price}</span>
            </span>
          </div>

          <Link
            href={`/events/${event.id}`}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition flex items-center ${
              isSoldOut
                ? "bg-slate-800 hover:bg-slate-700 text-rose-300 border border-rose-500/30"
                : "bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 text-white shadow-md shadow-indigo-600/20"
            }`}
          >
            {isSoldOut ? (
              <>
                <Users className="w-3.5 h-3.5 mr-1.5" />
                Join Waitlist
              </>
            ) : (
              <>
                <Ticket className="w-3.5 h-3.5 mr-1.5" />
                Select Seats
              </>
            )}
          </Link>
        </div>
      </div>
    </div>
  );
}
