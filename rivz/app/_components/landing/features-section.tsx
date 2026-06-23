"use client";
import { motion, useReducedMotion } from "motion/react";
import {
  CheckSquare,
  Flame,
  Target,
  Layers,
  FileText,
} from "lucide-react";

type Feature = {
  icon: React.ComponentType<{ className?: string; strokeWidth?: number }>;
  title: string;
  description: string;
  span?: string;
  bg: string;
};

const features: Feature[] = [
  {
    icon: CheckSquare,
    title: "Tasks & Projects",
    description:
      "List, kanban, calendar, gantt, and weekly views. Subtasks, tags, priorities, due dates, and automations — wired into your workflow.",
    span: "md:col-span-2",
    bg: "bg-zinc-900",
  },
  {
    icon: Flame,
    title: "Habits",
    description:
      "Track daily habits. Build streaks. See what sticks and what doesn't — without judgment.",
    bg: "bg-zinc-900/50",
  },
  {
    icon: Target,
    title: "Goals & OKRs",
    description:
      "Set ambitious goals. Define key results. Stay accountable to yourself — not a team.",
    bg: "bg-zinc-950",
  },
  {
    icon: Layers,
    title: "Sprints",
    description:
      "Time-boxed execution. Batch work into focused sprints and ship with clarity.",
    bg: "bg-zinc-900/50",
  },
  {
    icon: FileText,
    title: "Docs & Notes",
    description:
      "Rich documents linked to tasks and projects. Your thinking lives alongside your doing.",
    bg: "bg-zinc-950",
  },
];

function FeatureCard({
  feature,
  index,
  reduce,
}: {
  feature: Feature;
  index: number;
  reduce: boolean | null;
}) {
  const Icon = feature.icon;
  return (
    <motion.div
      initial={reduce ? false : { opacity: 0, y: 40 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, amount: 0.15 }}
      transition={{
        duration: 0.65,
        delay: index * 0.07,
        ease: [0.16, 1, 0.3, 1],
      }}
      className={`${feature.span ?? ""} ${feature.bg} rounded-2xl p-8 border border-white/[0.06] hover:border-white/[0.14] transition-all duration-300 group cursor-default`}
    >
      <Icon
        className="w-5 h-5 text-zinc-400 mb-8 group-hover:text-white transition-colors duration-300"
        strokeWidth={1.5}
      />
      <h3 className="text-white text-xl font-semibold tracking-tight mb-3">
        {feature.title}
      </h3>
      <p className="text-zinc-500 text-sm leading-relaxed">{feature.description}</p>
    </motion.div>
  );
}

export function FeaturesSection() {
  const reduce = useReducedMotion();

  return (
    <section className="px-6 md:px-12 lg:px-24 py-32">
      <div className="max-w-[1400px] mx-auto">
        <motion.div
          initial={reduce ? false : { opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.2 }}
          transition={{ duration: 0.7, ease: [0.16, 1, 0.3, 1] }}
          className="mb-20"
        >
          <h2 className="text-5xl md:text-6xl font-bold tracking-tight text-white leading-[1] mb-5">
            Everything you need.
            <br />
            <span className="text-zinc-600">Nothing you don&apos;t.</span>
          </h2>
          <p className="text-zinc-500 text-lg max-w-[48ch]">
            A complete suite of personal productivity tools, wired together so nothing falls through the cracks.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {features.map((f, i) => (
            <FeatureCard key={f.title} feature={f} index={i} reduce={reduce} />
          ))}
        </div>
      </div>
    </section>
  );
}
