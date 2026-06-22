"use client";
import { Table2, Kanban, BarChart2, CalendarDays } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export type View = "table" | "kanban" | "gantt" | "weekly";

const VIEWS: { id: View; label: string; icon: React.ReactNode }[] = [
  { id: "table",   label: "Table view",   icon: <Table2 className="size-3.5" /> },
  { id: "kanban",  label: "Kanban view",  icon: <Kanban className="size-3.5" /> },
  { id: "gantt",   label: "Gantt view",   icon: <BarChart2 className="size-3.5" /> },
  { id: "weekly",  label: "Weekly view",  icon: <CalendarDays className="size-3.5" /> },
];

export function ViewToggle({ view, onChange }: { view: View; onChange: (v: View) => void }) {
  return (
    <div className="flex items-center gap-1 rounded-lg border border-border p-0.5">
      {VIEWS.map(({ id, label, icon }) => (
        <Button
          key={id}
          type="button"
          variant="ghost"
          size="icon-sm"
          className={cn("h-6 w-6", view === id && "bg-background shadow-sm")}
          onClick={() => onChange(id)}
          aria-label={label}
        >
          {icon}
        </Button>
      ))}
    </div>
  );
}
