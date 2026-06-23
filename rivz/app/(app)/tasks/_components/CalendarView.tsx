"use client";
import { useMemo, useState, useCallback, useEffect } from "react";
import {
  startOfMonth,
  endOfMonth,
  startOfWeek,
  endOfWeek,
  eachDayOfInterval,
  isSameMonth,
  isSameDay,
  isToday,
  format,
  addMonths,
  parseISO,
} from "date-fns";
import {
  DndContext,
  useDraggable,
  useDroppable,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import { ChevronLeft, ChevronRight, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { Task } from "@/lib/tasks-hooks";

const PRIORITY: Record<string, string> = {
  high: "border-l-rose-500",
  medium: "border-l-amber-500",
  low: "border-l-blue-500",
};

function TaskChip({ task, onClick }: { task: Task; onClick: () => void }) {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({ id: task.id });
  return (
    <button
      ref={setNodeRef}
      {...attributes}
      {...listeners}
      onClick={onClick}
      className={cn(
        "block w-full truncate rounded border-l-2 bg-card px-1.5 py-0.5 text-left text-[11px] shadow-sm hover:bg-muted transition-colors",
        PRIORITY[task.priority] ?? "border-l-border",
        (task.status === "done" || task.status === "failed") && "line-through opacity-60",
        task.status === "done" && "text-emerald-500",
        task.status === "failed" && "text-rose-500",
        isDragging && "opacity-40"
      )}
      title={task.title}
    >
      {task.title}
    </button>
  );
}

function DayCell({
  day,
  monthAnchor,
  tasks,
  onTaskClick,
  onNewTask,
}: {
  day: Date;
  monthAnchor: Date;
  tasks: Task[];
  onTaskClick: (t: Task) => void;
  onNewTask: (date: string) => void;
}) {
  const [hovered, setHovered] = useState(false);
  const { setNodeRef, isOver } = useDroppable({ id: format(day, "yyyy-MM-dd") });
  const outside = !isSameMonth(day, monthAnchor);
  const dateKey = format(day, "yyyy-MM-dd");
  return (
    <div
      ref={setNodeRef}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      className={cn(
        "min-h-24 border-b border-r border-border p-1 flex flex-col gap-0.5 relative",
        outside && "bg-muted/30",
        isOver && "bg-primary/10 ring-1 ring-inset ring-primary"
      )}
    >
      <div className="flex items-center justify-between mb-0.5">
        <button
          onClick={() => onNewTask(dateKey)}
          className={cn(
            "size-4 flex items-center justify-center rounded border border-dashed border-muted-foreground/40 text-muted-foreground/50 hover:border-primary/60 hover:text-primary transition-all duration-150",
            hovered ? "opacity-100" : "opacity-0"
          )}
          title={`Add task on ${format(day, "MMM d")}`}
        >
          <Plus className="size-2.5" />
        </button>
        <span
          className={cn(
            "text-[11px] font-medium size-5 flex items-center justify-center rounded-full",
            isToday(day) ? "bg-primary text-primary-foreground" : "text-muted-foreground"
          )}
        >
          {format(day, "d")}
        </span>
      </div>
      <div className="flex flex-col gap-0.5 overflow-hidden">
        {tasks.slice(0, 4).map((t) => (
          <TaskChip key={t.id} task={t} onClick={() => onTaskClick(t)} />
        ))}
        {tasks.length > 4 && (
          <span className="text-[10px] text-muted-foreground pl-1">+{tasks.length - 4} more</span>
        )}
      </div>
    </div>
  );
}

export function getCalendarRange(anchor: Date) {
  const start = startOfWeek(startOfMonth(anchor), { weekStartsOn: 0 });
  const end = endOfWeek(endOfMonth(anchor), { weekStartsOn: 0 });
  return { from: start.toISOString(), to: end.toISOString() };
}

export function CalendarView({
  tasks,
  onTaskClick,
  onReschedule,
  onNewTask,
  onRangeChange,
}: {
  tasks: Task[];
  onTaskClick: (t: Task) => void;
  onReschedule: (taskId: string, date: string) => void;
  onNewTask: (date: string) => void;
  onRangeChange?: (from: string, to: string) => void;
}) {
  const [anchor, setAnchor] = useState(() => new Date());
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 4 } }));
  const handleNewTask = useCallback((date: string) => onNewTask(date), [onNewTask]);

  useEffect(() => {
    if (!onRangeChange) return;
    const { from, to } = getCalendarRange(anchor);
    onRangeChange(from, to);
  }, [anchor, onRangeChange]);

  const days = useMemo(() => {
    const start = startOfWeek(startOfMonth(anchor), { weekStartsOn: 0 });
    const end = endOfWeek(endOfMonth(anchor), { weekStartsOn: 0 });
    return eachDayOfInterval({ start, end });
  }, [anchor]);

  const byDay = useMemo(() => {
    const map = new Map<string, Task[]>();
    for (const t of tasks) {
      if (!t.due_date) continue;
      const key = format(parseISO(t.due_date), "yyyy-MM-dd");
      if (!map.has(key)) map.set(key, []);
      map.get(key)!.push(t);
    }
    return map;
  }, [tasks]);

  const handleDragEnd = (e: DragEndEvent) => {
    if (!e.over) return;
    const taskId = String(e.active.id);
    const date = String(e.over.id); // yyyy-MM-dd
    const task = tasks.find((t) => t.id === taskId);
    if (task && (!task.due_date || !isSameDay(parseISO(task.due_date), parseISO(date)))) {
      onReschedule(taskId, date);
    }
  };

  return (
    <div className="animate-in fade-in-0 duration-300">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-semibold">{format(anchor, "MMMM yyyy")}</h3>
        <div className="flex items-center gap-1">
          <Button variant="outline" size="icon-sm" onClick={() => setAnchor((a) => addMonths(a, -1))}>
            <ChevronLeft className="size-4" />
          </Button>
          <Button variant="outline" size="sm" onClick={() => setAnchor(new Date())}>
            Today
          </Button>
          <Button variant="outline" size="icon-sm" onClick={() => setAnchor((a) => addMonths(a, 1))}>
            <ChevronRight className="size-4" />
          </Button>
        </div>
      </div>
      <div className="rounded-xl border-t border-l border-border overflow-hidden bg-card shadow-sm">
        <div className="grid grid-cols-7">
          {["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].map((d) => (
            <div key={d} className="border-b border-r border-border bg-muted/40 px-2 py-1.5 text-[11px] font-medium text-muted-foreground">
              {d}
            </div>
          ))}
        </div>
        <DndContext sensors={sensors} onDragEnd={handleDragEnd}>
          <div className="grid grid-cols-7">
            {days.map((day) => (
              <DayCell
                key={day.toISOString()}
                day={day}
                monthAnchor={anchor}
                tasks={byDay.get(format(day, "yyyy-MM-dd")) ?? []}
                onTaskClick={onTaskClick}
                onNewTask={handleNewTask}
              />
            ))}
          </div>
        </DndContext>
      </div>
    </div>
  );
}
