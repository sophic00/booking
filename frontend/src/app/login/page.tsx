"use client";

import React, { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Ticket,
  Lock,
  Mail,
  User,
  Phone,
  Eye,
  EyeOff,
  ArrowLeft,
  ShieldCheck,
  Zap,
  RefreshCw,
  QrCode,
  Sparkles,
  CheckCircle2,
  Building2,
} from "lucide-react";
import { useAuth } from "../../context/AuthContext";
import { UserRole } from "../../lib/types";

export default function AuthPage() {
  const router = useRouter();
  const { login, register } = useAuth();

  const [mode, setMode] = useState<"LOGIN" | "SIGNUP">("LOGIN");
  const [role, setRole] = useState<UserRole>("CUSTOMER");
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  // Form states
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [fullName, setFullName] = useState("");
  const [phone, setPhone] = useState("");
  const [agreeTerms, setAgreeTerms] = useState(true);

  // Password strength calculation
  const getPasswordStrength = (pass: string) => {
    if (!pass) return 0;
    let score = 0;
    if (pass.length >= 8) score += 25;
    if (/[A-Z]/.test(pass)) score += 25;
    if (/[0-9]/.test(pass)) score += 25;
    if (/[^A-Za-z0-9]/.test(pass)) score += 25;
    return score;
  };

  const strength = getPasswordStrength(password);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage("");
    setLoading(true);

    try {
      if (mode === "LOGIN") {
        await login(email, password);
      } else {
        if (!agreeTerms) {
          throw new Error("Please agree to terms and reservation policies");
        }
        await register({
          email,
          password,
          full_name: fullName,
          phone,
          role,
        });
      }

      // Redirect based on role
      if (role === "ORGANISER") {
        router.push("/organiser");
      } else {
        router.push("/my-bookings");
      }
    } catch (err: any) {
      setErrorMessage(err.message || "Authentication failed. Please check your credentials.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[calc(100vh-5rem)] flex items-center justify-center px-4 sm:px-6 lg:px-8 py-10">
      <div className="w-full max-w-5xl glass-panel rounded-3xl border border-slate-800/90 shadow-2xl overflow-hidden bg-[#131A26]/90 backdrop-blur-2xl grid grid-cols-1 lg:grid-cols-12">
        {/* ========================================================
            LEFT PANEL: AUTHENTICATION FORM
            ======================================================== */}
        <div className="lg:col-span-7 p-6 sm:p-10 flex flex-col justify-between space-y-6">
          <div className="space-y-5">
            {/* Top Navigation & Breadcrumb */}
            <div className="flex items-center justify-between">
              <Link
                href="/"
                className="inline-flex items-center text-xs font-semibold text-slate-400 hover:text-indigo-300 transition"
              >
                <ArrowLeft className="w-3.5 h-3.5 mr-1.5" />
                Back to Browse Events
              </Link>
              <div className="flex items-center space-x-1.5 text-xs text-slate-500 font-mono">
                <ShieldCheck className="w-4 h-4 text-emerald-400" />
                <span>256-bit SSL</span>
              </div>
            </div>

            {/* Header / Title */}
            <div>
              <div className="flex items-center space-x-3 mb-2">
                <div className="w-8 h-8 rounded-lg bg-gradient-to-tr from-indigo-600 to-violet-500 flex items-center justify-center shadow-md">
                  <Ticket className="w-4 h-4 text-white" />
                </div>
                <span className="text-sm font-black tracking-wider text-slate-300 uppercase">
                  VELVET<span className="text-indigo-400">SEATS</span> AUTH
                </span>
              </div>
              <h1 className="text-2xl sm:text-3xl font-black text-white tracking-tight">
                {mode === "LOGIN" ? "Welcome to VelvetSeats" : "Create Your VelvetSeats Account"}
              </h1>
              <p className="text-xs text-slate-400 mt-1">
                {mode === "LOGIN"
                  ? "Sign in to access your held seats, waitlist priority, and live QR tickets."
                  : "Join hundreds of thousands of fans and venue organisers."}
              </p>
            </div>

            {/* Role Switcher Segmented Control */}
            <div className="p-1 rounded-xl bg-slate-900/90 border border-slate-800 grid grid-cols-2 gap-1 text-xs font-semibold">
              <button
                type="button"
                onClick={() => setRole("CUSTOMER")}
                className={`py-2 rounded-lg flex items-center justify-center space-x-2 transition ${
                  role === "CUSTOMER"
                    ? "bg-indigo-600 text-white shadow-md shadow-indigo-600/30"
                    : "text-slate-400 hover:text-slate-200"
                }`}
              >
                <Ticket className="w-3.5 h-3.5" />
                <span>🎟️ Fan / Customer</span>
              </button>
              <button
                type="button"
                onClick={() => setRole("ORGANISER")}
                className={`py-2 rounded-lg flex items-center justify-center space-x-2 transition ${
                  role === "ORGANISER"
                    ? "bg-violet-600 text-white shadow-md shadow-violet-600/30"
                    : "text-slate-400 hover:text-slate-200"
                }`}
              >
                <Building2 className="w-3.5 h-3.5" />
                <span>🏢 Event Organiser</span>
              </button>
            </div>

            {/* Tabs: Sign In vs Create Account */}
            <div className="flex border-b border-slate-800 text-xs font-bold">
              <button
                type="button"
                onClick={() => {
                  setMode("LOGIN");
                  setErrorMessage("");
                }}
                className={`pb-3 pr-4 transition ${
                  mode === "LOGIN"
                    ? "text-indigo-400 border-b-2 border-indigo-500"
                    : "text-slate-400 hover:text-slate-200"
                }`}
              >
                Sign In
              </button>
              <button
                type="button"
                onClick={() => {
                  setMode("SIGNUP");
                  setErrorMessage("");
                }}
                className={`pb-3 px-4 transition ${
                  mode === "SIGNUP"
                    ? "text-indigo-400 border-b-2 border-indigo-500"
                    : "text-slate-400 hover:text-slate-200"
                }`}
              >
                Create Account
              </button>
            </div>

            {/* Social Logins */}
            <div className="grid grid-cols-2 gap-3">
              <button
                type="button"
                onClick={() => {
                  setEmail("alex.rivera@example.com");
                  setPassword("Password123!");
                }}
                className="flex items-center justify-center space-x-2 py-2.5 px-3 rounded-xl bg-slate-900/80 hover:bg-slate-800 border border-slate-700/80 text-xs font-semibold text-slate-200 transition"
              >
                <svg className="w-4 h-4" viewBox="0 0 24 24">
                  <path
                    fill="currentColor"
                    d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                  />
                  <path
                    fill="#34A853"
                    d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                  />
                  <path
                    fill="#FBBC05"
                    d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z"
                  />
                  <path
                    fill="#EA4335"
                    d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"
                  />
                </svg>
                <span>Google SSO</span>
              </button>

              <button
                type="button"
                onClick={() => {
                  setEmail("demo.organiser@velvetseats.com");
                  setPassword("Password123!");
                  setRole("ORGANISER");
                }}
                className="flex items-center justify-center space-x-2 py-2.5 px-3 rounded-xl bg-slate-900/80 hover:bg-slate-800 border border-slate-700/80 text-xs font-semibold text-slate-200 transition"
              >
                <svg className="w-4 h-4 fill-current" viewBox="0 0 24 24">
                  <path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M15.97 6.85c.62-.75 1.04-1.8 0.93-2.85-.9.04-1.98.6-2.63 1.35-.57.65-1.07 1.72-.94 2.74 1 .08 2.02-.49 2.64-1.24z" />
                </svg>
                <span>Apple ID</span>
              </button>
            </div>

            <div className="relative flex items-center justify-center">
              <div className="border-t border-slate-800 w-full" />
              <span className="bg-[#131A26] px-3 text-[11px] text-slate-500 font-mono uppercase">
                or with email
              </span>
            </div>

            {/* Error Banner */}
            {errorMessage && (
              <div className="p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs flex items-center space-x-2">
                <span>⚠️ {errorMessage}</span>
              </div>
            )}

            {/* Form Fields */}
            <form onSubmit={handleSubmit} className="space-y-4">
              {mode === "SIGNUP" && (
                <div>
                  <label className="text-[11px] font-semibold text-slate-400 block mb-1">
                    Full Name
                  </label>
                  <div className="relative">
                    <User className="w-4 h-4 text-slate-500 absolute left-3 top-3" />
                    <input
                      type="text"
                      required
                      value={fullName}
                      onChange={(e) => setFullName(e.target.value)}
                      placeholder="Alex Rivera"
                      className="w-full bg-[#0B0F17] border border-slate-700/80 rounded-xl pl-9 pr-3 py-2.5 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
                    />
                  </div>
                </div>
              )}

              <div>
                <label className="text-[11px] font-semibold text-slate-400 block mb-1">
                  Email Address
                </label>
                <div className="relative">
                  <Mail className="w-4 h-4 text-slate-500 absolute left-3 top-3" />
                  <input
                    type="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="name@example.com"
                    className="w-full bg-[#0B0F17] border border-slate-700/80 rounded-xl pl-9 pr-3 py-2.5 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
                  />
                </div>
              </div>

              {mode === "SIGNUP" && (
                <div>
                  <label className="text-[11px] font-semibold text-slate-400 block mb-1">
                    Mobile Phone (For Instant QR Ticket SMS Alerts)
                  </label>
                  <div className="relative">
                    <Phone className="w-4 h-4 text-slate-500 absolute left-3 top-3" />
                    <input
                      type="tel"
                      value={phone}
                      onChange={(e) => setPhone(e.target.value)}
                      placeholder="+1 (555) 000-0000"
                      className="w-full bg-[#0B0F17] border border-slate-700/80 rounded-xl pl-9 pr-3 py-2.5 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
                    />
                  </div>
                </div>
              )}

              <div>
                <div className="flex items-center justify-between mb-1">
                  <label className="text-[11px] font-semibold text-slate-400 block">
                    Password
                  </label>
                  {mode === "LOGIN" && (
                    <button
                      type="button"
                      onClick={() => alert("Password reset link sent to registered email")}
                      className="text-[11px] text-indigo-400 hover:text-indigo-300 transition"
                    >
                      Forgot password?
                    </button>
                  )}
                </div>
                <div className="relative">
                  <Lock className="w-4 h-4 text-slate-500 absolute left-3 top-3" />
                  <input
                    type={showPassword ? "text" : "password"}
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="••••••••"
                    className="w-full bg-[#0B0F17] border border-slate-700/80 rounded-xl pl-9 pr-10 py-2.5 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500 transition"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-3 text-slate-500 hover:text-slate-300"
                  >
                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>

                {/* Password Strength Meter */}
                {mode === "SIGNUP" && password && (
                  <div className="mt-2 space-y-1">
                    <div className="h-1.5 w-full bg-slate-800 rounded-full overflow-hidden flex">
                      <div
                        className={`h-full transition-all duration-300 ${
                          strength <= 25
                            ? "w-1/4 bg-rose-500"
                            : strength <= 50
                            ? "w-2/4 bg-amber-500"
                            : strength <= 75
                            ? "w-3/4 bg-indigo-500"
                            : "w-full bg-emerald-500"
                        }`}
                      />
                    </div>
                    <div className="flex justify-between text-[10px] text-slate-400 font-mono">
                      <span>Security: {strength < 50 ? "Weak" : strength < 100 ? "Good" : "Strong"}</span>
                      <span>Min 8 chars, 1 uppercase, 1 symbol</span>
                    </div>
                  </div>
                )}
              </div>

              {mode === "SIGNUP" && (
                <div className="flex items-start space-x-2 pt-1">
                  <input
                    type="checkbox"
                    id="terms"
                    checked={agreeTerms}
                    onChange={(e) => setAgreeTerms(e.target.checked)}
                    className="mt-0.5 rounded border-slate-700 bg-[#0B0F17] text-indigo-600 focus:ring-indigo-500"
                  />
                  <label htmlFor="terms" className="text-[11px] text-slate-400 leading-tight">
                    I agree to the <span className="text-slate-200">Terms of Service</span>, <span className="text-slate-200">10-Minute Hold Policy</span>, and automated waitlist reallocation terms.
                  </label>
                </div>
              )}

              {/* Submit CTA Button */}
              <button
                type="submit"
                disabled={loading}
                className="w-full py-3 rounded-xl font-bold text-xs sm:text-sm text-white bg-gradient-to-r from-indigo-600 via-indigo-500 to-violet-600 hover:from-indigo-500 hover:to-violet-500 shadow-xl shadow-indigo-600/30 transition flex items-center justify-center disabled:opacity-50"
              >
                {loading ? (
                  <RefreshCw className="w-4 h-4 animate-spin" />
                ) : mode === "LOGIN" ? (
                  `Sign In as ${role === "CUSTOMER" ? "Fan" : "Organiser"}`
                ) : (
                  `Create ${role === "CUSTOMER" ? "Fan" : "Organiser"} Account`
                )}
              </button>
            </form>
          </div>

          {/* Bottom Security Note */}
          <div className="pt-4 border-t border-slate-800 text-[11px] text-slate-500 text-center">
            Zero-conflict seat reservations powered by PostgreSQL atomic transactional locking.
          </div>
        </div>

        {/* ========================================================
            RIGHT PANEL: PLATFORM PERKS & LIVE QR TICKET SHOWCASE
            ======================================================== */}
        <div className="lg:col-span-5 p-8 lg:p-10 bg-gradient-to-b from-indigo-950/40 via-[#0E1522] to-[#0B0F17] border-t lg:border-t-0 lg:border-l border-slate-800 flex flex-col justify-between space-y-6">
          <div className="space-y-6">
            <div className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-indigo-500/20 text-indigo-300 border border-indigo-500/40">
              <Sparkles className="w-3.5 h-3.5 mr-1.5" />
              LIVE TICKETING ADVANTAGE
            </div>

            <div>
              <h2 className="text-xl sm:text-2xl font-black text-white tracking-tight">
                The Smartest Way to Book & Attend Live Shows
              </h2>
              <p className="text-xs text-slate-400 mt-1.5 leading-relaxed">
                Experience guaranteed anti-scalp seat holds, automated waitlists, and instant verified mobile passes.
              </p>
            </div>

            {/* Visual Mockup Card 1: Live QR Ticket */}
            <div className="glass-panel rounded-2xl p-4 border border-indigo-500/30 bg-slate-900/80 shadow-xl space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <div className="w-6 h-6 rounded bg-indigo-600 flex items-center justify-center text-white">
                    <Ticket className="w-3.5 h-3.5" />
                  </div>
                  <span className="text-xs font-bold text-slate-200">Hans Zimmer Live 2026</span>
                </div>
                <span className="px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-300 text-[10px] font-mono font-bold flex items-center">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 mr-1.5 animate-pulse" />
                  Confirmed & Active
                </span>
              </div>

              <div className="grid grid-cols-3 gap-2 bg-[#0B0F17] p-2.5 rounded-xl text-center text-xs font-mono">
                <div>
                  <span className="text-[10px] text-slate-500 block">TIER</span>
                  <span className="text-indigo-400 font-bold">VIP Diamond</span>
                </div>
                <div>
                  <span className="text-[10px] text-slate-500 block">SEAT</span>
                  <span className="text-slate-100 font-bold">Row A - 14</span>
                </div>
                <div>
                  <span className="text-[10px] text-slate-500 block">REF</span>
                  <span className="text-slate-300 font-bold">#TB-7AF2</span>
                </div>
              </div>

              <div className="flex items-center justify-between text-[11px] text-slate-400 pt-1">
                <span className="flex items-center">
                  <QrCode className="w-3.5 h-3.5 mr-1.5 text-indigo-400" />
                  Encrypted Scannable QR Pass
                </span>
                <span className="text-emerald-400 font-bold">Valid for Entry</span>
              </div>
            </div>

            {/* Visual Mockup Card 2: 10-Min Hold Guarantee */}
            <div className="glass-panel rounded-2xl p-3.5 border border-slate-800 bg-slate-900/60 flex items-start space-x-3">
              <div className="w-8 h-8 rounded-xl bg-violet-500/20 text-violet-400 flex items-center justify-center shrink-0">
                <Zap className="w-4 h-4" />
              </div>
              <div className="text-xs space-y-0.5">
                <h4 className="font-bold text-slate-200">10-Minute Fair-Hold Guarantee</h4>
                <p className="text-slate-400 text-[11px] leading-relaxed">
                  Seats are locked exclusively while you check out. Zero risk of double booking or race conditions.
                </p>
              </div>
            </div>

            {/* Visual Mockup Card 3: Auto-Waitlist Reallocation */}
            <div className="glass-panel rounded-2xl p-3.5 border border-slate-800 bg-slate-900/60 flex items-start space-x-3">
              <div className="w-8 h-8 rounded-xl bg-emerald-500/20 text-emerald-400 flex items-center justify-center shrink-0">
                <RefreshCw className="w-4 h-4" />
              </div>
              <div className="text-xs space-y-0.5">
                <h4 className="font-bold text-slate-200">Automated Waitlist Priority Link</h4>
                <p className="text-slate-400 text-[11px] leading-relaxed">
                  When cancellations occur, seats are offered immediately to waitlisted fans with time-limited direct claim links.
                </p>
              </div>
            </div>
          </div>

          {/* Social Proof Counter */}
          <div className="pt-4 border-t border-slate-800/80 flex items-center justify-between text-xs text-slate-400">
            <div>
              <span className="text-base font-black text-white block">500,000+</span>
              <span className="text-[10px] text-slate-500 font-mono">Reserved Tickets</span>
            </div>
            <div className="text-right">
              <span className="text-base font-black text-emerald-400 block">0</span>
              <span className="text-[10px] text-slate-500 font-mono">Double Bookings</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
