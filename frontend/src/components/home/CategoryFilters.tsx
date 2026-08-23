"use client";

import React from "react";

interface CategoryFiltersProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

export function CategoryFilters({ activeTab, onTabChange }: CategoryFiltersProps) {
  const tabs = [
    { id: "ALL", label: "🔥 Trending Now" },
    { id: "MOVIE", label: "🎬 IMAX & Cinema" },
    { id: "CONCERT", label: "🎸 Concerts & Tours" },
    { id: "THEATRE", label: "🎭 Theater & Musicals" },
    { id: "SELLING_FAST", label: "⚡ Selling Out Fast" },
    { id: "WAITLIST", label: "🎟️ Active Waitlists" },
  ];

  return (
    <div className="flex items-center gap-2 overflow-x-auto pb-2 scrollbar-none">
      {tabs.map((tab) => {
        const isActive = activeTab === tab.id;
        return (
          <button
            key={tab.id}
            onClick={() => onTabChange(tab.id)}
            className={`px-4 py-2 rounded-full text-xs font-semibold whitespace-nowrap transition-all duration-200 shrink-0 ${
              isActive
                ? "bg-gradient-to-r from-indigo-600 to-violet-600 text-white shadow-lg shadow-indigo-500/25 ring-2 ring-indigo-400/40"
                : "bg-slate-900/80 hover:bg-slate-800 text-slate-300 border border-slate-700/60 hover:text-white"
            }`}
          >
            {tab.label}
          </button>
        );
      })}
    </div>
  );
}
