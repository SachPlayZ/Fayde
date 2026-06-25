"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { CursorFollower } from "./_components/landing/cursor-follower";
import { LandingNav } from "./_components/landing/landing-nav";
import { HeroSection } from "./_components/landing/hero-section";
import { MarqueeStrip } from "./_components/landing/marquee-strip";
import { FeaturesSection } from "./_components/landing/features-section";
import { ManifestoSection } from "./_components/landing/manifesto-section";
import { CtaSection } from "./_components/landing/cta-section";

export default function HomePage() {
  const { user, loading } = useAuth();
  const router = useRouter();
  const [isTauri, setIsTauri] = useState(false);

  useEffect(() => {
    const checkTauri =
      typeof window !== "undefined" &&
      (window as unknown as { __TAURI_INTERNALS__?: unknown }).__TAURI_INTERNALS__ !== undefined;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setIsTauri(checkTauri);
  }, []);

  useEffect(() => {
    if (loading) return;

    if (user) {
      router.replace("/tasks");
    } else if (isTauri) {
      router.replace("/login");
    }
  }, [user, loading, isTauri, router]);

  // While loading authentication state, or if we are redirecting, render a clean loading skeleton
  if (loading || user || (isTauri && !user)) {
    return (
      <main className="bg-[#0a0a0a] min-h-screen flex items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <div className="size-7 rounded-full border-2 border-white/20 border-t-white animate-spin" />
        </div>
      </main>
    );
  }

  // Web app guest (unauthenticated) - show landing page
  return (
    <main className="bg-[#0a0a0a] text-white min-h-screen overflow-x-hidden selection:bg-white/20">
      <CursorFollower />
      <LandingNav />
      <HeroSection />
      <MarqueeStrip />
      <FeaturesSection />
      <ManifestoSection />
      <CtaSection />
    </main>
  );
}
