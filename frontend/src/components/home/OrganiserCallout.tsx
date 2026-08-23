import React from "react";
import Link from "next/link";
import { Building2, BarChart3, Users, ArrowRight, Shield, Timer } from "lucide-react";

export function OrganiserCallout() {
  return (
    <section className="py-12">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="relative overflow-hidden rounded-3xl border border-indigo-500/30 bg-gradient-to-r from-indigo-950/60 via-[#131A26] to-violet-950/60 p-8 sm:p-12 shadow-2xl backdrop-blur-xl">
          {/* Subtle glowing elements */}
          <div className="absolute -top-24 -right-24 w-72 h-72 bg-indigo-500/20 rounded-full blur-3xl pointer-events-none" />
          <div className="absolute -bottom-24 -left-24 w-72 h-72 bg-violet-500/20 rounded-full blur-3xl pointer-events-none" />

          <div className="relative z-10 grid grid-cols-1 lg:grid-cols-3 gap-8 items-center">
            <div className="lg:col-span-2 space-y-4">
              <div className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-indigo-500/20 text-indigo-300 border border-indigo-500/40">
                <Building2 className="w-3.5 h-3.5 mr-1.5" />
                VENUE MANAGERS & EVENT ORGANISERS
              </div>

              <h2 className="text-2xl sm:text-3xl font-black text-white tracking-tight">
                Host Your Next Movie or Concert on VelvetSeats
              </h2>

              <p className="text-slate-300 text-xs sm:text-sm max-w-xl leading-relaxed">
                Create event listings with custom venue seat layouts, configure per-category pricing tiers, customize reservation hold timers, and monitor real-time ticket sales and waitlist conversions.
              </p>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pt-2">
                <div className="flex items-center space-x-2 text-xs text-slate-300">
                  <BarChart3 className="w-4 h-4 text-indigo-400 shrink-0" />
                  <span>Live Revenue & Occupancy</span>
                </div>
                <div className="flex items-center space-x-2 text-xs text-slate-300">
                  <Users className="w-4 h-4 text-violet-400 shrink-0" />
                  <span>Automated Waitlist Queues</span>
                </div>
                <div className="flex items-center space-x-2 text-xs text-slate-300">
                  <Timer className="w-4 h-4 text-emerald-400 shrink-0" />
                  <span>Custom Cart Hold Controls</span>
                </div>
              </div>
            </div>

            <div className="flex flex-col sm:flex-row lg:flex-col gap-3 justify-center">
              <Link
                href="/organiser"
                className="px-6 py-3.5 rounded-xl font-bold text-xs sm:text-sm text-white bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 shadow-xl shadow-indigo-600/30 transition text-center flex items-center justify-center"
              >
                Access Organiser Portal
                <ArrowRight className="w-4 h-4 ml-2" />
              </Link>
              <Link
                href="/signup"
                className="px-6 py-3.5 rounded-xl font-semibold text-xs sm:text-sm text-slate-200 hover:text-white bg-slate-900/80 hover:bg-slate-800 border border-slate-700/80 transition text-center"
              >
                Create Partner Account
              </Link>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
