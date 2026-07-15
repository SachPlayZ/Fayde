"use client";
import { useState } from "react";
import {
  useProjects,
  useDeleteProject,
  type Project,
} from "@/lib/projects-hooks";
import { Button } from "@/components/ui/button";
import { FolderKanban, Plus } from "lucide-react";
import { toast } from "sonner";
import { ProjectCard } from "./_components/ProjectCard";
import { ProjectForm } from "./_components/ProjectForm";

export default function ProjectsPage() {
  const { data: entries, isLoading } = useProjects();
  const deleteEntry = useDeleteProject();

  const [formOpen, setFormOpen] = useState(false);
  const [editingEntry, setEditingEntry] = useState<Project | undefined>(undefined);

  const openCreate = () => {
    setEditingEntry(undefined);
    setFormOpen(true);
  };

  const openEdit = (entry: Project) => {
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
          <h2 className="text-xl font-bold tracking-tight">Projects</h2>
          <p className="text-sm text-muted-foreground mt-0.5">
            Publish your portfolio projects on a public page and embed them anywhere via the API.
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
          <FolderKanban className="size-8" />
          <p className="text-sm">No projects yet. Add your first project above.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {entries!.map((entry) => (
            <ProjectCard
              key={entry.id}
              entry={entry}
              onEdit={() => openEdit(entry)}
              onDelete={() => handleDelete(entry.id)}
            />
          ))}
        </div>
      )}

      <ProjectForm open={formOpen} onOpenChange={setFormOpen} entry={editingEntry} />
    </div>
  );
}
