"use client";

import React, { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Ticket,
  Search,
  MapPin,
  ChevronDown,
  Bell,
  User as UserIcon,
  LogOut,
  CalendarCheck,
  Building2,
  Sparkles,
  ShieldCheck,
} from "lucide-react";
import { useAuth } from "../../context/AuthContext";
import { SearchModal } from "./SearchModal";

export function Navbar() {
  const pathname = usePathname();
  const { user, logout, switchRole } = useAuth();
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [selectedCity, setSelectedCity] = useState("San Francisco, CA");
  const [isCityDropdownOpen, setIsCityDropdownOpen] = useState(false);
  const [isProfileOpen, setIsProfileOpen] = useState(false);

  const cities = ["San Francisco, CA", "New York, NY", "London, UK", "Los Angeles, CA", "Chicago, IL"];

  return (
    <>
      <header className="sticky top-0 z-40 w-full glass-nav backdrop-blur-xl border-b border-slate-800/80 transition">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-18 flex items-center justify-between gap-4">
          {/* Brand Logo */}
          <div className="flex items-center space-x-6">
            <Link href="/" className="flex items-center space-x-3 group">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-600 to-violet-500 flex items-center justify-center shadow-lg shadow-indigo-500/25 group-hover:scale-105 transition">
                <Ticket className="w-5 h-5 text-white" />
              </div>
              <div>
                <span className="text-xl font-black tracking-tight bg-gradient-to-r from-white via-slate-100 to-slate-300 bg-clip-text text-transparent">
                  VELVET<span className="text-indigo-400">SEATS</span>
                </span>
                <span className="hidden sm:block text-[10px] tracking-wider uppercase text-slate-400 font-mono">
                  Official Tickets & Live Shows
                </span>
              </div>
            </Link>

            {/* City Selector */}
            <div className="relative hidden md:block">
              <button
                onClick={() => setIsCityDropdownOpen(!isCityDropdownOpen)}
                className="flex items-center space-x-2 px-3 py-1.5 rounded-lg bg-slate-800/60 hover:bg-slate-800 border border-slate-700/50 text-xs text-slate-300 transition"
              >
                <MapPin className="w-3.5 h-3.5 text-indigo-400" />
                <span className="font-medium">{selectedCity}</span>
                <ChevronDown className="w-3 h-3 text-slate-400" />
              </button>

              {isCityDropdownOpen && (
                <div className="absolute top-full left-0 mt-1.5 w-48 bg-[#131A26] border border-slate-700 rounded-xl shadow-xl py-1.5 z-50 animate-in fade-in zoom-in-95">
                  <div className="px-3 py-1 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                    Select Location
                  </div>
                  {cities.map((city) => (
                    <button
                      key={city}
                      onClick={() => {
                        setSelectedCity(city);
                        setIsCityDropdownOpen(false);
                      }}
                      className={`w-full text-left px-3 py-2 text-xs transition flex items-center justify-between ${
                        selectedCity === city
                          ? "bg-indigo-600/20 text-indigo-300 font-semibold"
                          : "text-slate-300 hover:bg-slate-800"
                      }`}
                    >
                      {city}
                      {selectedCity === city && <span className="w-1.5 h-1.5 rounded-full bg-indigo-400" />}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Quick Search Bar Trigger */}
          <div className="hidden lg:flex flex-1 max-w-md mx-4">
            <button
              onClick={() => setIsSearchOpen(true)}
              className="w-full flex items-center justify-between px-3.5 py-2 rounded-xl bg-slate-900/80 hover:bg-slate-800/80 border border-slate-700/60 text-slate-400 text-xs transition shadow-inner"
            >
              <span className="flex items-center">
                <Search className="w-4 h-4 mr-2 text-indigo-400" />
                Search movies, concerts, artists, venues...
              </span>
              <kbd className="px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700 text-[10px] font-mono shadow-sm">
                ⌘K
              </kbd>
            </button>
          </div>

          {/* Navigation Links & Role Actions */}
          <div className="flex items-center space-x-3 sm:space-x-4">
            {/* Mobile Search Button */}
            <button
              onClick={() => setIsSearchOpen(true)}
              className="lg:hidden p-2 rounded-lg bg-slate-800/60 text-slate-300 hover:text-white"
            >
              <Search className="w-5 h-5" />
            </button>

            {/* Role Nav Links */}
            <nav className="hidden xl:flex items-center space-x-5 text-xs font-medium text-slate-300">
              <Link
                href="/"
                className={`transition hover:text-indigo-300 ${
                  pathname === "/" ? "text-indigo-400 font-semibold" : ""
                }`}
              >
                Explore
              </Link>
              <Link
                href="/my-bookings"
                className={`transition hover:text-indigo-300 ${
                  pathname === "/my-bookings" ? "text-indigo-400 font-semibold" : ""
                }`}
              >
                My Tickets
              </Link>
              <Link
                href="/organiser"
                className={`transition hover:text-indigo-300 ${
                  pathname === "/organiser" ? "text-indigo-400 font-semibold" : ""
                }`}
              >
                Organiser Portal
              </Link>
              <Link
                href="/admin"
                className={`transition hover:text-amber-300 ${
                  pathname === "/admin" ? "text-amber-400 font-semibold" : ""
                }`}
              >
                Admin
              </Link>
            </nav>

            {/* Notification Bell */}
            <div className="relative">
              <button
                aria-label="Notifications"
                className="p-2 rounded-lg bg-slate-800/60 hover:bg-slate-800 text-slate-300 hover:text-white transition relative"
              >
                <Bell className="w-4 h-4" />
                <span className="absolute top-1.5 right-1.5 w-2 h-2 rounded-full bg-emerald-500 ring-2 ring-[#0B0F17]" />
              </button>
            </div>

            {/* User Account / Profile / Auth CTA */}
            {user ? (
              <div className="relative">
                <button
                  onClick={() => setIsProfileOpen(!isProfileOpen)}
                  className="flex items-center space-x-2 p-1.5 pr-3 rounded-xl bg-slate-800/80 hover:bg-slate-800 border border-slate-700/80 transition"
                >
                  <div className="w-7 h-7 rounded-lg bg-gradient-to-tr from-indigo-500 to-violet-500 flex items-center justify-center text-xs font-bold text-white shadow-sm">
                    {user.full_name.charAt(0)}
                  </div>
                  <div className="text-left hidden sm:block">
                    <span className="text-xs font-medium text-slate-200 block leading-tight truncate max-w-[100px]">
                      {user.full_name}
                    </span>
                    <span className="text-[10px] text-indigo-400 font-mono block">
                      {user.role === "CUSTOMER" ? "Fan" : user.role === "ADMIN" ? "Admin" : "Organiser"}
                    </span>
                  </div>
                  <ChevronDown className="w-3 h-3 text-slate-400" />
                </button>

                {isProfileOpen && (
                  <div className="absolute right-0 mt-2 w-64 bg-[#131A26] border border-slate-700 rounded-2xl shadow-2xl py-2 z-50 animate-in fade-in zoom-in-95">
                    <div className="px-4 py-2 border-b border-slate-800">
                      <p className="text-xs font-semibold text-slate-200">{user.full_name}</p>
                      <p className="text-[11px] text-slate-400 truncate">{user.email}</p>
                      <div className="mt-1.5 flex items-center">
                        <span className="text-[10px] px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 font-mono font-medium border border-indigo-500/30">
                          {user.role} ACCOUNT
                        </span>
                      </div>
                    </div>

                    <div className="py-1">
                      <Link
                        href="/my-bookings"
                        onClick={() => setIsProfileOpen(false)}
                        className="flex items-center px-4 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white transition"
                      >
                        <CalendarCheck className="w-4 h-4 mr-2.5 text-indigo-400" />
                        My Tickets & Passes
                      </Link>

                      <Link
                        href="/organiser"
                        onClick={() => setIsProfileOpen(false)}
                        className="flex items-center px-4 py-2 text-xs text-slate-300 hover:bg-slate-800 hover:text-white transition"
                      >
                        <Building2 className="w-4 h-4 mr-2.5 text-violet-400" />
                        Organiser Dashboard
                      </Link>

                      <Link
                        href="/admin"
                        onClick={() => setIsProfileOpen(false)}
                        className="flex items-center px-4 py-2 text-xs text-amber-300 hover:bg-slate-800 hover:text-amber-200 transition"
                      >
                        <ShieldCheck className="w-4 h-4 mr-2.5 text-amber-400" />
                        Admin Venue Manager
                      </Link>

                      {/* Quick Role Switcher */}
                      <div className="px-4 py-2 border-t border-slate-800/80 my-1">
                        <span className="text-[10px] text-slate-500 font-semibold uppercase tracking-wider block mb-1">
                          Account Mode
                        </span>
                        <div className="grid grid-cols-3 gap-1 bg-slate-900 p-1 rounded-lg">
                          <button
                            onClick={() => {
                              switchRole("CUSTOMER");
                              setIsProfileOpen(false);
                            }}
                            className={`py-1 rounded text-[10px] font-medium transition ${
                              user.role === "CUSTOMER"
                                ? "bg-indigo-600 text-white shadow"
                                : "text-slate-400 hover:text-slate-200"
                            }`}
                          >
                            Fan
                          </button>
                          <button
                            onClick={() => {
                              switchRole("ORGANISER");
                              setIsProfileOpen(false);
                            }}
                            className={`py-1 rounded text-[10px] font-medium transition ${
                              user.role === "ORGANISER"
                                ? "bg-violet-600 text-white shadow"
                                : "text-slate-400 hover:text-slate-200"
                            }`}
                          >
                            Organiser
                          </button>
                          <button
                            onClick={() => {
                              switchRole("ADMIN");
                              setIsProfileOpen(false);
                            }}
                            className={`py-1 rounded text-[10px] font-medium transition ${
                              user.role === "ADMIN"
                                ? "bg-amber-600 text-white shadow"
                                : "text-slate-400 hover:text-slate-200"
                            }`}
                          >
                            Admin
                          </button>
                        </div>
                      </div>

                      <button
                        onClick={() => {
                          logout();
                          setIsProfileOpen(false);
                        }}
                        className="w-full flex items-center px-4 py-2 text-xs text-rose-400 hover:bg-rose-500/10 transition border-t border-slate-800/80"
                      >
                        <LogOut className="w-4 h-4 mr-2.5" />
                        Sign Out
                      </button>
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <div className="flex items-center space-x-2.5">
                <Link
                  href="/login"
                  className="px-3.5 py-1.5 rounded-xl text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800/80 border border-slate-700/60 transition"
                >
                  Sign In
                </Link>
                <Link
                  href="/signup"
                  className="px-4 py-1.5 rounded-xl text-xs font-semibold text-white bg-gradient-to-r from-indigo-600 to-violet-600 hover:from-indigo-500 hover:to-violet-500 shadow-md shadow-indigo-500/20 transition flex items-center"
                >
                  <Sparkles className="w-3.5 h-3.5 mr-1.5" />
                  Get Started
                </Link>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* Global ⌘K Search Modal */}
      <SearchModal isOpen={isSearchOpen} onClose={() => setIsSearchOpen(false)} />
    </>
  );
}
