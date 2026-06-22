"use client";
import { Table2, Kanban } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type View = "table" | "kanban";

export function ViewToggle({ view, onChange }: { view: View; onChange: (v: View) => void }) {
  return (
    <div className="flex items-center gap-1 rounded-lg border border-border p-0.5">
      <Button
        type="button"
        variant="ghost"
        size="icon-sm"
        className={cn("h-6 w-6", view === "table" && "bg-background shadow-sm")}
        onClick={() => onChange("table")}
        aria-label="Table view"
      >
        <Table2 className="size-3.5" />
      </Button>
      <Button
        type="button"
        variant="ghost"
        size="icon-sm"
        className={cn("h-6 w-6", view === "kanban" && "bg-background shadow-sm")}
        onClick={() => onChange("kanban")}
        aria-label="Kanban view"
      >
        <Kanban className="size-3.5" />
      </Button>
    </div>
  );
}
