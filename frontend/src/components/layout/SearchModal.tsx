"use client";

import React, { useState, useEffect, useRef } from "react";
import { Search, X, Calendar, MapPin, ArrowRight, Ticket, RefreshCw, AlertCircle } from "lucide-react";
import { EventItem } from "../../lib/types";
import { fetchEvents } from "../../lib/api";
import Link from "next/link";

interface SearchModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export function SearchModal({ isOpen, onClose }: SearchModalProps) {
  const [query, setQuery] = useState("");
  const [events, setEvents] = useState<EventItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isOpen) {
      setTimeout(() => inputRef.current?.focus(), 50);
      searchEvents("");
    } else {
      setQuery("");
      setEvents([]);
      setError(null);
    }
  }, [isOpen]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        if (isOpen) onClose();
      }
      if (e.key === "Escape" && isOpen) {
        onClose();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose]);

  const searchEvents = async (searchQuery: string) => {
    setLoading(true);
    setError(null);
    try {
      const data = await fetchEvents({ search: searchQuery.trim() || undefined, limit: 10 });
      setEvents(data);
    } catch (err: any) {
      setError(err.message || "Failed to search events from backend API");
      setEvents([]);
    } finally {
      setLoading(false);
    }
  };

  const handleQueryChange = (val: string) => {
    setQuery(val);
    searchEvents(val);
  };

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-20 px-4 bg-black/80 backdrop-blur-md transition-opacity"
      onClick={onClose}
    >
      <div
        className="w-full max-w-2xl bg-[#131A26] border border-slate-700/60 rounded-2xl shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Search Input Bar */}
        <div className="flex items-center px-4 py-3.5 border-b border-slate-800 bg-[#0B0F17]/70">
          <Search className="w-5 h-5 text-indigo-400 mr-3 shrink-0" />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => handleQueryChange(e.target.value)}
            placeholder="Search movies, concerts, artists, venues..."
            className="w-full bg-transparent text-slate-100 placeholder-slate-400 text-base focus:outline-none"
          />
          {query && (
            <button
              onClick={() => handleQueryChange("")}
              className="text-slate-400 hover:text-slate-200 p-1 mr-1 text-xs"
            >
              Clear
            </button>
          )}
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-200 p-1.5 rounded-lg hover:bg-slate-800 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Quick Filter Tags */}
        <div className="flex items-center gap-2 px-4 py-2.5 bg-slate-900/50 border-b border-slate-800/80 text-xs text-slate-400 overflow-x-auto">
          <span className="shrink-0 text-slate-500 font-medium">Quick Filters:</span>
          {["All", "Concert", "Movie", "Theatre", "Sports"].map((tag) => (
            <button
              key={tag}
              onClick={() => handleQueryChange(tag === "All" ? "" : tag)}
              className="px-2.5 py-1 rounded-full bg-slate-800 hover:bg-slate-700 text-slate-300 transition shrink-0"
            >
              {tag}
            </button>
          ))}
        </div>

        {/* Results List */}
        <div className="max-h-96 overflow-y-auto p-3 space-y-2">
          {loading ? (
            <div className="py-12 text-center text-slate-400">
              <RefreshCw className="w-8 h-8 mx-auto mb-3 animate-spin text-indigo-400" />
              <p className="text-xs font-semibold text-slate-300">Searching live events...</p>
            </div>
          ) : error ? (
            <div className="p-6 text-center text-rose-300 space-y-2">
              <AlertCircle className="w-8 h-8 mx-auto text-rose-400 mb-1" />
              <p className="text-xs font-semibold">{error}</p>
              <button
                onClick={() => searchEvents(query)}
                className="px-3 py-1.5 rounded-lg text-xs bg-rose-600 hover:bg-rose-500 text-white font-bold transition"
              >
                Retry
              </button>
            </div>
          ) : events.length === 0 ? (
            <div className="py-12 text-center text-slate-400">
              <Ticket className="w-10 h-10 mx-auto mb-3 text-slate-600" />
              <p className="font-medium text-slate-300">No events found</p>
              <p className="text-xs text-slate-500 mt-1">
                {query ? `No events matched "${query}" on the backend` : "No published events available right now"}
              </p>
            </div>
          ) : (
            events.map((event) => (
              <Link
                key={event.id}
                href={`/events/${event.id}`}
                onClick={onClose}
                className="flex items-center justify-between p-3 rounded-xl hover:bg-slate-800/60 border border-transparent hover:border-indigo-500/20 transition group"
              >
                <div className="flex items-center space-x-3.5">
                  <div className="w-12 h-14 rounded-lg bg-slate-800 overflow-hidden shrink-0 border border-slate-700">
                    {event.poster_url ? (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img
                        src={event.poster_url}
                        alt={event.title}
                        className="w-full h-full object-cover group-hover:scale-105 transition"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center bg-slate-800 text-slate-500 text-[10px]">
                        Event
                      </div>
                    )}
                  </div>
                  <div>
                    <div className="flex items-center space-x-2">
                      <span className="text-xs font-semibold px-2 py-0.5 rounded bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 uppercase">
                        {event.event_type}
                      </span>
                      <h4 className="text-sm font-semibold text-slate-100 group-hover:text-indigo-300 transition">
                        {event.title}
                      </h4>
                    </div>
                    <div className="flex items-center gap-3 text-xs text-slate-400 mt-1">
                      {event.venue_name && (
                        <span className="flex items-center">
                          <MapPin className="w-3 h-3 mr-1 text-slate-500" />
                          {event.venue_name}
                        </span>
                      )}
                      <span className="flex items-center">
                        <Calendar className="w-3 h-3 mr-1 text-slate-500" />
                        {new Date(event.start_time).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
                      </span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center space-x-3">
                  <ArrowRight className="w-4 h-4 text-slate-500 group-hover:text-indigo-400 group-hover:translate-x-1 transition" />
                </div>
              </Link>
            ))
          )}
        </div>

        {/* Footer Hint */}
        <div className="px-4 py-2.5 bg-[#0B0F17] border-t border-slate-800 text-[11px] text-slate-500 flex justify-between items-center">
          <span>Live API Search</span>
          <span className="flex items-center gap-1">
            <kbd className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700 text-[10px]">ESC</kbd> to close
          </span>
        </div>
      </div>
    </div>
  );
}

