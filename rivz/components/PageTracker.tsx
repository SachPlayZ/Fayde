"use client";
import { useEffect } from "react";
import { usePathname } from "next/navigation";
import { trackPageView } from "@/lib/tracker";

export function PageTracker({ userId }: { userId?: string | null }) {
  const pathname = usePathname();

  useEffect(() => {
    trackPageView(pathname, userId);
  }, [pathname, userId]);

  return null;
}
