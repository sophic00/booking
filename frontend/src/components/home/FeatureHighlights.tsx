import React from "react";
import { Grid3X3, Timer, RefreshCw, QrCode, ArrowRight, ShieldCheck, Sparkles } from "lucide-react";
import Link from "next/link";

export function FeatureHighlights() {
  const features = [
    {
      icon: Grid3X3,
      tag: "Exact Seating",
      title: "Interactive Visual Seat Map",
      description:
        "Select your exact seat from an interactive stage and screen layout. Choose between VIP, Premium, and Standard tiers with live availability.",
      color: "from-indigo-500/20 to-indigo-600/5",
      borderColor: "border-indigo-500/30",
      iconColor: "text-indigo-400",
    },
    {
      icon: Timer,
      tag: "Stress-Free Booking",
      title: "10-Minute Cart Hold",
      description:
        "Take your time at checkout. Your selected seats are held exclusively in your cart for 10 minutes so you never lose them while paying.",
      color: "from-violet-500/20 to-violet-600/5",
      borderColor: "border-violet-500/30",
      iconColor: "text-violet-400",
    },
    {
      icon: RefreshCw,
      tag: "Never Miss a Show",
      title: "Smart Waitlist Alerts",
      description:
        "Sold-out show? Join category waitlists to get priority notifications with a direct booking link as soon as seats open up.",
      color: "from-emerald-500/20 to-emerald-600/5",
      borderColor: "border-emerald-500/30",
      iconColor: "text-emerald-400",
    },
    {
      icon: QrCode,
      tag: "Contactless Entry",
      title: "Digital Mobile Passes",
      description:
        "Instant scannable mobile QR passes delivered directly to your account and email for smooth admission at the venue turnstiles.",
      color: "from-amber-500/20 to-amber-600/5",
      borderColor: "border-amber-500/30",
      iconColor: "text-amber-400",
    },
  ];

  return (
    <section className="py-16 border-t border-slate-800/80">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Section Header */}
        <div className="text-center max-w-2xl mx-auto mb-12 space-y-2">
          <span className="text-xs font-bold uppercase tracking-widest text-indigo-400 font-mono flex items-center justify-center">
            <Sparkles className="w-3.5 h-3.5 mr-1.5" />
            THE ULTIMATE LIVE EXPERIENCE
          </span>
          <h2 className="text-2xl sm:text-3xl font-black text-white tracking-tight">
            Why Fans Choose VelvetSeats
          </h2>
          <p className="text-slate-400 text-xs sm:text-sm">
            From seamless seat selection to instant mobile passes, we make discovering and attending live entertainment effortless.
          </p>
        </div>

        {/* Feature Cards Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {features.map((f, idx) => {
            const Icon = f.icon;
            return (
              <div
                key={idx}
                className={`glass-panel rounded-2xl p-6 border ${f.borderColor} bg-gradient-to-b ${f.color} flex flex-col justify-between space-y-4 hover:translate-y-[-2px] transition duration-300`}
              >
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="w-10 h-10 rounded-xl bg-slate-900/80 border border-slate-700/60 flex items-center justify-center">
                      <Icon className={`w-5 h-5 ${f.iconColor}`} />
                    </div>
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded-full bg-slate-900/80 text-slate-300 border border-slate-700/60 font-semibold">
                      {f.tag}
                    </span>
                  </div>

                  <h3 className="text-base font-bold text-slate-100">{f.title}</h3>
                  <p className="text-xs text-slate-400 leading-relaxed">{f.description}</p>
                </div>

                <div className="pt-2 text-[11px] font-semibold text-slate-300 flex items-center group cursor-pointer">
                  <span>Learn more</span>
                  <ArrowRight className="w-3.5 h-3.5 ml-1 text-slate-400 group-hover:translate-x-1 transition" />
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
