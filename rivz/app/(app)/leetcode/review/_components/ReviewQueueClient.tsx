"use client";
import { useState, useEffect } from "react";
import Link from "next/link";
import { useReviewQueue, useReviewProblem, useLeetcodeStats } from "@/lib/leetcode-hooks";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { ArrowLeft, ExternalLink, Brain, Flame, PartyPopper } from "lucide-react";
import { toast } from "sonner";

const difficultyStyle: Record<string, string> = {
  easy: "bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/20",
  medium: "bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/20",
  hard: "bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/20",
};

const ratingButtons = [
  { rating: 1 as const, label: "Again", hint: "Forgot", className: "border-rose-500/40 text-rose-600 dark:text-rose-400 hover:bg-rose-500/10" },
  { rating: 2 as const, label: "Hard", hint: "Barely", className: "border-amber-500/40 text-amber-600 dark:text-amber-400 hover:bg-amber-500/10" },
  { rating: 3 as const, label: "Good", hint: "Recalled", className: "border-emerald-500/40 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-500/10" },
  { rating: 4 as const, label: "Easy", hint: "Trivial", className: "border-blue-500/40 text-blue-600 dark:text-blue-400 hover:bg-blue-500/10" },
];

export function ReviewQueueClient() {
  const { data: queue, isLoading } = useReviewQueue();
  const review = useReviewProblem();
  const { data: stats } = useLeetcodeStats();
  const [queueIndex, setQueueIndex] = useState(0);
  const [snapshot, setSnapshot] = useState<typeof queue>(undefined);

  // Freeze the queue on first load so re-fetches (triggered by review mutations)
  // don't reshuffle the card the user is currently looking at.
  useEffect(() => {
    if (queue && snapshot === undefined) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setSnapshot(queue);
    }
  }, [queue, snapshot]);

  if (isLoading || snapshot === undefined) {
    return <p className="text-sm text-muted-foreground text-center py-16">Loading queue…</p>;
  }

  const current = snapshot[queueIndex];

  const handleRate = async (rating: 1 | 2 | 3 | 4) => {
    if (!current) return;
    try {
      await review.mutateAsync({ id: current.id, rating });
      setQueueIndex((i) => i + 1);
    } catch {
      toast.error("Failed to submit review");
    }
  };

  if (!current) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-20 max-w-md mx-auto text-center">
        <PartyPopper className="size-10 text-primary" />
        <h2 className="text-lg font-bold">Queue cleared!</h2>
        {stats && (
          <p className="text-sm text-muted-foreground flex items-center gap-1.5">
            <Flame className="size-4 text-orange-500" /> Review streak: {stats.review_streak} day{stats.review_streak === 1 ? "" : "s"}
          </p>
        )}
        <p className="text-sm text-muted-foreground">Come back later for the next batch.</p>
        <Link href="/leetcode" className="text-sm text-primary hover:underline mt-2">
          Back to Problems
        </Link>
      </div>
    );
  }

  const progress = ((queueIndex) / snapshot.length) * 100;

  return (
    <div className="flex flex-col gap-6 max-w-xl mx-auto">
      <Link href="/leetcode" className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground w-fit">
        <ArrowLeft className="size-3.5" /> DSA Grind
      </Link>

      <div className="flex flex-col gap-1.5">
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span>Problem {queueIndex + 1} of {snapshot.length}</span>
          <span>{Math.round(progress)}%</span>
        </div>
        <div className="w-full h-1.5 rounded-full bg-muted overflow-hidden">
          <div className="h-full bg-primary rounded-full transition-all" style={{ width: `${progress}%` }} />
        </div>
      </div>

      <div className="rounded-xl border border-border bg-card p-6 flex flex-col gap-4">
        <div className="flex items-center gap-2">
          <Badge variant="outline" className={cn("capitalize", difficultyStyle[current.difficulty])}>
            {current.difficulty}
          </Badge>
          {current.card && (
            <Badge variant="outline" className="capitalize">{current.card.card_state}</Badge>
          )}
        </div>

        <div>
          <h2 className="text-lg font-bold">
            {current.lc_number && <span className="text-muted-foreground">#{current.lc_number} </span>}
            {current.title}
          </h2>
          {current.topics.length > 0 && (
            <div className="flex flex-wrap gap-1.5 mt-2">
              {current.topics.map((t) => (
                <span key={t} className="text-[11px] px-2 py-0.5 rounded-full bg-muted text-muted-foreground">{t}</span>
              ))}
            </div>
          )}
        </div>

        {current.url && (
          <a
            href={current.url}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 text-sm text-primary hover:underline w-fit"
          >
            Open on LeetCode <ExternalLink className="size-3.5" />
          </a>
        )}

        <p className="text-xs text-muted-foreground flex items-center gap-1.5">
          <Brain className="size-3.5" /> Solve it, then rate how it went:
        </p>

        <div className="grid grid-cols-4 gap-2">
          {ratingButtons.map((b) => (
            <button
              key={b.rating}
              type="button"
              disabled={review.isPending}
              onClick={() => handleRate(b.rating)}
              className={cn(
                "flex flex-col items-center gap-0.5 rounded-lg border-2 py-3 text-xs font-semibold transition-colors disabled:opacity-50",
                b.className
              )}
            >
              {b.label}
              <span className="text-[10px] font-normal opacity-70">{b.hint}</span>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
