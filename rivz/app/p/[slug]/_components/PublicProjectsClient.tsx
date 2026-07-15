"use client";
import { useEffect, useState } from "react";
import Image from "next/image";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { Loader2, ExternalLink, GitBranch, PlayCircle, Star } from "lucide-react";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

type TechItem = { name: string; is_sponsor: boolean };

type PublicEntry = {
  id: string;
  title: string;
  tagline: string;
  problem: string;
  solution: string;
  tech_stack: TechItem[];
  demo_url: string | null;
  repo_url: string | null;
  live_url: string | null;
  logo_url: string | null;
  banner_url: string | null;
};

type PublicProjects = {
  owner_display_name: string;
  projects: PublicEntry[];
};

function absoluteImageURL(url: string | null): string | null {
  if (!url) return null;
  return url.startsWith("http") ? url : `${BASE_URL}${url}`;
}

export function PublicProjectsClient({ slug }: { slug: string }) {
  const [data, setData] = useState<PublicProjects | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!slug) return;
    fetch(`${BASE_URL}/p/${slug}`)
      .then((res) => {
        if (!res.ok) throw new Error(res.status === 404 ? "Projects page not found" : "Failed to load projects");
        return res.json() as Promise<PublicProjects>;
      })
      .then((d) => {
        setData(d);
        setLoading(false);
      })
      .catch((err: Error) => {
        setError(err.message);
        setLoading(false);
      });
  }, [slug]);

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <header className="border-b border-border bg-background/90 backdrop-blur-md">
        <div className="max-w-3xl mx-auto px-4 h-14 flex items-center">
          <span className="flex items-center gap-2">
            <Image src="/logo.png" alt="Fayde" width={24} height={24} className="size-6 rounded-md" />
            <span className="font-bold text-sm tracking-tight">Fayde</span>
          </span>
        </div>
      </header>

      <main className="flex-1 max-w-3xl w-full mx-auto px-4 py-12">
        {loading ? (
          <div className="flex flex-col items-center justify-center py-24 gap-3">
            <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
            <p className="text-sm text-muted-foreground">Loading projects…</p>
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center py-24 gap-3 text-center">
            <p className="text-sm font-medium">{error}</p>
            <p className="text-xs text-muted-foreground">This link may be invalid.</p>
          </div>
        ) : data ? (
          <div className="flex flex-col gap-8 animate-in fade-in-0 slide-in-from-bottom-3 duration-400">
            <div>
              <h1 className="text-2xl font-bold tracking-tight leading-tight">
                {data.owner_display_name}&rsquo;s Projects
              </h1>
              <p className="text-sm text-muted-foreground mt-1">
                {data.projects.length} project{data.projects.length === 1 ? "" : "s"}
              </p>
            </div>

            {data.projects.length === 0 ? (
              <p className="text-sm text-muted-foreground">No projects published yet.</p>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {data.projects.map((entry) => (
                  <div key={entry.id} className="rounded-xl border border-border bg-card overflow-hidden">
                    {entry.banner_url ? (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img src={absoluteImageURL(entry.banner_url)!} alt="" className="h-28 w-full object-cover" />
                    ) : (
                      <div className="h-28 w-full bg-muted" />
                    )}
                    <div className="p-4 flex flex-col gap-3">
                      <div className="flex items-start gap-3">
                        {entry.logo_url ? (
                          // eslint-disable-next-line @next/next/no-img-element
                          <img
                            src={absoluteImageURL(entry.logo_url)!}
                            alt=""
                            className="size-10 rounded-lg border border-border object-cover shrink-0 -mt-8 bg-background"
                          />
                        ) : (
                          <div className="size-10 rounded-lg border border-border bg-muted shrink-0 -mt-8" />
                        )}
                        <div className="min-w-0 flex-1">
                          <h3 className="text-sm font-semibold leading-snug">{entry.title}</h3>
                          {entry.tagline && <p className="text-xs text-muted-foreground">{entry.tagline}</p>}
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

                      {(entry.problem || entry.solution) && (
                        <p className="text-xs text-muted-foreground leading-relaxed">
                          {entry.problem && <span>{entry.problem} </span>}
                          {entry.solution && <span>{entry.solution}</span>}
                        </p>
                      )}

                      <div className="flex items-center gap-3 text-xs text-muted-foreground pt-1 border-t border-border -mx-4 px-4 mt-1">
                        {entry.demo_url && (
                          <a href={entry.demo_url} target="_blank" rel="noopener noreferrer" className="flex items-center gap-1 hover:text-foreground mt-2">
                            <PlayCircle className="size-3.5" /> Demo
                          </a>
                        )}
                        {entry.repo_url && (
                          <a href={entry.repo_url} target="_blank" rel="noopener noreferrer" className="flex items-center gap-1 hover:text-foreground mt-2">
                            <GitBranch className="size-3.5" /> Repo
                          </a>
                        )}
                        {entry.live_url && (
                          <a href={entry.live_url} target="_blank" rel="noopener noreferrer" className="flex items-center gap-1 hover:text-foreground mt-2">
                            <ExternalLink className="size-3.5" /> Live
                          </a>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}

            <div className="pt-4 border-t border-border">
              <p className="text-xs text-muted-foreground">
                Published via <span className="font-semibold text-foreground">Fayde</span> — productivity suite
              </p>
            </div>
          </div>
        ) : null}
      </main>
    </div>
  );
}
