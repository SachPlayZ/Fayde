"use client";
import { useEffect, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { projectSchema, type ProjectInput } from "@/lib/schemas";
import {
  useCreateProject,
  useUpdateProject,
  useUploadProjectLogo,
  useUploadProjectBanner,
  useDeleteProjectLogo,
  useDeleteProjectBanner,
  type Project,
} from "@/lib/projects-hooks";
import { ApiError } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { TechStackInput } from "./TechStackInput";
import { ImageUploadField } from "./ImageUploadField";
import { toast } from "sonner";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  entry?: Project;
};

export function ProjectForm({ open, onOpenChange, entry }: Props) {
  const isEdit = !!entry;
  const createEntry = useCreateProject();
  const updateEntry = useUpdateProject();
  const uploadLogo = useUploadProjectLogo();
  const uploadBanner = useUploadProjectBanner();
  const deleteLogo = useDeleteProjectLogo();
  const deleteBanner = useDeleteProjectBanner();

  // Tracks the entry once it exists server-side (immediately after create),
  // so image uploads can happen in the same dialog session without reopening.
  const [savedEntry, setSavedEntry] = useState<Project | undefined>(entry);
  useEffect(() => {
    if (open) Promise.resolve().then(() => setSavedEntry(entry));
  }, [open, entry]);

  const {
    register,
    handleSubmit,
    setValue,
    control,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<ProjectInput>({
    resolver: zodResolver(projectSchema),
    defaultValues: {
      title: entry?.title ?? "",
      tagline: entry?.tagline ?? "",
      problem: entry?.problem ?? "",
      solution: entry?.solution ?? "",
      tech_stack: entry?.tech_stack ?? [],
      demo_url: entry?.demo_url ?? "",
      repo_url: entry?.repo_url ?? "",
      live_url: entry?.live_url ?? "",
    },
  });

  const [techStack] = useWatch({ control, name: ["tech_stack"] });

  const handleClose = () => {
    reset();
    setSavedEntry(undefined);
    onOpenChange(false);
  };

  const onSubmit = async (data: ProjectInput) => {
    const payload = {
      ...data,
      demo_url: data.demo_url || null,
      repo_url: data.repo_url || null,
      live_url: data.live_url || null,
    };
    try {
      if (savedEntry) {
        await updateEntry.mutateAsync({ id: savedEntry.id, ...payload });
        toast.success("Project updated");
        handleClose();
      } else {
        const created = await createEntry.mutateAsync(payload);
        toast.success("Project created — add a logo/banner below");
        setSavedEntry(created);
      }
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Something went wrong");
    }
  };

  return (
    <Dialog open={open} onOpenChange={(o) => (o ? onOpenChange(true) : handleClose())}>
      <DialogContent className="sm:max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit project" : "New project"}</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="title">Title <span className="text-rose-500">*</span></Label>
            <Input id="title" placeholder="Your hackathon project name" aria-invalid={!!errors.title} {...register("title")} />
            {errors.title && <p className="text-xs text-rose-500">{errors.title.message}</p>}
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="tagline">Tagline</Label>
            <Input id="tagline" placeholder="One sentence summary" {...register("tagline")} />
            {errors.tagline && <p className="text-xs text-rose-500">{errors.tagline.message}</p>}
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="problem">Problem</Label>
            <Textarea id="problem" placeholder="What problem does this solve?" rows={2} {...register("problem")} />
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="solution">Solution</Label>
            <Textarea id="solution" placeholder="How does your project solve it?" rows={2} {...register("solution")} />
          </div>

          <div className="flex flex-col gap-1.5">
            <Label>Tech stack</Label>
            <TechStackInput value={techStack ?? []} onChange={(items) => setValue("tech_stack", items)} />
          </div>

          <div className="grid grid-cols-1 gap-3">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="demo_url">Demo video URL</Label>
              <Input id="demo_url" placeholder="https://youtube.com/..." {...register("demo_url")} />
              {errors.demo_url && <p className="text-xs text-rose-500">{errors.demo_url.message}</p>}
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="repo_url">Repository URL</Label>
              <Input id="repo_url" placeholder="https://github.com/you/project" {...register("repo_url")} />
              {errors.repo_url && <p className="text-xs text-rose-500">{errors.repo_url.message}</p>}
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="live_url">Live URL</Label>
              <Input id="live_url" placeholder="https://your-project.vercel.app" {...register("live_url")} />
              {errors.live_url && <p className="text-xs text-rose-500">{errors.live_url.message}</p>}
            </div>
          </div>

          {savedEntry ? (
            <div className="grid grid-cols-2 gap-4">
              <ImageUploadField
                label="Logo"
                currentUrl={savedEntry.logo_url}
                aspect="square"
                onUpload={(file) =>
                  uploadLogo.mutateAsync({ id: savedEntry.id, file }).then((e) => setSavedEntry(e))
                }
                onRemove={() =>
                  deleteLogo.mutateAsync(savedEntry.id).then(() => setSavedEntry({ ...savedEntry, logo_url: null }))
                }
              />
              <ImageUploadField
                label="Banner"
                currentUrl={savedEntry.banner_url}
                aspect="wide"
                onUpload={(file) =>
                  uploadBanner.mutateAsync({ id: savedEntry.id, file }).then((e) => setSavedEntry(e))
                }
                onRemove={() =>
                  deleteBanner.mutateAsync(savedEntry.id).then(() => setSavedEntry({ ...savedEntry, banner_url: null }))
                }
              />
            </div>
          ) : (
            <p className="text-xs text-muted-foreground">Save the project first to add a logo or banner.</p>
          )}

          <DialogFooter className="mt-2 gap-2">
            <Button type="button" variant="outline" onClick={handleClose}>
              {savedEntry && !isEdit ? "Done" : "Cancel"}
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Saving..." : savedEntry ? "Save changes" : "Create project"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
