"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type TechItem = {
  name: string;
  is_sponsor: boolean;
};

export type Project = {
  id: string;
  user_id: string;
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
  sort_order: number;
  created_at: string;
  updated_at: string;
};

export type ProjectPayload = {
  title: string;
  tagline: string;
  problem: string;
  solution: string;
  tech_stack: TechItem[];
  demo_url: string | null;
  repo_url: string | null;
  live_url: string | null;
};

export type PublicProjects = {
  owner_display_name: string;
  projects: Project[];
};

function absoluteImageURL(url: string | null): string | null {
  if (!url) return null;
  return url.startsWith("http") ? url : `${BASE_URL}${url}`;
}

function withAbsoluteImages(e: Project): Project {
  return { ...e, logo_url: absoluteImageURL(e.logo_url), banner_url: absoluteImageURL(e.banner_url) };
}

// --- Entries ---

export function useProjects() {
  return useQuery<Project[]>({
    queryKey: ["projects"],
    queryFn: async () => (await api.get<Project[]>("/projects")).map(withAbsoluteImages),
  });
}

export function useCreateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: ProjectPayload) => api.post<Project>("/projects", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}

export function useUpdateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...body }: Partial<ProjectPayload> & { id: string }) =>
      api.patch<Project>(`/projects/${id}`, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}

export function useDeleteProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/projects/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}

// --- Images ---
// Raw fetch (not api.ts) — multipart bodies must not be JSON-encoded, and the
// browser needs to set its own Content-Type with the multipart boundary.

async function uploadProjectImage(entryId: string, kind: "logo" | "banner", file: File): Promise<Project> {
  const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
  const form = new FormData();
  form.append("file", file);
  const res = await fetch(`${BASE_URL}/projects/${entryId}/${kind}`, {
    method: "POST",
    headers: token ? { Authorization: `Bearer ${token}` } : {},
    body: form,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body?.error?.message ?? "Upload failed");
  }
  return res.json();
}

export function useUploadProjectLogo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, file }: { id: string; file: File }) => uploadProjectImage(id, "logo", file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}

export function useUploadProjectBanner() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, file }: { id: string; file: File }) => uploadProjectImage(id, "banner", file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}

export function useDeleteProjectLogo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/projects/${id}/logo`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}

export function useDeleteProjectBanner() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/projects/${id}/banner`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}
