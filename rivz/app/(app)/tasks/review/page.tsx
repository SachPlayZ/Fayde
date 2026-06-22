import { Suspense } from "react";
import { DailyReviewClient } from "./_components/DailyReviewClient";

export default function DailyReviewPage() {
  return (
    <Suspense fallback={<div className="h-48 animate-pulse rounded-xl bg-muted" />}>
      <DailyReviewClient />
    </Suspense>
  );
}
