"use client";
import { useState } from "react";
import { format, parseISO } from "date-fns";
import {
  useGoals,
  useCreateGoal,
  useUpdateGoal,
  useDeleteGoal,
  useAddKR,
  useUpdateKR,
  useDeleteKR,
  type Goal,
  type KeyResult,
} from "@/lib/goals-hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import { Target, Plus, Trash2, ChevronDown, ChevronRight, Calendar } from "lucide-react";
import { toast } from "sonner";

const STATUS: Record<string, { label: string; color: string }> = {
  on_track: { label: "On track", color: "text-emerald-500 bg-emerald-500/10" },
  at_risk: { label: "At risk", color: "text-amber-500 bg-amber-500/10" },
  off_track: { label: "Off track", color: "text-rose-500 bg-rose-500/10" },
  done: { label: "Done", color: "text-blue-500 bg-blue-500/10" },
  archived: { label: "Archived", color: "text-muted-foreground bg-muted" },
};

function ProgressRing({ value }: { value: number }) {
  const r = 18;
  const c = 2 * Math.PI * r;
  const off = c - (Math.min(value, 100) / 100) * c;
  return (
    <svg width="44" height="44" viewBox="0 0 44 44" className="shrink-0 -rotate-90">
      <circle cx="22" cy="22" r={r} fill="none" stroke="currentColor" strokeWidth="4" className="text-muted" />
      <circle
        cx="22"
        cy="22"
        r={r}
        fill="none"
        stroke="currentColor"
        strokeWidth="4"
        strokeDasharray={c}
        strokeDashoffset={off}
        strokeLinecap="round"
        className="text-primary transition-all"
      />
    </svg>
  );
}

function KRRow({ goalId, kr }: { goalId: string; kr: KeyResult }) {
  const update = useUpdateKR(goalId);
  const del = useDeleteKR(goalId);
  const pct = kr.target_val > 0 ? Math.min((kr.current_val / kr.target_val) * 100, 100) : 0;
  const editable = kr.metric_type !== "task_completion";

  return (
    <div className="flex items-center gap-3 py-2">
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between gap-2">
          <span className="text-sm truncate">{kr.title}</span>
          <span className="text-xs text-muted-foreground tabular-nums shrink-0">
            {kr.current_val}/{kr.target_val}
            {kr.metric_type === "task_completion" && " tasks"}
          </span>
        </div>
        <div className="h-1.5 rounded-full bg-muted mt-1.5 overflow-hidden">
          <div className="h-full bg-primary transition-all" style={{ width: `${pct}%` }} />
        </div>
      </div>
      {editable && (
        <input
          type="number"
          defaultValue={kr.current_val}
          onBlur={(e) => {
            const v = parseFloat(e.target.value);
            if (!isNaN(v) && v !== kr.current_val) update.mutate({ krId: kr.id, patch: { current_val: v } });
          }}
          className="w-16 h-7 rounded-md border border-border bg-background px-2 text-xs"
        />
      )}
      <Trash2
        className="size-3.5 text-muted-foreground hover:text-destructive cursor-pointer shrink-0"
        onClick={() => del.mutate(kr.id)}
      />
    </div>
  );
}

