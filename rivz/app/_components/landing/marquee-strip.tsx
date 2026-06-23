const ITEMS = [
  "Tasks",
  "Habits",
  "Goals",
  "Projects",
  "Sprints",
  "Docs",
  "Time Tracking",
  "Pomodoro",
  "Reminders",
  "Automations",
  "Search",
  "Notes",
  "Streaks",
  "OKRs",
  "Templates",
];

export function MarqueeStrip() {
  const doubled = [...ITEMS, ...ITEMS];

  return (
    <section className="border-y border-white/[0.05] py-5 overflow-hidden">
      <div
        className="flex gap-0 w-max"
        style={{ animation: "marquee-scroll 32s linear infinite" }}
      >
        {doubled.map((item, i) => (
          <span
            key={i}
            className="font-mono text-zinc-700 text-xs uppercase tracking-[0.2em] whitespace-nowrap px-8"
          >
            {item}
            <span className="ml-8 text-zinc-800">·</span>
          </span>
        ))}
      </div>
    </section>
  );
}
