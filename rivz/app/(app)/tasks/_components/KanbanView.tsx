"use client";
import { useState } from "react";
import {
  DndContext,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
  DragOverlay,
} from "@dnd-kit/core";
import {
  SortableContext,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useDroppable } from "@dnd-kit/core";
import { format, parseISO } from "date-fns";
import { CheckCircle2, Circle, Clock, XCircle, AlertCircle } from "lucide-react";
import { useUpdateTask, type Task } from "@/lib/tasks-hooks";
import { TaskForm } from "./TaskForm";
import { cn } from "@/lib/utils";

type Status = "todo" | "in_progress" | "done" | "failed";

const columns: { id: Status; label: string; color: string }[] = [
  { id: "todo", label: "Todo", color: "border-t-muted-foreground/30" },
  { id: "in_progress", label: "In Progress", color: "border-t-blue-500" },
  { id: "done", label: "Done", color: "border-t-emerald-500" },
  { id: "failed", label: "Failed", color: "border-t-rose-500" },
];

const statusIcon: Record<string, React.ReactNode> = {
  todo: <Circle className="size-3 text-muted-foreground" />,
  in_progress: <Clock className="size-3 text-blue-500" />,
  done: <CheckCircle2 className="size-3 text-emerald-500" />,
  failed: <XCircle className="size-3 text-rose-500" />,
};

const priorityDot: Record<string, string> = {
  low: "bg-emerald-500",
  medium: "bg-amber-500",
  high: "bg-rose-500",
};

function TaskCard({ task, onClick }: { task: Task; onClick: () => void }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: task.id });

  const isOverdue = task.due_date && new Date(task.due_date) < new Date() && task.status !== "done";

  return (
    <div
      ref={setNodeRef}
      style={{ transform: CSS.Transform.toString(transform), transition }}
      {...attributes}
      {...listeners}
      onClick={onClick}
      className={cn(
        "bg-card border border-border rounded-lg p-3 cursor-pointer hover:shadow-md transition-shadow select-none",
        isDragging && "opacity-50"
      )}
    >
      <div className="flex items-start gap-2">
        <span className={cn("mt-0.5 size-2 rounded-full shrink-0", priorityDot[task.priority])} />
        <p className="text-xs font-medium leading-snug line-clamp-2 flex-1">{task.title}</p>
      </div>
      {task.due_date && (
        <div className={cn("flex items-center gap-1 mt-2 text-[10px]", isOverdue ? "text-rose-500" : "text-muted-foreground")}>
          {isOverdue && <AlertCircle className="size-3" />}
          {format(parseISO(task.due_date), "MMM d")}
        </div>
      )}
      {task.tags.length > 0 && (
        <div className="flex flex-wrap gap-1 mt-2">
          {task.tags.slice(0, 3).map((tag) => (
            <span key={tag.id} className="px-1.5 py-0.5 rounded-full text-[10px] font-medium text-white" style={{ backgroundColor: tag.color }}>
              {tag.name}
            </span>
          ))}
        </div>
      )}
      <div className="flex items-center justify-between mt-2">
        {task.subtask_count > 0 && (
          <span className="text-[10px] text-muted-foreground">{task.subtasks_done}/{task.subtask_count} subtasks</span>
        )}
        {task.assignee_email && (
          <span className="ml-auto flex h-5 w-5 items-center justify-center rounded-full bg-secondary text-[10px] font-bold" title={task.assignee_email}>
            {task.assignee_email.slice(0, 2).toUpperCase()}
          </span>
        )}
      </div>
    </div>
  );
}

function Column({ id, label, color, tasks, onCardClick }: {
  id: Status; label: string; color: string;
  tasks: Task[];
  onCardClick: (task: Task) => void;
}) {
  const { setNodeRef, isOver } = useDroppable({ id });

  return (
    <div
      ref={setNodeRef}
      className={cn(
        "flex flex-col gap-2 min-h-64 rounded-xl border-t-2 border border-border bg-muted/20 p-3 transition-colors",
        color,
        isOver && "bg-primary/5 border-primary/30"
      )}
    >
      <div className="flex items-center justify-between mb-1">
        <span className="flex items-center gap-1.5 text-xs font-semibold">
          {statusIcon[id]}
          {label}
        </span>
        <span className="text-[10px] text-muted-foreground bg-muted rounded-full px-1.5 py-0.5">{tasks.length}</span>
      </div>
      <SortableContext items={tasks.map((t) => t.id)} strategy={verticalListSortingStrategy}>
        {tasks.map((task) => (
          <TaskCard key={task.id} task={task} onClick={() => onCardClick(task)} />
        ))}
      </SortableContext>
    </div>
  );
}

export function KanbanView({ tasks }: { tasks: Task[] }) {
  const updateTask = useUpdateTask();
  const sensors = useSensors(useSensor(PointerSensor));
  const [activeId, setActiveId] = useState<string | null>(null);
  const [editTask, setEditTask] = useState<Task | null>(null);

  const tasksByStatus = columns.reduce((acc, col) => {
    acc[col.id] = tasks.filter((t) => t.status === col.id);
    return acc;
  }, {} as Record<Status, Task[]>);

  const activeTask = activeId ? tasks.find((t) => t.id === activeId) : null;

  const handleDragEnd = (event: DragEndEvent) => {
    setActiveId(null);
    const { active, over } = event;
    if (!over) return;
    const overId = String(over.id);
    const isColumn = columns.some((c) => c.id === overId);
    if (isColumn && active.id !== over.id) {
      updateTask.mutate({ id: String(active.id), status: overId as Status });
    }
  };

  return (
    <>
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragStart={(e) => setActiveId(String(e.active.id))}
        onDragEnd={handleDragEnd}
      >
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {columns.map((col) => (
            <Column
              key={col.id}
              id={col.id}
              label={col.label}
              color={col.color}
              tasks={tasksByStatus[col.id]}
              onCardClick={setEditTask}
            />
          ))}
        </div>
        <DragOverlay>
          {activeTask && (
            <div className="bg-card border border-border rounded-lg p-3 shadow-lg opacity-90 text-xs font-medium">
              {activeTask.title}
            </div>
          )}
        </DragOverlay>
      </DndContext>

      {editTask && (
        <TaskForm
          open={!!editTask}
          onOpenChange={(o) => { if (!o) setEditTask(null); }}
          task={editTask}
        />
      )}
    </>
  );
}
