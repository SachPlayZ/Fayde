"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogTrigger,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import { Plus, Trash2, FolderKanban } from "lucide-react";
import { useProjects, useCreateProject, useDeleteProject } from "@/lib/projects-hooks";

const COLOR_SWATCHES = [
  { label: "Indigo", value: "#6366f1" },
  { label: "Rose", value: "#f43f5e" },
  { label: "Emerald", value: "#10b981" },
  { label: "Amber", value: "#f59e0b" },
  { label: "Sky", value: "#0ea5e9" },
  { label: "Purple", value: "#a855f7" },
];

export default function ProjectsPage() {
  const router = useRouter();
  const { data: projects, isLoading } = useProjects();
  const createProject = useCreateProject();
  const deleteProject = useDeleteProject();

  const [createOpen, setCreateOpen] = useState(false);
  const [form, setForm] = useState({
    name: "",
    description: "",
    color: COLOR_SWATCHES[0].value,
  });

  const handleCreate = () => {
    if (!form.name.trim()) return;
    createProject.mutate(
      {
        name: form.name.trim(),
        description: form.description.trim(),
        color: form.color,
      },
      {
        onSuccess: () => {
          toast.success("Project created");
          setCreateOpen(false);
          setForm({ name: "", description: "", color: COLOR_SWATCHES[0].value });
        },
        onError: () => toast.error("Failed to create project"),
      }
    );
  };

  const handleDelete = (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    deleteProject.mutate(id, {
      onSuccess: () => toast.success("Project deleted"),
      onError: () => toast.error("Failed to delete project"),
    });
  };

  const handleClose = (open: boolean) => {
    if (!open) setForm({ name: "", description: "", color: COLOR_SWATCHES[0].value });
    setCreateOpen(open);
  };

  const list = projects ?? [];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between animate-in fade-in-0 slide-in-from-bottom-3 duration-400">
        <div>
          <h2 className="text-xl font-bold tracking-tight">Projects</h2>
          <p className="text-sm text-muted-foreground mt-0.5">
            Organize your tasks into projects
          </p>
        </div>

        <Dialog open={createOpen} onOpenChange={handleClose}>
          <DialogTrigger asChild>
            <Button size="sm" className="gap-1.5">
              <Plus className="w-3.5 h-3.5" />
              New Project
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-md">
            <DialogHeader>
              <DialogTitle>New project</DialogTitle>
            </DialogHeader>
            <div className="flex flex-col gap-3 py-2">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="proj-name">Name</Label>
                <Input
                  id="proj-name"
                  placeholder="e.g. Website redesign"
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  onKeyDown={(e) => e.key === "Enter" && handleCreate()}
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="proj-desc">
                  Description{" "}
                  <span className="text-muted-foreground font-normal">(optional)</span>
                </Label>
                <Input
                  id="proj-desc"
                  placeholder="What's this project about?"
                  value={form.description}
                  onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                />
              </div>
              <div className="flex flex-col gap-2">
                <Label>Color</Label>
                <div className="flex gap-2">
                  {COLOR_SWATCHES.map((swatch) => (
                    <button
                      key={swatch.value}
                      type="button"
                      title={swatch.label}
                      onClick={() => setForm((f) => ({ ...f, color: swatch.value }))}
                      className={cn(
                        "w-7 h-7 rounded-full transition-all duration-150 ring-offset-background",
                        form.color === swatch.value
                          ? "ring-2 ring-offset-2 ring-foreground/50 scale-110"
                          : "hover:scale-105"
                      )}
                      style={{ backgroundColor: swatch.value }}
                    />
                  ))}
                </div>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => handleClose(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleCreate}
                disabled={!form.name.trim() || createProject.isPending}
              >
                {createProject.isPending ? "Creating..." : "Create"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="h-28 rounded-xl border border-border bg-card animate-pulse" />
          ))}
        </div>
      ) : list.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 gap-3 rounded-xl border border-border bg-card">
          <FolderKanban className="size-10 text-muted-foreground" />
          <div className="flex flex-col items-center gap-1">
            <p className="text-sm font-medium">No projects yet</p>
            <p className="text-xs text-muted-foreground">Create a project to organize your tasks</p>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          {list.map((project, i) => (
            <div
              key={project.id}
              onClick={() => router.push(`/tasks?project_id=${project.id}`)}
              className="group relative flex flex-col gap-3 rounded-xl border border-border bg-card p-4 cursor-pointer hover:-translate-y-0.5 hover:shadow-md hover:border-border/80 transition-all duration-300 animate-in fade-in-0 slide-in-from-bottom-3 duration-400"
              style={{ animationDelay: `${i * 40}ms` }}
            >
              <div className="flex items-start justify-between gap-2">
                <div className="flex items-center gap-2.5">
                  <div
                    className="w-8 h-8 rounded-full shrink-0"
                    style={{ backgroundColor: project.color }}
                  />
                  <span className="text-sm font-semibold leading-tight">{project.name}</span>
                </div>
                <Button
                  size="icon-sm"
                  variant="ghost"
                  className="text-muted-foreground hover:text-destructive opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
                  onClick={(e) => handleDelete(e, project.id)}
                  disabled={deleteProject.isPending}
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </Button>
              </div>
              {project.description && (
                <p className="text-xs text-muted-foreground line-clamp-2">
                  {project.description}
                </p>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
