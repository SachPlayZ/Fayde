"use client";
import {
  CheckCircle,
  Calendar,
  Target,
  FolderOpen,
  Zap,
  FileText,
  Clock,
  Timer,
  Bell,
  Workflow,
  Search,
  BookOpen,
  Award,
  BarChart3,
  Layers,
} from "lucide-react";

const ITEMS = [
  { name: "Tasks", icon: CheckCircle },
  { name: "Habits", icon: Calendar },
  { name: "Goals", icon: Target },
  { name: "Projects", icon: FolderOpen },
  { name: "Sprints", icon: Zap },
  { name: "Docs", icon: FileText },
  { name: "Time Tracking", icon: Clock },
  { name: "Pomodoro", icon: Timer },
  { name: "Reminders", icon: Bell },
  { name: "Automations", icon: Workflow },
  { name: "Search", icon: Search },
  { name: "Notes", icon: BookOpen },
  { name: "Streaks", icon: Award },
  { name: "OKRs", icon: BarChart3 },
  { name: "Templates", icon: Layers },
];

export function MarqueeStrip() {
  const doubled = [...ITEMS, ...ITEMS];

  return (
    <section className="relative border-y border-white/[0.04] py-8 bg-[#080808] overflow-hidden select-none">
      {/* Soft edge blur vignettes */}
      <div className="absolute inset-y-0 left-0 w-36 bg-gradient-to-r from-[#0a0a0a] to-transparent pointer-events-none z-10" />
      <div className="absolute inset-y-0 right-0 w-36 bg-gradient-to-l from-[#0a0a0a] to-transparent pointer-events-none z-10" />

      <div
        className="flex gap-6 w-max hover:[animation-play-state:paused]"
        style={{ animation: "marquee-scroll 45s linear infinite" }}
      >
        {doubled.map((item, i) => {
          const Icon = item.icon;
          return (
            <div
              key={i}
              className="flex items-center gap-3 px-6 py-3.5 rounded-full bg-white/[0.02] hover:bg-white/[0.06] border border-white/[0.05] hover:border-white/[0.12] transition-all duration-300 group cursor-default"
            >
              <Icon className="size-4 text-zinc-500 group-hover:text-white transition-colors duration-300" />
              <span className="font-mono text-xs uppercase tracking-[0.18em] text-zinc-400 group-hover:text-white transition-colors duration-300">
                {item.name}
              </span>
            </div>
          );
        })}
      </div>
    </section>
  );
}
