"use client";
import { useState } from "react";
import Link from "next/link";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { leetcodeProblemSchema, type LeetcodeProblemInput } from "@/lib/schemas";
import {
  useProblems,
  useCreateProblem,
  useDeleteProblem,
  useLeetcodeStats,
  type Problem,
} from "@/lib/leetcode-hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import { Code2, Plus, Trash2, Brain, CheckCircle2, Clock, ExternalLink } from "lucide-react";
import { toast } from "sonner";
import { ApiError } from "@/lib/api";

const TOPICS = [
  "array", "string", "hash-table", "dynamic-programming", "math", "sorting",
  "greedy", "depth-first-search", "binary-search", "breadth-first-search",
  "tree", "matrix", "binary-tree", "two-pointers", "bit-manipulation",
  "stack", "heap", "graph", "prefix-sum", "simulation", "backtracking",
  "sliding-window", "union-find", "monotonic-stack", "trie", "linked-list",
  "recursion", "divide-and-conquer",
];

const difficultyStyle: Record<string, string> = {
  easy: "bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/20",
  medium: "bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/20",
  hard: "bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/20",
};

const stateStyle: Record<string, string> = {
  new: "bg-muted text-muted-foreground",
  learning: "bg-blue-500/15 text-blue-700 dark:text-blue-400",
  review: "bg-emerald-500/15 text-emerald-700 dark:text-emerald-400",
  relearning: "bg-rose-500/15 text-rose-700 dark:text-rose-400",
};

function AddProblemDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (o: boolean) => void }) {
  const create = useCreateProblem();
  const [selectedTopics, setSelectedTopics] = useState<string[]>([]);
  const {
    register,
    handleSubmit,
    reset,
    setValue,
    control,
    formState: { errors, isSubmitting },
  } = useForm<LeetcodeProblemInput>({
    resolver: zodResolver(leetcodeProblemSchema),
    defaultValues: { title: "", difficulty: "medium", topics: [], notes: "" },
  });
  const difficultyValue = useWatch({ control, name: "difficulty" });

  const close = () => {
    reset();
    setSelectedTopics([]);
    onOpenChange(false);
  };

  const onSubmit = async (data: LeetcodeProblemInput) => {
    try {
      await create.mutateAsync({
        title: data.title,
        lc_number: data.lc_number || null,
        slug: data.slug || null,
        url: data.url || null,
        difficulty: data.difficulty,
        topics: selectedTopics,
        notes: data.notes || "",
      });
      toast.success("Problem added");
      close();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Something went wrong");
    }
  };

  return (
    <Dialog open={open} onOpenChange={(o) => (o ? onOpenChange(true) : close())}>
      <DialogContent className="sm:max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Add problem</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
          <div className="grid grid-cols-[1fr_100px] gap-3">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="title">Title <span className="text-rose-500">*</span></Label>
              <Input id="title" placeholder="Two Sum" aria-invalid={!!errors.title} {...register("title")} />
              {errors.title && <p className="text-xs text-rose-500">{errors.title.message}</p>}
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="lc_number">#</Label>
              <Input id="lc_number" type="number" placeholder="1" {...register("lc_number", { valueAsNumber: true })} />
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="url">LeetCode URL</Label>
            <Input id="url" placeholder="https://leetcode.com/problems/two-sum/" {...register("url")} />
            {errors.url && <p className="text-xs text-rose-500">{errors.url.message}</p>}
          </div>

          <div className="flex flex-col gap-1.5">
            <Label>Difficulty</Label>
            <Select value={difficultyValue} onValueChange={(v) => setValue("difficulty", v as LeetcodeProblemInput["difficulty"])}>
              <SelectTrigger className="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="easy">Easy</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="hard">Hard</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label>Topics</Label>
            <div className="flex flex-wrap gap-1.5 max-h-32 overflow-y-auto rounded-lg border border-border p-2">
              {TOPICS.map((t) => {
                const active = selectedTopics.includes(t);
                return (
                  <button
                    key={t}
                    type="button"
                    onClick={() =>
                      setSelectedTopics((prev) =>
                        active ? prev.filter((x) => x !== t) : [...prev, t]
                      )
                    }
                    className={cn(
                      "px-2 py-0.5 rounded-full text-[11px] font-medium border transition-colors",
                      active
                        ? "bg-primary text-primary-foreground border-primary"
                        : "border-border hover:border-primary/50 hover:bg-muted"
                    )}
                  >
                    {t}
                  </button>
                );
              })}
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="notes">Notes / approach</Label>
            <Textarea id="notes" rows={3} placeholder="Approach, complexity, gotchas…" {...register("notes")} />
          </div>

          <DialogFooter className="mt-2 gap-2">
            <Button type="button" variant="outline" onClick={close}>Cancel</Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Adding…" : "Add problem"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export default function LeetcodePage() {
  const [difficulty, setDifficulty] = useState<string>("");
  const [status, setStatus] = useState<"" | "due" | "overdue">("");
  const [topic, setTopic] = useState<string>("");
  const [dialogOpen, setDialogOpen] = useState(false);

  const { data: problems, isLoading } = useProblems({ difficulty, status, topic });
  const { data: stats } = useLeetcodeStats();
  const del = useDeleteProblem();

  return (
    <div className="flex flex-col gap-6 max-w-5xl">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold tracking-tight">DSA Grind</h2>
          <p className="text-sm text-muted-foreground mt-0.5">
            Track problems and let FSRS bring them back before you forget.
          </p>
        </div>
        <Button onClick={() => setDialogOpen(true)}>
          <Plus className="size-4 mr-1" /> Add Problem
        </Button>
      </div>

      {stats && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <StatCard label="Problems" value={stats.total_problems} icon={Code2} />
          <StatCard label="Solved" value={stats.total_solved} icon={CheckCircle2} />
          <StatCard label="Due today" value={stats.due_today_count} icon={Clock} highlight={stats.due_today_count > 0} />
          <StatCard label="Streak" value={stats.review_streak} icon={Brain} suffix=" days" />
        </div>
      )}

      {stats && stats.due_today_count > 0 && (
        <Link
          href="/leetcode/review"
          className="flex items-center justify-between rounded-xl border border-primary/30 bg-primary/5 px-4 py-3 text-sm font-medium hover:bg-primary/10 transition-colors"
        >
          <span className="flex items-center gap-2">
            <Brain className="size-4 text-primary" />
            {stats.due_today_count} problem{stats.due_today_count === 1 ? "" : "s"} ready for review
          </span>
          <span className="text-primary">Start review →</span>
        </Link>
      )}

      <div className="flex flex-wrap items-center gap-3">
        <Tabs value={status} onValueChange={(v) => setStatus(v as "" | "due" | "overdue")}>
          <TabsList>
            <TabsTrigger value="">All</TabsTrigger>
            <TabsTrigger value="due">Due Today</TabsTrigger>
            <TabsTrigger value="overdue">Overdue</TabsTrigger>
          </TabsList>
        </Tabs>

        <Select value={difficulty || "all"} onValueChange={(v) => setDifficulty(v === "all" ? "" : v)}>
          <SelectTrigger className="w-36"><SelectValue placeholder="Difficulty" /></SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">All difficulties</SelectItem>
              <SelectItem value="easy">Easy</SelectItem>
              <SelectItem value="medium">Medium</SelectItem>
              <SelectItem value="hard">Hard</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>

        <Select value={topic || "all"} onValueChange={(v) => setTopic(v === "all" ? "" : v)}>
          <SelectTrigger className="w-44"><SelectValue placeholder="Topic" /></SelectTrigger>
          <SelectContent className="max-h-64">
            <SelectGroup>
              <SelectItem value="all">All topics</SelectItem>
              {TOPICS.map((t) => (
                <SelectItem key={t} value={t}>{t}</SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-12 rounded-lg border border-border bg-card animate-pulse" />
          ))}
        </div>
      ) : (problems?.length ?? 0) === 0 ? (
        <div className="flex flex-col items-center justify-center gap-3 py-16 rounded-xl border border-border bg-card text-muted-foreground">
          <Code2 className="size-8" />
          <p className="text-sm">No problems match. Add one to get started.</p>
        </div>
      ) : (
        <div className="rounded-xl border border-border bg-card overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-14">#</TableHead>
                <TableHead>Title</TableHead>
                <TableHead>Difficulty</TableHead>
                <TableHead>Topics</TableHead>
                <TableHead>State</TableHead>
                <TableHead>Due</TableHead>
                <TableHead className="w-10" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {problems!.map((p) => (
                <ProblemRow key={p.id} problem={p} onDelete={() => del.mutate(p.id)} />
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      <AddProblemDialog open={dialogOpen} onOpenChange={setDialogOpen} />
    </div>
  );
}

function StatCard({
  label, value, icon: Icon, suffix, highlight,
}: {
  label: string; value: number; icon: React.ElementType; suffix?: string; highlight?: boolean;
}) {
  return (
    <div className={cn("rounded-xl border border-border bg-card p-3 flex items-center gap-3", highlight && "border-primary/40 bg-primary/5")}>
      <Icon className={cn("size-5", highlight ? "text-primary" : "text-muted-foreground")} />
      <div>
        <div className="text-lg font-bold leading-none">{value}{suffix}</div>
        <div className="text-xs text-muted-foreground mt-0.5">{label}</div>
      </div>
    </div>
  );
}

function ProblemRow({ problem, onDelete }: { problem: Problem; onDelete: () => void }) {
  const due = problem.card && new Date(problem.card.due_date) <= new Date();
  return (
    <TableRow className="group">
      <TableCell className="text-muted-foreground">{problem.lc_number ?? "—"}</TableCell>
      <TableCell>
        <Link href={`/leetcode/${problem.id}`} className="font-medium hover:text-primary hover:underline">
          {problem.title}
        </Link>
      </TableCell>
      <TableCell>
        <Badge variant="outline" className={cn("capitalize", difficultyStyle[problem.difficulty])}>
          {problem.difficulty}
        </Badge>
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-1 max-w-56">
          {problem.topics.slice(0, 3).map((t) => (
            <span key={t} className="text-[10px] px-1.5 py-0.5 rounded-full bg-muted text-muted-foreground">{t}</span>
          ))}
          {problem.topics.length > 3 && (
            <span className="text-[10px] px-1.5 py-0.5 text-muted-foreground">+{problem.topics.length - 3}</span>
          )}
        </div>
      </TableCell>
      <TableCell>
        {problem.card && (
          <Badge variant="outline" className={cn("capitalize border-transparent", stateStyle[problem.card.card_state])}>
            {problem.card.card_state}
          </Badge>
        )}
      </TableCell>
      <TableCell className={cn("text-xs", due && "text-primary font-medium")}>
        {problem.card ? new Date(problem.card.due_date).toLocaleDateString() : "—"}
      </TableCell>
      <TableCell>
        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          {problem.url && (
            <a href={problem.url} target="_blank" rel="noopener noreferrer" className="text-muted-foreground hover:text-foreground">
              <ExternalLink className="size-3.5" />
            </a>
          )}
          <button onClick={onDelete} className="text-muted-foreground hover:text-rose-500">
            <Trash2 className="size-3.5" />
          </button>
        </div>
      </TableCell>
    </TableRow>
  );
}
