import { CursorFollower } from "./_components/landing/cursor-follower";
import { LandingNav } from "./_components/landing/landing-nav";
import { HeroSection } from "./_components/landing/hero-section";
import { MarqueeStrip } from "./_components/landing/marquee-strip";
import { FeaturesSection } from "./_components/landing/features-section";
import { ManifestoSection } from "./_components/landing/manifesto-section";
import { CtaSection } from "./_components/landing/cta-section";

export default function HomePage() {
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
