"use client";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Pencil, Trash2, ExternalLink, GitBranch, PlayCircle, Star } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Project } from "@/lib/projects-hooks";

type Props = {
  entry: Project;
  onEdit: () => void;
  onDelete: () => void;
};

export function ProjectCard({ entry, onEdit, onDelete }: Props) {
  return (
    <div className="rounded-xl border border-border bg-card overflow-hidden group">
      {entry.banner_url ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img src={entry.banner_url} alt="" className="h-28 w-full object-cover" />
      ) : (
        <div className="h-28 w-full bg-muted" />
      )}
      <div className="p-4 flex flex-col gap-3">
        <div className="flex items-start gap-3">
          {entry.logo_url ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={entry.logo_url} alt="" className="size-10 rounded-lg border border-border object-cover shrink-0 -mt-8 bg-background" />
          ) : (
            <div className="size-10 rounded-lg border border-border bg-muted shrink-0 -mt-8" />
          )}
          <div className="min-w-0 flex-1">
            <h3 className="text-sm font-semibold leading-snug truncate">{entry.title}</h3>
            {entry.tagline && <p className="text-xs text-muted-foreground line-clamp-2">{entry.tagline}</p>}
          </div>
        </div>

        {entry.tech_stack.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {entry.tech_stack.map((t, i) => (
              <Badge
                key={i}
                variant="outline"
                className={cn(
                  "text-[10px] gap-1",
                  t.is_sponsor && "border-amber-500/30 text-amber-700 dark:text-amber-400"
                )}
              >
                {t.is_sponsor && <Star className="size-2.5 fill-current" />}
                {t.name}
              </Badge>
            ))}
          </div>
        )}

        <div className="flex items-center gap-3 text-xs text-muted-foreground">
          {entry.demo_url && (
            <a href={entry.demo_url} target="_blank" rel="noopener noreferrer" className="flex items-center gap-1 hover:text-foreground">
              <PlayCircle className="size-3.5" /> Demo
            </a>
          )}
          {entry.repo_url && (
            <a href={entry.repo_url} target="_blank" rel="noopener noreferrer" className="flex items-center gap-1 hover:text-foreground">
              <GitBranch className="size-3.5" /> Repo
            </a>
          )}
          {entry.live_url && (
            <a href={entry.live_url} target="_blank" rel="noopener noreferrer" className="flex items-center gap-1 hover:text-foreground">
              <ExternalLink className="size-3.5" /> Live
            </a>
          )}
        </div>

        <div className="flex items-center gap-2 pt-1 border-t border-border -mx-4 px-4 mt-1">
          <Button type="button" variant="ghost" size="sm" className="h-7 text-xs gap-1.5 mt-2" onClick={onEdit}>
            <Pencil className="size-3" /> Edit
          </Button>
          <Button type="button" variant="ghost" size="sm" className="h-7 text-xs gap-1.5 mt-2 text-muted-foreground hover:text-rose-500" onClick={onDelete}>
            <Trash2 className="size-3" /> Delete
          </Button>
        </div>
      </div>
    </div>
  );
}
