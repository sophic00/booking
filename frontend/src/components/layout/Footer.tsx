import React from "react";
import Link from "next/link";
import { Ticket, ShieldCheck, Zap, QrCode, RefreshCw, Mail, ArrowRight } from "lucide-react";

export function Footer() {
  return (
    <footer className="border-t border-slate-800 bg-[#0B0F17] text-slate-400 text-xs">
      {/* Guarantees Ribbon */}
      <div className="border-b border-slate-800/80 py-8 bg-[#0D121C]">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            <div className="flex items-start space-x-3.5">
              <div className="w-9 h-9 rounded-xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center shrink-0">
                <Zap className="w-4 h-4 text-indigo-400" />
              </div>
              <div>
                <h4 className="text-slate-200 font-semibold text-sm">10-Min Fair Hold</h4>
                <p className="text-slate-500 text-xs mt-0.5">
                  Exclusive seat reservation with automated TTL auto-release countdown.
                </p>
              </div>
            </div>

            <div className="flex items-start space-x-3.5">
              <div className="w-9 h-9 rounded-xl bg-violet-500/10 border border-violet-500/20 flex items-center justify-center shrink-0">
                <ShieldCheck className="w-4 h-4 text-violet-400" />
              </div>
              <div>
                <h4 className="text-slate-200 font-semibold text-sm">Zero Concurrency Clashes</h4>
                <p className="text-slate-500 text-xs mt-0.5">
                  Transactional atomic seat locks ensure no double bookings occur.
                </p>
              </div>
            </div>

            <div className="flex items-start space-x-3.5">
              <div className="w-9 h-9 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center shrink-0">
                <RefreshCw className="w-4 h-4 text-emerald-400" />
              </div>
              <div>
                <h4 className="text-slate-200 font-semibold text-sm">Auto-Waitlist Engine</h4>
                <p className="text-slate-500 text-xs mt-0.5">
                  Cancellations instantly re-allocate to waitlist with time-limited links.
                </p>
              </div>
            </div>

            <div className="flex items-start space-x-3.5">
              <div className="w-9 h-9 rounded-xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center shrink-0">
                <QrCode className="w-4 h-4 text-amber-400" />
              </div>
              <div>
                <h4 className="text-slate-200 font-semibold text-sm">Instant QR E-Tickets</h4>
                <p className="text-slate-500 text-xs mt-0.5">
                  Encrypted, scannable digital passes delivered straight to your email.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main Footer Links */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid grid-cols-1 md:grid-cols-5 gap-8">
          {/* Brand Info */}
          <div className="md:col-span-2 space-y-4">
            <Link href="/" className="flex items-center space-x-3 group">
              <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-indigo-600 to-violet-500 flex items-center justify-center shadow-lg shadow-indigo-500/25">
                <Ticket className="w-4 h-4 text-white" />
              </div>
              <span className="text-lg font-black tracking-tight text-white">
                VELVET<span className="text-indigo-400">SEATS</span>
              </span>
            </Link>
            <p className="text-slate-400 max-w-sm text-xs leading-relaxed">
              High-concurrency live event and cinema reservation platform. Designed for instant seat holds, verified QR ticketing, and automated queue re-allocation.
            </p>
            <div className="flex items-center space-x-2 text-slate-500 text-[11px] font-mono">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
              <span>System operational • Postgres Concurrency Engine active</span>
            </div>
          </div>

          {/* Customer Links */}
          <div>
            <h5 className="text-slate-200 font-semibold text-xs uppercase tracking-wider mb-3">
              Explore Shows
            </h5>
            <ul className="space-y-2">
              <li>
                <Link href="/" className="hover:text-indigo-400 transition">IMAX & Cinema</Link>
              </li>
              <li>
                <Link href="/" className="hover:text-indigo-400 transition">Stadium Concerts</Link>
              </li>
              <li>
                <Link href="/" className="hover:text-indigo-400 transition">Theatres & Musicals</Link>
              </li>
              <li>
                <Link href="/my-bookings" className="hover:text-indigo-400 transition">My Booking History</Link>
              </li>
              <li>
                <Link href="/" className="hover:text-indigo-400 transition">Waitlist Priority Hub</Link>
              </li>
            </ul>
          </div>

          {/* Organiser & Partners */}
          <div>
            <h5 className="text-slate-200 font-semibold text-xs uppercase tracking-wider mb-3">
              Organisers
            </h5>
            <ul className="space-y-2">
              <li>
                <Link href="/organiser" className="hover:text-indigo-400 transition">Organiser Dashboard</Link>
              </li>
              <li>
                <Link href="/organiser" className="hover:text-indigo-400 transition">Create Event Listing</Link>
              </li>
              <li>
                <Link href="/organiser" className="hover:text-indigo-400 transition">Real-Time Analytics</Link>
              </li>
              <li>
                <Link href="/organiser" className="hover:text-indigo-400 transition">Venue Layout Tools</Link>
              </li>
              <li>
                <Link href="/login" className="hover:text-indigo-400 transition">Partner Portal Sign In</Link>
              </li>
            </ul>
          </div>

          {/* Newsletter Subscribe */}
          <div>
            <h5 className="text-slate-200 font-semibold text-xs uppercase tracking-wider mb-3">
              Stay in the Loop
            </h5>
            <p className="text-slate-400 text-xs mb-3">
              Get notified first when sold-out headliners release waitlist seats.
            </p>
            <div className="flex items-center space-x-1.5">
              <div className="relative flex-1">
                <Mail className="w-3.5 h-3.5 text-slate-500 absolute left-3 top-2.5" />
                <input
                  type="email"
                  placeholder="Enter your email"
                  className="w-full bg-slate-900 border border-slate-700/80 rounded-lg pl-8 pr-3 py-1.5 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
                />
              </div>
              <button
                aria-label="Subscribe"
                className="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-medium transition shrink-0"
              >
                <ArrowRight className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="border-t border-slate-800/80 mt-10 pt-6 flex flex-col sm:flex-row items-center justify-between gap-4 text-[11px] text-slate-500">
          <p>© 2026 VelvetSeats Inc. Built for high-demand concerts and cinemas.</p>
          <div className="flex items-center space-x-4">
            <Link href="/" className="hover:text-slate-400 transition">Privacy Policy</Link>
            <span>•</span>
            <Link href="/" className="hover:text-slate-400 transition">Terms of Reservation</Link>
            <span>•</span>
            <Link href="/" className="hover:text-slate-400 transition">Anti-Scalping Rules</Link>
          </div>
        </div>
      </div>
    </footer>
  );
}