function GoalCard({ goal }: { goal: Goal }) {
  const [open, setOpen] = useState(false);
  const update = useUpdateGoal();
  const del = useDeleteGoal();
  const addKR = useAddKR(goal.id);
  const [krTitle, setKrTitle] = useState("");
  const [krType, setKrType] = useState("percent");
  const st = STATUS[goal.status] ?? STATUS.on_track;

  const handleAddKR = () => {
    if (!krTitle.trim()) return;
    addKR.mutate(
      { title: krTitle.trim(), metric_type: krType },
      { onSuccess: () => setKrTitle("") }
    );
  };

  return (
    <div className="rounded-xl border border-border bg-card overflow-hidden hover:-translate-y-0.5 hover:shadow-md hover:border-border/80 transition-all duration-300">
      <div className="flex items-center gap-4 p-4">
        <ProgressRing value={goal.progress} />
        <button onClick={() => setOpen((v) => !v)} className="flex-1 min-w-0 text-left">
          <div className="flex items-center gap-2">
            {open ? <ChevronDown className="size-4 shrink-0" /> : <ChevronRight className="size-4 shrink-0" />}
            <span className="font-semibold truncate">{goal.title}</span>
          </div>
          <div className="flex items-center gap-2 mt-1 pl-6 text-xs text-muted-foreground">
            <span className="tabular-nums">{Math.round(goal.progress)}%</span>
            {goal.target_date && (
              <span className="inline-flex items-center gap-1">
                <Calendar className="size-3" />
                {format(parseISO(goal.target_date), "MMM d, yyyy")}
              </span>
            )}
          </div>
        </button>
        <Select value={goal.status} onValueChange={(v) => update.mutate({ id: goal.id, patch: { status: v } })}>
          <SelectTrigger className={cn("h-7 w-auto gap-1 border-0 text-xs font-medium", st.color)}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {Object.entries(STATUS).map(([k, v]) => (
              <SelectItem key={k} value={k}>{v.label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Trash2
          className="size-4 text-muted-foreground hover:text-destructive cursor-pointer shrink-0"
          onClick={() => del.mutate(goal.id, { onSuccess: () => toast.success("Goal deleted") })}
        />
      </div>

      {open && (
        <div className="border-t border-border px-4 py-3">
          <p className="text-xs font-medium text-muted-foreground mb-1">Key results</p>
          <div className="divide-y divide-border">
            {goal.key_results.length === 0 && (
              <p className="text-xs text-muted-foreground py-2">No key results yet.</p>
            )}
            {goal.key_results.map((kr) => (
              <KRRow key={kr.id} goalId={goal.id} kr={kr} />
            ))}
          </div>
          <div className="flex items-center gap-2 mt-3">
            <Input
              value={krTitle}
              onChange={(e) => setKrTitle(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleAddKR()}
              placeholder="Add a key result…"
              className="h-8 flex-1"
            />
            <Select value={krType} onValueChange={setKrType}>
              <SelectTrigger className="h-8 w-36 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="percent">Percent</SelectItem>
                <SelectItem value="number">Number</SelectItem>
                <SelectItem value="task_completion">Linked tasks</SelectItem>
              </SelectContent>
            </Select>
            <Button size="sm" className="h-8" onClick={handleAddKR}>
              <Plus className="size-3.5" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

export default function GoalsPage() {
  const { data: goals, isLoading } = useGoals();
  const create = useCreateGoal();
  const [title, setTitle] = useState("");
  const [date, setDate] = useState("");

  const handleCreate = () => {
    if (!title.trim()) return;
    create.mutate(
      { title: title.trim(), target_date: date || undefined },
      {
        onSuccess: () => {
          setTitle("");
          setDate("");
          toast.success("Goal created");
        },
      }
    );
  };

  return (
    <div className="flex flex-col gap-6 max-w-3xl">
      <div>
        <h2 className="text-xl font-bold tracking-tight">Goals</h2>
        <p className="text-sm text-muted-foreground mt-0.5">
          Objectives and key results. Progress rolls up automatically.
        </p>
      </div>

      <div className="flex items-center gap-2 rounded-xl border border-border bg-card p-2">
        <Input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleCreate()}
          placeholder="New goal…"
          className="flex-1 border-0 shadow-none focus-visible:ring-0"
        />
        <Input
          type="date"
          value={date}
          onChange={(e) => setDate(e.target.value)}
          className="w-40 h-9 text-sm"
        />
        <Button onClick={handleCreate} disabled={!title.trim() || create.isPending}>
          <Plus className="size-4 mr-1" /> Add
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-20 rounded-xl border border-border bg-card animate-pulse" />
          ))}
        </div>
      ) : (goals?.length ?? 0) === 0 ? (
        <div className="flex flex-col items-center justify-center gap-3 py-16 rounded-xl border border-border bg-card text-muted-foreground">
          <Target className="size-8" />
          <p className="text-sm">No goals yet. Set your first objective above.</p>
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {goals!.map((g, i) => (
            <div
              key={g.id}
              className="animate-in fade-in-0 slide-in-from-bottom-2 duration-400"
              style={{ animationDelay: `${i * 50}ms`, animationFillMode: "both" }}
            >
              <GoalCard goal={g} />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
