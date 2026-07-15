"use client";
import { useState } from "react";
import {
  useShowcaseEntries,
  useDeleteShowcaseEntry,
  type ShowcaseEntry,
} from "@/lib/showcase-hooks";
import { Button } from "@/components/ui/button";
import { Sparkles, Plus } from "lucide-react";
import { toast } from "sonner";
import { ShowcaseEntryCard } from "./_components/ShowcaseEntryCard";
import { ShowcaseEntryForm } from "./_components/ShowcaseEntryForm";

export default function ShowcasePage() {
  const { data: entries, isLoading } = useShowcaseEntries();
  const deleteEntry = useDeleteShowcaseEntry();

  const [formOpen, setFormOpen] = useState(false);
  const [editingEntry, setEditingEntry] = useState<ShowcaseEntry | undefined>(undefined);

  const openCreate = () => {
    setEditingEntry(undefined);
    setFormOpen(true);
  };

  const openEdit = (entry: ShowcaseEntry) => {
    setEditingEntry(entry);
    setFormOpen(true);
  };

  const handleDelete = (id: string) => {
    if (!confirm("Delete this project? This cannot be undone.")) return;
    deleteEntry.mutate(id, {
      onSuccess: () => toast.success("Project deleted"),
      onError: () => toast.error("Failed to delete project"),
    });
  };

  return (
    <div className="flex flex-col gap-6 max-w-4xl">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="text-xl font-bold tracking-tight">Showcase</h2>
          <p className="text-sm text-muted-foreground mt-0.5">
            Publish your hackathon projects on a public page and embed them anywhere via API.
          </p>
        </div>
        <Button onClick={openCreate}>
          <Plus className="size-4 mr-1" /> New project
        </Button>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {Array.from({ length: 2 }).map((_, i) => (
            <div key={i} className="h-48 rounded-xl border border-border bg-card animate-pulse" />
          ))}
        </div>
      ) : (entries?.length ?? 0) === 0 ? (
        <div className="flex flex-col items-center justify-center gap-3 py-16 rounded-xl border border-border bg-card text-muted-foreground">
          <Sparkles className="size-8" />
          <p className="text-sm">No projects yet. Add your first hackathon project above.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {entries!.map((entry) => (
            <ShowcaseEntryCard
              key={entry.id}
              entry={entry}
              onEdit={() => openEdit(entry)}
              onDelete={() => handleDelete(entry.id)}
            />
          ))}
        </div>
      )}

      <ShowcaseEntryForm open={formOpen} onOpenChange={setFormOpen} entry={editingEntry} />
    </div>
  );
}
