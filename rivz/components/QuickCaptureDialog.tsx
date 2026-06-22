"use client";
import { useState, useRef, useEffect } from "react";
import { format, addDays, addWeeks, startOfDay } from "date-fns";
import { useQuickCapture } from "@/lib/quick-capture-context";
import { useCreateTask } from "@/lib/tasks-hooks";
import { parseNLDate, formatNLHint } from "@/lib/nldate";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { X, Zap } from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

const DATE_CHIPS = [
  { label: "Today",     date: () => startOfDay(new Date()) },
  { label: "Tomorrow",  date: () => addDays(startOfDay(new Date()), 1) },
  { label: "Next week", date: () => addWeeks(startOfDay(new Date()), 1) },
];

export function QuickCaptureDialog() {
  const { open, closeCapture } = useQuickCapture();
  const createTask = useCreateTask();
  const titleRef = useRef<HTMLInputElement>(null);

  const [title, setTitle] = useState("");
  const [nlInput, setNlInput] = useState("");
  const [parsedDate, setParsedDate] = useState<Date | null>(null);
  const [activeChip, setActiveChip] = useState<number | null>(null);

  // Focus title on open
  useEffect(() => {
    if (open) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setTitle("");
      setNlInput("");
      setParsedDate(null);
      setActiveChip(null);
      setTimeout(() => titleRef.current?.focus(), 50);
    }
  }, [open]);

  const handleNlChange = (val: string) => {
    setNlInput(val);
    setActiveChip(null);
    const parsed = parseNLDate(val);
    setParsedDate(parsed);
  };

  const selectChip = (idx: number) => {
    const date = DATE_CHIPS[idx].date();
    setParsedDate(date);
    setNlInput(DATE_CHIPS[idx].label.toLowerCase());
    setActiveChip(idx);
  };

  const handleSubmit = async (e?: React.FormEvent) => {
    e?.preventDefault();
    if (!title.trim()) return;
    try {
      await createTask.mutateAsync({
        title: title.trim(),
        due_date: parsedDate ? `${format(parsedDate, "yyyy-MM-dd")}T00:00:00Z` : null,
        status: "todo",
        priority: "medium",
      });
      toast.success("Task created");
      closeCapture();
    } catch {
      toast.error("Failed to create task");
    }
  };

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-[60] flex items-start justify-center pt-24 px-4"
      onClick={closeCapture}
    >
      <div
        className="w-full max-w-md bg-background border border-border rounded-2xl shadow-2xl overflow-hidden animate-in fade-in-0 zoom-in-95 duration-150"
        onClick={(e) => e.stopPropagation()}
      >
        <form onSubmit={handleSubmit}>
          {/* Header */}
          <div className="flex items-center gap-2 px-4 pt-4 pb-2">
            <Zap className="size-4 text-amber-500 shrink-0" />
            <p className="text-sm font-semibold text-foreground">Quick capture</p>
            <button
              type="button"
              onClick={closeCapture}
              className="ml-auto text-muted-foreground hover:text-foreground transition-colors"
            >
              <X className="size-4" />
            </button>
          </div>

          {/* Title */}
          <div className="px-4 pb-3">
            <Input
              ref={titleRef}
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="What needs to be done?"
              className="h-10 text-base border-0 border-b rounded-none px-0 focus-visible:ring-0 focus-visible:border-foreground/30 bg-transparent"
              onKeyDown={(e) => {
                if (e.key === "Escape") closeCapture();
              }}
            />
          </div>

          {/* Date section */}
          <div className="px-4 pb-4 flex flex-col gap-2">
            {/* Quick chips */}
            <div className="flex gap-1.5 flex-wrap">
              {DATE_CHIPS.map((chip, i) => (
                <button
                  key={chip.label}
                  type="button"
                  onClick={() => selectChip(i)}
                  className={cn(
                    "px-2.5 py-1 rounded-full text-xs font-medium border transition-all",
                    activeChip === i
                      ? "bg-primary text-primary-foreground border-primary"
                      : "border-border text-muted-foreground hover:border-foreground/40 hover:text-foreground"
                  )}
                >
                  {chip.label}
                </button>
              ))}
            </div>

            {/* NL date input */}
            <div className="relative">
              <Input
                value={nlInput}
                onChange={(e) => handleNlChange(e.target.value)}
                placeholder="or type: next friday, in 3 days…"
                className="h-8 text-xs text-muted-foreground pr-20"
                onKeyDown={(e) => {
                  if (e.key === "Enter") { e.preventDefault(); handleSubmit(); }
                  if (e.key === "Escape") closeCapture();
                }}
              />
              {parsedDate && nlInput && (
                <span className="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-emerald-600 dark:text-emerald-400 font-medium pointer-events-none">
                  {formatNLHint(parsedDate)}
                </span>
              )}
            </div>
          </div>

          {/* Footer */}
          <div className="px-4 pb-4 flex items-center justify-between">
            <p className="text-[10px] text-muted-foreground">Esc to close · Enter to create</p>
            <Button
              type="submit"
              size="sm"
              disabled={!title.trim() || createTask.isPending}
              className="h-8"
            >
              {createTask.isPending ? "Creating…" : "Create task"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
