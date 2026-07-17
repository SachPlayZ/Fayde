import { Suspense } from "react";
import { ReviewQueueClient } from "./_components/ReviewQueueClient";
import { Skeleton } from "@/components/ui/skeleton";

function ReviewSkeleton() {
  return (
    <div className="flex flex-col gap-4 max-w-xl mx-auto">
      <Skeleton className="h-4 w-full rounded-full" />
      <Skeleton className="h-64 w-full rounded-xl" />
    </div>
  );
}

export default function ReviewPage() {
  return (
    <Suspense fallback={<ReviewSkeleton />}>
      <ReviewQueueClient />
    </Suspense>
  );
}
