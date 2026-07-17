"use client";
import { useState } from "react";
import Link from "next/link";
import { useProblem, useUpdateProblem } from "@/lib/leetcode-hooks";
import { Button, buttonVariants } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Textarea } from "@/components/ui/textarea";
import {
  Table, TableHeader, TableBody, TableRow, TableHead, TableCell,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import { ArrowLeft, ExternalLink, Pencil, Check, Brain, Flame, Repeat, XCircle } from "lucide-react";
import { toast } from "sonner";

const difficultyStyle: Record<string, string> = {
  easy: "bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/20",
  medium: "bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/20",
  hard: "bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/20",
};

const ratingLabel: Record<number, { label: string; className: string }> = {
  1: { label: "Again", className: "text-rose-500" },
  2: { label: "Hard", className: "text-amber-500" },
  3: { label: "Good", className: "text-emerald-500" },
  4: { label: "Easy", className: "text-blue-500" },
};

export function ProblemDetailClient({ id }: { id: string }) {
  const { data: problem, isLoading } = useProblem(id);
  const update = useUpdateProblem();
  const [editingNotes, setEditingNotes] = useState(false);
  const [notesValue, setNotesValue] = useState("");

  if (isLoading || !problem) {
    return <p className="text-sm text-muted-foreground">Loading…</p>;
  }

  const due = problem.card && new Date(problem.card.due_date) <= new Date();

  const startEditingNotes = () => {
    setNotesValue(problem.notes);
    setEditingNotes(true);
  };

  const saveNotes = async () => {
    await update.mutateAsync({ id: problem.id, patch: { notes: notesValue } });
    setEditingNotes(false);
    toast.success("Notes saved");
  };

  return (
    <div className="flex flex-col gap-6 max-w-4xl mx-auto">
      <Link href="/leetcode" className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground w-fit">
        <ArrowLeft className="size-3.5" /> DSA Grind
      </Link>

      <div className="flex flex-col gap-2">
        <div className="flex items-center gap-2 flex-wrap">
          {problem.lc_number && <span className="text-muted-foreground text-sm">#{problem.lc_number}</span>}
          <h1 className="text-xl font-bold">{problem.title}</h1>
          <Badge variant="outline" className={cn("capitalize", difficultyStyle[problem.difficulty])}>
            {problem.difficulty}
          </Badge>
          {problem.url && (
            <a href={problem.url} target="_blank" rel="noopener noreferrer" className="text-muted-foreground hover:text-primary">
              <ExternalLink className="size-4" />
            </a>
          )}
        </div>
        {problem.topics.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {problem.topics.map((t) => (
              <span key={t} className="text-[11px] px-2 py-0.5 rounded-full bg-muted text-muted-foreground">{t}</span>
            ))}
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-[1fr_280px] gap-6">
        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold">Notes</h3>
            {!editingNotes && (
              <Button variant="ghost" size="sm" className="h-7 text-xs" onClick={startEditingNotes}>
                <Pencil className="size-3 mr-1" /> Edit
              </Button>
            )}
          </div>
          {editingNotes ? (
            <div className="flex flex-col gap-2">
              <Textarea
                value={notesValue}
                onChange={(e) => setNotesValue(e.target.value)}
                rows={6}
                placeholder="Approach, complexity, gotchas…"
              />
              <div className="flex gap-2 self-end">
                <Button variant="outline" size="sm" onClick={() => setEditingNotes(false)}>Cancel</Button>
                <Button size="sm" onClick={saveNotes} disabled={update.isPending}>
                  <Check className="size-3.5 mr-1" /> Save
                </Button>
              </div>
            </div>
          ) : (
            <div className="rounded-xl border border-border bg-card p-4 text-sm text-muted-foreground min-h-20 whitespace-pre-wrap">
              {problem.notes || "No notes yet."}
            </div>
          )}

          <h3 className="text-sm font-semibold mt-4">Review history</h3>
          {problem.reviews.length === 0 ? (
            <p className="text-xs text-muted-foreground">No reviews yet.</p>
          ) : (
            <div className="rounded-xl border border-border bg-card overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Date</TableHead>
                    <TableHead>Rating</TableHead>
                    <TableHead>Interval</TableHead>
                    <TableHead>Stability</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {problem.reviews.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell className="text-xs">{new Date(r.reviewed_at).toLocaleDateString()}</TableCell>
                      <TableCell className={cn("text-xs font-medium", ratingLabel[r.rating]?.className)}>
                        {ratingLabel[r.rating]?.label ?? r.rating}
                      </TableCell>
                      <TableCell className="text-xs">{r.scheduled_days}d</TableCell>
                      <TableCell className="text-xs">{r.stability.toFixed(1)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </div>

        <div className="flex flex-col gap-3">
          <div className="rounded-xl border border-border bg-card p-4 flex flex-col gap-3">
            <h3 className="text-sm font-semibold flex items-center gap-1.5">
              <Brain className="size-4 text-primary" /> FSRS Card
            </h3>
            {problem.card ? (
              <div className="flex flex-col gap-2 text-xs">
                <Stat label="Next review" value={new Date(problem.card.due_date).toLocaleString()} highlight={!!due} />
                <Stat label="State" value={problem.card.card_state} />
                <Stat label="Stability" value={`${problem.card.stability.toFixed(1)} days`} />
                <Stat label="Difficulty" value={problem.card.difficulty.toFixed(1)} />
                <Stat label="Reps" value={String(problem.card.reps)} icon={<Repeat className="size-3" />} />
                <Stat label="Lapses" value={String(problem.card.lapses)} icon={<XCircle className="size-3" />} />
              </div>
            ) : (
              <p className="text-xs text-muted-foreground">No card yet.</p>
            )}
            {due && (
              <Link href="/leetcode/review" className={cn(buttonVariants({ size: "sm" }), "w-full")}>
                <Flame className="size-3.5 mr-1" /> Review Now
              </Link>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function Stat({ label, value, icon, highlight }: { label: string; value: string; icon?: React.ReactNode; highlight?: boolean }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground flex items-center gap-1">{icon}{label}</span>
      <span className={cn("font-medium capitalize", highlight && "text-primary")}>{value}</span>
    </div>
  );
}
