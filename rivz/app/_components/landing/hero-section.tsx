"use client";
import Link from "next/link";
import { motion, useReducedMotion } from "motion/react";

const words = [
  { text: "Think", muted: false },
  { text: "less.", muted: true },
  { text: "Do", muted: false },
  { text: "more.", muted: false },
];

export function HeroSection() {
  const reduce = useReducedMotion();

  return (
    <section
      className="relative min-h-[100dvh] flex flex-col justify-center px-6 md:px-12 lg:px-24 pt-16 overflow-hidden"
      style={{
        backgroundImage:
          "radial-gradient(circle, rgba(255,255,255,0.035) 1px, transparent 1px)",
        backgroundSize: "28px 28px",
      }}
    >
      <div className="absolute bottom-0 left-0 right-0 h-40 bg-gradient-to-t from-[#0a0a0a] to-transparent pointer-events-none" />
      <div className="absolute right-0 top-0 bottom-0 w-1/3 bg-gradient-to-l from-[#0a0a0a]/60 to-transparent pointer-events-none hidden lg:block" />

      <div className="relative max-w-[1400px] mx-auto w-full">
        <motion.p
          initial={reduce ? false : { opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
          className="font-mono text-zinc-600 text-[11px] uppercase tracking-[0.24em] mb-10"
        >
          Personal productivity suite
        </motion.p>

        <h1 className="text-[clamp(3.8rem,10.5vw,10.5rem)] font-bold leading-[0.9] tracking-[-0.04em] mb-12">
          {words.map((word, i) => (
            <motion.span
              key={i}
              className={`block ${word.muted ? "text-zinc-600" : "text-white"}`}
              initial={reduce ? false : { opacity: 0, y: 70, filter: "blur(12px)" }}
              animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
              transition={{
                duration: 0.9,
                delay: 0.08 + i * 0.14,
                ease: [0.16, 1, 0.3, 1],
              }}
            >
              {word.text}
            </motion.span>
          ))}
        </h1>

        <div className="flex flex-col gap-8 md:flex-row md:items-center md:gap-12">
          <motion.p
            initial={reduce ? false : { opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, delay: 0.7, ease: [0.16, 1, 0.3, 1] }}
            className="text-zinc-400 text-lg leading-relaxed max-w-[40ch]"
          >
            Tasks, habits, goals, projects, and docs. All yours. Zero friction.
          </motion.p>

          <motion.div
            initial={reduce ? false : { opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, delay: 0.85, ease: [0.16, 1, 0.3, 1] }}
            className="flex items-center gap-5 flex-shrink-0"
          >
            <Link
              href="/signup"
              className="bg-white text-black font-semibold px-7 py-3.5 rounded-full hover:bg-zinc-100 transition-all duration-200 active:scale-[0.97] text-sm whitespace-nowrap"
            >
              Start free
            </Link>
            <Link
              href="/login"
              className="text-zinc-500 hover:text-white text-sm transition-colors duration-200 whitespace-nowrap"
            >
              Sign in
            </Link>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
