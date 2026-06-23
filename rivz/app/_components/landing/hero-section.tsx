"use client";
import Link from "next/link";
import Image from "next/image";
import { motion, useReducedMotion, useAnimate } from "motion/react";
import { useEffect } from "react";

const words = [
  { text: "Think", muted: false },
  { text: "less.", muted: true },
  { text: "Do", muted: false },
  { text: "more.", muted: false },
];

const RING_COUNT = 4;
const RING_DURATION = 4.5;

export function HeroSection() {
  const reduce = useReducedMotion();
  const [scope, animate] = useAnimate();

  useEffect(() => {
    if (!scope.current) return;
    let live = true;

    async function run() {
      if (reduce) {
        animate(scope.current, { opacity: 1, y: 0 }, { duration: 0 });
        if (live) {
          animate(scope.current, { y: [-6, 6, -6] }, {
            duration: 6, repeat: Infinity, ease: [0.45, 0, 0.55, 1],
          });
        }
        return;
      }

      // Phase 1: flat on ground (diagonal) → rise up like a 3D badge
      await animate(scope.current, {
        rotateX: [76, 8, 0],
        rotateY: [12, 3, 0],
        rotateZ: [-20, -4, 0],
        y: [110, -14, 0],
        opacity: [0, 1, 1],
        scale: [0.72, 1.07, 1],
      }, {
        duration: 1.7,
        ease: [0.16, 1, 0.3, 1],
      });

      // Phase 2: gentle float loop
      if (live) {
        animate(scope.current, { y: [-8, 8, -8] }, {
          duration: 6, repeat: Infinity, ease: [0.45, 0, 0.55, 1],
        });
      }
    }

    run();
    return () => { live = false; };
  }, [reduce]); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <section
      className="relative min-h-[100dvh] flex flex-col justify-center px-6 md:px-12 lg:px-24 pt-20 pb-16 overflow-hidden"
      style={{
        backgroundImage:
          "radial-gradient(circle, rgba(255,255,255,0.035) 1px, transparent 1px)",
        backgroundSize: "28px 28px",
      }}
    >
      <div className="absolute bottom-0 left-0 right-0 h-40 bg-gradient-to-t from-[#0a0a0a] to-transparent pointer-events-none" />
      <div className="absolute right-0 top-0 bottom-0 w-1/3 bg-gradient-to-l from-[#0a0a0a]/60 to-transparent pointer-events-none hidden lg:block" />

      <div className="relative max-w-[1400px] mx-auto w-full flex flex-col lg:flex-row items-center justify-between gap-16 lg:gap-8 z-10">
        {/* Left Column: Heading and CTA */}
        <div className="flex-1 min-w-0 w-full">
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

        {/* Right Column: Logo */}
        <div
          className="relative flex-shrink-0 w-full max-w-[380px] sm:max-w-[450px] aspect-square flex items-center justify-center select-none overflow-visible"
          style={{ perspective: "1200px" }}
        >
          {/* Layered ambient glow */}
          <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
            <motion.div
              animate={{ scale: [1, 1.18, 1], opacity: [0.7, 1, 0.7] }}
              transition={{ duration: 7, repeat: Infinity, ease: [0.45, 0, 0.55, 1] }}
              className="absolute size-[200px] rounded-full"
              style={{ background: "radial-gradient(circle, rgba(255,255,255,0.08) 0%, rgba(255,255,255,0.03) 45%, transparent 70%)" }}
            />
            <motion.div
              animate={{ scale: [1.12, 0.94, 1.12], opacity: [0.35, 0.6, 0.35] }}
              transition={{ duration: 10, repeat: Infinity, ease: [0.45, 0, 0.55, 1], delay: 2 }}
              className="absolute size-[360px] rounded-full"
              style={{ background: "radial-gradient(circle, rgba(255,255,255,0.04) 0%, transparent 60%)" }}
            />
          </div>

          {/* Concentric ring pulses */}
          {!reduce &&
            Array.from({ length: RING_COUNT }).map((_, i) => (
              <div key={i} className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <motion.div
                  className="rounded-full"
                  style={{ width: 160, height: 160, border: "1px solid rgba(255,255,255,0.45)" }}
                  initial={{ scale: 0.35, opacity: 0 }}
                  animate={{ scale: [0.35, 2.4], opacity: [0, 0.22, 0.14, 0] }}
                  transition={{
                    duration: RING_DURATION,
                    repeat: Infinity,
                    delay: i * (RING_DURATION / RING_COUNT),
                    ease: [0.25, 0.46, 0.45, 0.94],
                  }}
                />
              </div>
            ))}

          {/* Outer orbital */}
          {!reduce && (
            <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
              <motion.div
                animate={{ rotate: 360 }}
                transition={{ duration: 22, repeat: Infinity, ease: "linear" }}
                className="rounded-full"
                style={{
                  width: 290,
                  height: 290,
                  border: "1px solid transparent",
                  borderTopColor: "rgba(255,255,255,0.09)",
                  borderRightColor: "rgba(255,255,255,0.04)",
                }}
              />
            </div>
          )}

          {/* Inner orbital */}
          {!reduce && (
            <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
              <motion.div
                animate={{ rotate: -360 }}
                transition={{ duration: 34, repeat: Infinity, ease: "linear" }}
                className="rounded-full"
                style={{
                  width: 230,
                  height: 230,
                  border: "1px solid transparent",
                  borderBottomColor: "rgba(255,255,255,0.07)",
                  borderLeftColor: "rgba(255,255,255,0.03)",
                }}
              />
            </div>
          )}

          {/* Logo — 3D badge rising from ground */}
          <motion.div
            ref={scope}
            style={{ opacity: 0 }}
            className="relative z-10 size-44 sm:size-56 rounded-[3rem] bg-zinc-950/80 border border-white/10 p-6 sm:p-8 shadow-[0_0_50px_rgba(0,0,0,0.8),_0_0_30px_rgba(255,255,255,0.03)] flex items-center justify-center backdrop-blur-xl"
          >
            <Image
              src="/logo.png"
              alt="Fayde Logo"
              width={224}
              height={224}
              className="size-full object-contain rounded-[2.2rem] select-none"
              priority
            />
          </motion.div>
        </div>
      </div>
    </section>
  );
}
