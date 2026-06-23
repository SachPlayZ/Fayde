"use client";
import { motion, useReducedMotion } from "motion/react";

const STATS = [
  { value: "5", label: "task views", sub: "list, kanban, calendar, gantt, weekly" },
  { value: "0", label: "context switches", sub: "everything lives in one tab" },
  { value: "∞", label: "yours forever", sub: "your data, your export, no lock-in" },
];

export function ManifestoSection() {
  const reduce = useReducedMotion();

  return (
    <section className="px-6 md:px-12 lg:px-24 py-32 border-t border-white/[0.05]">
      <div className="max-w-[1400px] mx-auto">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-px bg-white/[0.05]">
          {STATS.map((stat, i) => (
            <motion.div
              key={stat.label}
              initial={reduce ? false : { opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, amount: 0.3 }}
              transition={{
                duration: 0.7,
                delay: i * 0.1,
                ease: [0.16, 1, 0.3, 1],
              }}
              className="bg-[#0a0a0a] px-10 py-14"
            >
              <p className="font-mono text-[5rem] md:text-[6rem] font-bold text-white leading-none tracking-tighter mb-4">
                {stat.value}
              </p>
              <p className="text-zinc-200 text-lg font-semibold mb-2">{stat.label}</p>
              <p className="text-zinc-600 text-sm">{stat.sub}</p>
            </motion.div>
          ))}
        </div>

        <motion.div
          initial={reduce ? false : { opacity: 0, y: 40 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.2 }}
          transition={{ duration: 0.8, delay: 0.2, ease: [0.16, 1, 0.3, 1] }}
          className="mt-24 max-w-[700px]"
        >
          <p className="text-3xl md:text-4xl text-white font-semibold leading-[1.3] tracking-tight">
            Most productivity apps are built for teams.
          </p>
          <p className="mt-6 text-zinc-500 text-lg leading-relaxed">
            Fayde is built for one person: you. No shared workspaces, no noise from others,
            no permission levels. Just a quiet, powerful tool that does exactly what you need.
          </p>
        </motion.div>
      </div>
    </section>
  );
}
