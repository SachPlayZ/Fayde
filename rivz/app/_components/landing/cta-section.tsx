"use client";
import Link from "next/link";
import { motion, useReducedMotion } from "motion/react";

export function CtaSection() {
  const reduce = useReducedMotion();

  return (
    <>
      <section className="px-6 md:px-12 lg:px-24 py-40 border-t border-white/[0.05]">
        <div className="max-w-[1400px] mx-auto">
          <motion.div
            initial={reduce ? false : { opacity: 0, y: 50 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, amount: 0.3 }}
            transition={{ duration: 0.9, ease: [0.16, 1, 0.3, 1] }}
            className="relative"
          >
            <p className="font-mono text-zinc-600 text-[11px] uppercase tracking-[0.24em] mb-8">
              Ready when you are
            </p>
            <h2 className="text-[clamp(3rem,8vw,8rem)] font-bold tracking-tight text-white leading-[0.92] mb-12 max-w-[12ch]">
              Your most productive year starts now.
            </h2>
            <Link
              href="/signup"
              className="inline-block bg-white text-black font-semibold px-10 py-4 rounded-full hover:bg-zinc-100 transition-all duration-200 active:scale-[0.97] text-base"
            >
              Start free
            </Link>
          </motion.div>
        </div>
      </section>

      <footer className="border-t border-white/[0.05] px-6 md:px-12 lg:px-24 py-12">
        <div className="max-w-[1400px] mx-auto flex flex-col md:flex-row items-start md:items-center justify-between gap-6">
          <p className="text-white font-bold text-lg tracking-tight">Fayde</p>
          <div className="flex items-center gap-8">
            <Link href="/login" className="text-zinc-600 hover:text-zinc-400 text-sm transition-colors duration-200">
              Sign in
            </Link>
            <Link href="/signup" className="text-zinc-600 hover:text-zinc-400 text-sm transition-colors duration-200">
              Sign up
            </Link>
          </div>
          <p className="text-zinc-700 text-xs font-mono">
            Personal productivity. No noise.
          </p>
        </div>
      </footer>
    </>
  );
}
