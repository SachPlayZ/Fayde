"use client";
import { useEffect } from "react";
import { motion, useMotionValue, useSpring, useReducedMotion } from "motion/react";

export function CursorFollower() {
  const reduce = useReducedMotion();
  const mx = useMotionValue(-100);
  const my = useMotionValue(-100);
  const x = useSpring(mx, { stiffness: 130, damping: 14 });
  const y = useSpring(my, { stiffness: 130, damping: 14 });

  useEffect(() => {
    if (reduce) return;
    const move = (e: MouseEvent) => {
      mx.set(e.clientX - 16);
      my.set(e.clientY - 16);
    };
    window.addEventListener("mousemove", move);
    return () => window.removeEventListener("mousemove", move);
  }, [mx, my, reduce]);

  if (reduce) return null;

  return (
    <motion.div
      className="fixed top-0 left-0 w-8 h-8 rounded-full border border-white/20 pointer-events-none z-[100] mix-blend-difference hidden md:block"
      style={{ x, y }}
    />
  );
}
