import { Suspense } from "react";
import { ProblemDetailClient } from "./_components/ProblemDetailClient";
import { Skeleton } from "@/components/ui/skeleton";

function ProblemDetailSkeleton() {
  return (
    <div className="flex flex-col gap-6 max-w-4xl mx-auto">
      <div className="flex items-center justify-between">
        <Skeleton className="h-8 w-28" />
      </div>
      <Skeleton className="h-24 w-full rounded-xl" />
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_280px] gap-6">
        <Skeleton className="h-40 w-full rounded-xl" />
        <Skeleton className="h-64 w-full rounded-xl" />
      </div>
    </div>
  );
}

export async function generateStaticParams() {
  return [{ id: "placeholder" }];
}

export default async function ProblemDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return (
    <Suspense fallback={<ProblemDetailSkeleton />}>
      <ProblemDetailClient id={id} />
    </Suspense>
  );
}
