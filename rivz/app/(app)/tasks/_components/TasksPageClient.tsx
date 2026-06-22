"use client";
import { useSearchParams, useRouter, usePathname } from "next/navigation";
import { useState, useRef, useEffect, useCallback } from "react";
import { useTasks, useBulkUpdateTasks, useBulkDeleteTasks, type Task } from "@/lib/tasks-hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableHeader,
  TableHead,
  TableBody,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { TaskRow } from "./TaskRow";
import { TaskForm } from "./TaskForm";
import { Pagination } from "./Pagination";
import { KanbanView } from "./KanbanView";
import { ViewToggle } from "./ViewToggle";
import { Plus, Search, ClipboardList, ArrowUp, ArrowDown, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

const PAGE_LIMIT = 10;

const statusFilters = [
  { value: "all", label: "All" },
  { value: "todo", label: "Todo" },
  { value: "in_progress", label: "In Progress" },
  { value: "done", label: "Done" },
  { value: "failed", label: "Failed" },
];

type View = "table" | "kanban";

export function TasksPageClient() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const [newTaskOpen, setNewTaskOpen] = useState(false);
  const [editTask, setEditTask] = useState<Task | null>(null);

  const status = searchParams.get("status") ?? "";
  const search = searchParams.get("search") ?? "";
  const sort = searchParams.get("sort") ?? "created_at";
  const order = searchParams.get("order") ?? "desc";
  const page = Number(searchParams.get("page") ?? "1");

  const [searchInput, setSearchInput] = useState(search);
  const [prevSearch, setPrevSearch] = useState(search);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const searchRef = useRef<HTMLInputElement>(null);

  // View preference
  const [view, setView] = useState<View>(() => {
    if (typeof window !== "undefined") {
      return (localStorage.getItem("task-view") as View) ?? "table";
    }
    return "table";
  });

  // Bulk selection
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const bulkUpdate = useBulkUpdateTasks();
  const bulkDelete = useBulkDeleteTasks();

  if (prevSearch !== search) {
    setPrevSearch(search);
    setSearchInput(search);
  }

  // Open edit from URL param ?edit=<id>
  const editId = searchParams.get("edit");
  const { data: allData } = useTasks({ limit: 100 });
  useEffect(() => {
    if (editId && allData) {
      const t = allData.data.find((t) => t.id === editId);
      if (t) setEditTask(t);
    }
  }, [editId, allData]);

  // Open new task from URL param ?new=1
  const newParam = searchParams.get("new");
  useEffect(() => {
    if (newParam === "1") {
      setNewTaskOpen(true);
      const p = new URLSearchParams(searchParams.toString());
      p.delete("new");
      router.replace(`${pathname}?${p.toString()}`);
    }
  }, [newParam, pathname, router, searchParams]);

  const updateParams = (updates: Record<string, string | undefined | null>) => {
    const params = new URLSearchParams(searchParams.toString());
    Object.entries(updates).forEach(([key, value]) => {
      if (value === undefined || value === null || value === "") {
        params.delete(key);
      } else {
        params.set(key, value);
      }
    });
    if (!("page" in updates)) params.set("page", "1");
    router.push(`${pathname}?${params.toString()}`);
  };

  const handleSearchChange = (value: string) => {
    setSearchInput(value);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      updateParams({ search: value || undefined });
    }, 400);
  };

  const handleToggleOrder = () => {
    updateParams({ order: order === "asc" ? "desc" : "asc" });
  };

  const handleViewChange = (v: View) => {
    setView(v);
    localStorage.setItem("task-view", v);
  };

  const { data, isLoading } = useTasks({
    status: status || undefined,
    search: search || undefined,
    sort,
    order,
    page,
    limit: PAGE_LIMIT,
  });

  const tasks = data?.data ?? [];
  const total = data?.total ?? 0;

  // Keyboard shortcuts
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;
      if (e.key === "n") { e.preventDefault(); setNewTaskOpen(true); }
      if (e.key === "/" ) { e.preventDefault(); searchRef.current?.focus(); }
      if (e.key === "Escape") { setNewTaskOpen(false); setEditTask(null); }
    },
    []
  );

  useEffect(() => {
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);

  const handleSelectChange = (id: string, checked: boolean) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (checked) next.add(id); else next.delete(id);
      return next;
    });
  };

  const handleSelectAll = (checked: boolean) => {
    setSelected(checked ? new Set(tasks.map((t) => t.id)) : new Set());
  };

  const handleBulkStatus = async (s: string) => {
    await bulkUpdate.mutateAsync({ ids: [...selected], status: s });
    setSelected(new Set());
    toast.success(`${selected.size} tasks updated`);
  };

  const handleBulkDelete = async () => {
    await bulkDelete.mutateAsync([...selected]);
    setSelected(new Set());
    toast.success(`${selected.size} tasks deleted`);
  };

  return (
    <div className="flex flex-col gap-6">
      {/* Page header */}
      <div className="flex items-center justify-between animate-in fade-in-0 slide-in-from-bottom-3 duration-400 stagger-1">
        <div>
          <h2 className="text-xl font-bold tracking-tight">My Tasks</h2>
          {!isLoading && (
            <p className="text-sm text-muted-foreground mt-0.5">
              {total} {total === 1 ? "task" : "tasks"}
            </p>
          )}
        </div>
        <Button onClick={() => setNewTaskOpen(true)} size="sm">
          <Plus className="w-4 h-4" />
          New Task
        </Button>
      </div>

      {/* Filter bar */}
      <div className="flex flex-wrap gap-2 items-center animate-in fade-in-0 slide-in-from-bottom-3 duration-400 stagger-2">
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground pointer-events-none" />
          <Input
            ref={searchRef}
            placeholder="Search tasks..."
            value={searchInput}
            onChange={(e) => handleSearchChange(e.target.value)}
            className="pl-8 w-52"
          />
        </div>

        <div className="flex items-center gap-1 rounded-xl bg-muted p-1">
          {statusFilters.map((f) => {
            const active = (status || "all") === f.value;
            return (
              <button
                key={f.value}
                onClick={() => updateParams({ status: f.value === "all" ? undefined : f.value })}
                className={cn(
                  "rounded-lg px-2.5 py-1 text-xs font-medium transition-all duration-150",
                  active ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground"
                )}
              >
                {f.label}
              </button>
            );
          })}
        </div>

        <Select value={sort} onValueChange={(val) => updateParams({ sort: val })}>
          <SelectTrigger className="w-36 text-xs h-7">
            <SelectValue placeholder="Sort by" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="created_at">Created date</SelectItem>
              <SelectItem value="due_date">Due date</SelectItem>
              <SelectItem value="priority">Priority</SelectItem>
              <SelectItem value="sort_order">Custom order</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>

        <Button variant="outline" size="sm" onClick={handleToggleOrder} className="h-7 text-xs gap-1">
          {order === "asc" ? <ArrowUp className="w-3 h-3" /> : <ArrowDown className="w-3 h-3" />}
          {order === "asc" ? "Asc" : "Desc"}
        </Button>

        <ViewToggle view={view} onChange={handleViewChange} />
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="flex flex-col gap-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-14 w-full rounded-xl" />
          ))}
        </div>
      ) : tasks.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 gap-4 text-center">
          <div className="flex items-center justify-center w-16 h-16 rounded-2xl bg-muted">
            <ClipboardList className="w-7 h-7 text-muted-foreground" />
          </div>
          <div>
            <p className="font-semibold">No tasks found</p>
            <p className="text-sm text-muted-foreground mt-1">
              {search || status ? "Try adjusting your filters" : "Create your first task to get started"}
            </p>
          </div>
          {!search && !status && (
            <Button onClick={() => setNewTaskOpen(true)} size="sm">
              <Plus className="w-4 h-4" />
              Create your first task
            </Button>
          )}
        </div>
      ) : view === "kanban" ? (
        <KanbanView tasks={tasks} />
      ) : (
        <div className="animate-in fade-in-0 slide-in-from-bottom-2 duration-500 stagger-3">
          <div className="hidden md:block rounded-xl border border-border overflow-hidden bg-card shadow-sm">
            <Table>
              <TableHeader>
                <TableRow className="bg-muted/40 hover:bg-muted/40">
                  <TableHead className="w-8">
                    <input
                      type="checkbox"
                      checked={selected.size === tasks.length && tasks.length > 0}
                      onChange={(e) => handleSelectAll(e.target.checked)}
                      className="size-4 rounded border-input accent-primary cursor-pointer"
                    />
                  </TableHead>
                  <TableHead>Title</TableHead>
                  <TableHead className="w-28">Status</TableHead>
                  <TableHead className="w-28">Priority</TableHead>
                  <TableHead className="w-36">Due Date</TableHead>
                  <TableHead className="w-28" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {tasks.map((task, i) => (
                  <TaskRow
                    key={task.id}
                    task={task}
                    index={i}
                    search={search}
                    selected={selected.has(task.id)}
                    onSelectChange={handleSelectChange}
                  />
                ))}
              </TableBody>
            </Table>
          </div>

          <div className="md:hidden flex flex-col gap-2">
            {tasks.map((task, i) => (
              <TaskRow key={task.id} task={task} index={i} search={search} />
            ))}
          </div>
        </div>
      )}

      {/* Pagination */}
      {!isLoading && total > PAGE_LIMIT && (
        <Pagination page={page} total={total} limit={PAGE_LIMIT} />
      )}

      {/* Bulk action bar */}
      {selected.size > 0 && (
        <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 bg-background border border-border rounded-xl shadow-2xl px-4 py-2.5 animate-in slide-in-from-bottom-3">
          <span className="text-sm font-medium whitespace-nowrap">{selected.size} selected</span>
          <Select onValueChange={handleBulkStatus}>
            <SelectTrigger className="h-7 text-xs w-32">
              <SelectValue placeholder="Set status" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="todo">Todo</SelectItem>
                <SelectItem value="in_progress">In Progress</SelectItem>
                <SelectItem value="done">Done</SelectItem>
                <SelectItem value="failed">Failed</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
          <Select onValueChange={(p) => bulkUpdate.mutateAsync({ ids: [...selected], priority: p }).then(() => { setSelected(new Set()); toast.success(`${selected.size} updated`); })}>
            <SelectTrigger className="h-7 text-xs w-32">
              <SelectValue placeholder="Set priority" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="low">Low</SelectItem>
                <SelectItem value="medium">Medium</SelectItem>
                <SelectItem value="high">High</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
          <Button variant="destructive" size="sm" className="h-7 text-xs" onClick={handleBulkDelete}>
            Delete
          </Button>
          <Button variant="ghost" size="icon-sm" className="h-7 w-7" onClick={() => setSelected(new Set())}>
            <X className="size-3.5" />
          </Button>
        </div>
      )}

      <TaskForm open={newTaskOpen} onOpenChange={setNewTaskOpen} />
      {editTask && (
        <TaskForm
          open={!!editTask}
          onOpenChange={(o) => { if (!o) { setEditTask(null); const p = new URLSearchParams(searchParams.toString()); p.delete("edit"); router.replace(`${pathname}?${p.toString()}`); } }}
          task={editTask}
        />
      )}
    </div>
  );
}
