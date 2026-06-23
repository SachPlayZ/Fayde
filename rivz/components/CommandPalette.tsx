"use client";
import { useEffect, useState, createContext, useContext, useCallback } from "react";
import { Command } from "cmdk";
import { useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import { useQuickCapture } from "@/lib/quick-capture-context";
import { useSearch } from "@/lib/search-hooks";
import {
  Search, Settings, ClipboardList, Inbox,
  Sun, Calendar, AlertCircle, CalendarClock, Zap,
  FileText, MessageSquare, LayoutDashboard, Flame, Target,
} from "lucide-react";

type PaletteContextType = {
  open: boolean;
  setOpen: (v: boolean) => void;
};

const PaletteContext = createContext<PaletteContextType>({ open: false, setOpen: () => {} });

export function useCommandPalette() {
  return useContext(PaletteContext);
}

export function CommandPaletteProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setOpen((v) => !v);
      }
    };
    window.addEventListener("keydown", down);
    return () => window.removeEventListener("keydown", down);
  }, []);

  return (
    <PaletteContext.Provider value={{ open, setOpen }}>
      {children}
      <CommandPaletteDialog open={open} setOpen={setOpen} />
    </PaletteContext.Provider>
  );
}

function CommandPaletteDialog({ open, setOpen }: { open: boolean; setOpen: (v: boolean) => void }) {
  const router = useRouter();
  const { setTheme, theme } = useTheme();
  const { openCapture } = useQuickCapture();
  const [query, setQuery] = useState("");
  const { data: results, isFetching } = useSearch(query);
  const searching = query.trim().length >= 2;

  const run = useCallback((fn: () => void) => {
    setOpen(false);
    setQuery("");
    fn();
  }, [setOpen]);

  if (!open) return null;

  const iconFor = (type: string) =>
    type === "note" ? (
      <FileText className="size-4 text-violet-500 shrink-0" />
    ) : type === "comment" ? (
      <MessageSquare className="size-4 text-blue-500 shrink-0" />
    ) : (
      <ClipboardList className="size-4 text-muted-foreground shrink-0" />
    );

  const gotoResult = (r: { type: string; id: string; task_id?: string }) => {
    if (r.type === "note") return router.push(`/docs?note=${r.id}`);
    if (r.type === "comment" && r.task_id) return router.push(`/tasks/${r.task_id}`);
    return router.push(`/tasks/${r.id}`);
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-20"
      onClick={() => setOpen(false)}
    >
      <div
        className="w-full max-w-lg bg-background border border-border rounded-xl shadow-2xl overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        <Command
          shouldFilter={false}
          className="[&_[cmdk-group-heading]]:text-xs [&_[cmdk-group-heading]]:text-muted-foreground [&_[cmdk-group-heading]]:px-3 [&_[cmdk-group-heading]]:pb-1 [&_[cmdk-group-heading]]:pt-2"
        >
          <div className="flex items-center border-b px-3">
            <Search className="size-4 text-muted-foreground mr-2 shrink-0" />
            <Command.Input
              value={query}
              onValueChange={setQuery}
              placeholder="Search tasks, docs, comments, or actions..."
              className="flex h-10 w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
              autoFocus
            />
          </div>
          <Command.List className="max-h-80 overflow-y-auto py-2">
            <Command.Empty className="py-6 text-center text-sm text-muted-foreground">
              {isFetching ? "Searching…" : "No results found."}
            </Command.Empty>

            {searching && (
              <Command.Group heading="Search results">
                {(results ?? []).map((r) => (
                  <Command.Item
                    key={`${r.type}-${r.id}`}
                    value={`${r.type}-${r.id}`}
                    onSelect={() => run(() => gotoResult(r))}
                    className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
                  >
                    {iconFor(r.type)}
                    <span className="flex flex-col min-w-0">
                      <span className="truncate">{r.title}</span>
                      {r.snippet && (
                        <span
                          className="truncate text-xs text-muted-foreground [&_b]:text-foreground [&_b]:font-semibold"
                          dangerouslySetInnerHTML={{ __html: r.snippet }}
                        />
                      )}
                    </span>
                    <span className="ml-auto text-[10px] uppercase text-muted-foreground shrink-0">
                      {r.type}
                    </span>
                  </Command.Item>
                ))}
              </Command.Group>
            )}

            {!searching && (
            <>
            <Command.Group heading="Actions">
              <Command.Item
                onSelect={() => run(() => openCapture())}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <Zap className="size-4 text-amber-500" /> Quick capture
                <span className="ml-auto text-xs text-muted-foreground">N</span>
              </Command.Item>
              <Command.Item
                onSelect={() => run(() => setTheme(theme === "dark" ? "light" : "dark"))}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <Sun className="size-4" /> Toggle theme
              </Command.Item>
            </Command.Group>

            <Command.Group heading="Views">
              <Command.Item
                onSelect={() => run(() => router.push("/dashboard"))}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <LayoutDashboard className="size-4 text-muted-foreground" /> Home
              </Command.Item>
              <Command.Item
                onSelect={() => run(() => router.push("/habits"))}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <Flame className="size-4 text-orange-500" /> Habits
              </Command.Item>
              <Command.Item
                onSelect={() => run(() => router.push("/goals"))}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <Target className="size-4 text-primary" /> Goals
              </Command.Item>
              <Command.Item
                onSelect={() => run(() => router.push("/tasks?list=inbox"))}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <Inbox className="size-4 text-muted-foreground" /> Inbox
              </Command.Item>
              <Command.Item
                onSelect={() => run(() => router.push("/tasks?list=today"))}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <Sun className="size-4 text-amber-500" /> Today
              </Command.Item>
              <Command.Item
                onSelect={() => run(() => router.push("/tasks?list=upcoming"))}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <Calendar className="size-4 text-blue-500" /> Upcoming
              </Command.Item>
              <Command.Item
                onSelect={() => run(() => router.push("/tasks?list=overdue"))}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <AlertCircle className="size-4 text-rose-500" /> Overdue
              </Command.Item>
              <Command.Item
                onSelect={() => run(() => router.push("/tasks/review"))}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <CalendarClock className="size-4 text-violet-500" /> Daily review
              </Command.Item>
              <Command.Item
                onSelect={() => run(() => router.push("/tasks"))}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <ClipboardList className="size-4" /> All tasks
              </Command.Item>
              <Command.Item
                onSelect={() => run(() => router.push("/docs"))}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <FileText className="size-4 text-violet-500" /> Docs
              </Command.Item>
              {/* Admin is conditionally shown via the existing pattern */}
              <Command.Item
                onSelect={() => run(() => router.push("/admin"))}
                className="flex items-center gap-2 px-3 py-1.5 text-sm cursor-pointer rounded-md hover:bg-muted aria-selected:bg-muted mx-1"
              >
                <Settings className="size-4" /> Admin
              </Command.Item>
            </Command.Group>
            </>
            )}
          </Command.List>
        </Command>
      </div>
    </div>
  );
}
