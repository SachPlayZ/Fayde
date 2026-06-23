"use client";
import { useState } from "react";
import { toast } from "sonner";
import { format } from "date-fns";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogTrigger,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import { Plus, Trash2, CalendarRange, ChevronDown, ChevronUp } from "lucide-react";
import {
  useSprints,
  useCreateSprint,
  useDeleteSprint,
  useSprintTaskIDs,
} from "@/lib/sprints-hooks";
import type { Sprint } from "@/lib/sprints-hooks";

function SprintCard({
  sprint,
  onDelete,
  deleteDisabled,
}: {
  sprint: Sprint;
  onDelete: (id: string) => void;
  deleteDisabled: boolean;
}) {
  const [expanded, setExpanded] = useState(false);
  const { data: taskIDs, isLoading: taskIDsLoading } = useSprintTaskIDs(
    expanded ? sprint.id : ""
  );

  return (
    <div className="rounded-xl border border-border bg-card overflow-hidden hover:-translate-y-0.5 hover:shadow-md hover:border-border/80 transition-all duration-300">
      <div className="flex items-start gap-3 px-4 py-4">
        <div className="flex-1 min-w-0 flex flex-col gap-1.5">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-sm font-semibold">{sprint.name}</span>
            <Badge variant="secondary" className="text-xs">
              {sprint.start_date && sprint.end_date
                ? `${format(new Date(sprint.start_date), "MMM d")} – ${format(
                    new Date(sprint.end_date),
                    "MMM d, yyyy"
                  )}`
                : "No dates"}
            </Badge>
          </div>
          {sprint.goal && (
            <p className="text-xs text-muted-foreground">{sprint.goal}</p>
          )}
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          <Button
            size="sm"
            variant="outline"
            className="gap-1.5 text-xs h-7 px-2.5"
            onClick={() => setExpanded((v) => !v)}
          >
            {expanded ? (
              <ChevronUp className="w-3.5 h-3.5" />
            ) : (
              <ChevronDown className="w-3.5 h-3.5" />
            )}
            View
          </Button>
          <Button
            size="icon-sm"
            variant="ghost"
            className="text-muted-foreground hover:text-destructive"
            onClick={() => onDelete(sprint.id)}
            disabled={deleteDisabled}
          >
            <Trash2 className="w-3.5 h-3.5" />
          </Button>
        </div>
      </div>

      {expanded && (
        <div className="border-t border-border px-4 py-3 bg-muted/30">
          {taskIDsLoading ? (
            <p className="text-xs text-muted-foreground">Loading tasks...</p>
          ) : !taskIDs || taskIDs.length === 0 ? (
            <p className="text-xs text-muted-foreground">No tasks in this sprint.</p>
          ) : (
            <div className="flex flex-col gap-1.5">
              <p className="text-xs text-muted-foreground mb-1">
                {taskIDs.length} task{taskIDs.length !== 1 ? "s" : ""} in this sprint
              </p>
              <div className="flex flex-wrap gap-1.5">
                {taskIDs.map((id) => (
                  <code
                    key={id}
                    className="text-[10px] font-mono rounded bg-muted px-1.5 py-0.5 text-muted-foreground border border-border"
                  >
                    {id.slice(0, 8)}…
                  </code>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default function SprintsPage() {
  const { data: sprints, isLoading } = useSprints();
  const createSprint = useCreateSprint();
  const deleteSprint = useDeleteSprint();

  const [createOpen, setCreateOpen] = useState(false);
  const [form, setForm] = useState({
    name: "",
    start_date: "",
    end_date: "",
    goal: "",
  });

  const handleCreate = () => {
    if (!form.name.trim()) return;
    createSprint.mutate(
      {
        name: form.name.trim(),
        start_date: form.start_date,
        end_date: form.end_date,
        goal: form.goal.trim(),
      },
      {
        onSuccess: () => {
          toast.success("Sprint created");
          setCreateOpen(false);
          setForm({ name: "", start_date: "", end_date: "", goal: "" });
        },
        onError: () => toast.error("Failed to create sprint"),
      }
    );
  };

  const handleDelete = (id: string) => {
    deleteSprint.mutate(id, {
      onSuccess: () => toast.success("Sprint deleted"),
      onError: () => toast.error("Failed to delete sprint"),
    });
  };

  const handleClose = (open: boolean) => {
    if (!open) setForm({ name: "", start_date: "", end_date: "", goal: "" });
    setCreateOpen(open);
  };

  const list = sprints ?? [];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between animate-in fade-in-0 slide-in-from-bottom-3 duration-400">
        <div>
          <h2 className="text-xl font-bold tracking-tight">Sprints</h2>
          <p className="text-sm text-muted-foreground mt-0.5">
            Plan and track time-boxed iterations
          </p>
        </div>

        <Dialog open={createOpen} onOpenChange={handleClose}>
          <DialogTrigger asChild>
            <Button size="sm" className="gap-1.5">
              <Plus className="w-3.5 h-3.5" />
              New Sprint
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-md">
            <DialogHeader>
              <DialogTitle>New sprint</DialogTitle>
            </DialogHeader>
            <div className="flex flex-col gap-3 py-2">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="sprint-name">Name</Label>
                <Input
                  id="sprint-name"
                  placeholder="e.g. Sprint 1"
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="sprint-start">Start date</Label>
                  <Input
                    id="sprint-start"
                    type="date"
                    value={form.start_date}
                    onChange={(e) => setForm((f) => ({ ...f, start_date: e.target.value }))}
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="sprint-end">End date</Label>
                  <Input
                    id="sprint-end"
                    type="date"
                    value={form.end_date}
                    onChange={(e) => setForm((f) => ({ ...f, end_date: e.target.value }))}
                  />
                </div>
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="sprint-goal">
                  Goal{" "}
                  <span className="text-muted-foreground font-normal">(optional)</span>
                </Label>
                <textarea
                  id="sprint-goal"
                  placeholder="What do you want to achieve this sprint?"
                  value={form.goal}
                  onChange={(e) => setForm((f) => ({ ...f, goal: e.target.value }))}
                  rows={3}
                  className={cn(
                    "flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm",
                    "placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1",
                    "focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 resize-none"
                  )}
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => handleClose(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleCreate}
                disabled={!form.name.trim() || createSprint.isPending}
              >
                {createSprint.isPending ? "Creating..." : "Create"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {isLoading ? (
        <div className="flex flex-col gap-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-20 rounded-xl border border-border bg-card animate-pulse" />
          ))}
        </div>
      ) : list.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 gap-3 rounded-xl border border-border bg-card">
          <CalendarRange className="size-10 text-muted-foreground" />
          <div className="flex flex-col items-center gap-1">
            <p className="text-sm font-medium">No sprints yet</p>
            <p className="text-xs text-muted-foreground">Create a sprint to start planning</p>
          </div>
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {list.map((sprint, i) => (
            <div
              key={sprint.id}
              className="animate-in fade-in-0 slide-in-from-bottom-2 duration-400"
              style={{ animationDelay: `${i * 50}ms` }}
            >
              <SprintCard
                sprint={sprint}
                onDelete={handleDelete}
                deleteDisabled={deleteSprint.isPending}
              />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
