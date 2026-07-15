"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type TechItem = {
  name: string;
  is_sponsor: boolean;
};

export type ShowcaseEntry = {
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

export type ShowcaseEntryPayload = {
  title: string;
  tagline: string;
  problem: string;
  solution: string;
  tech_stack: TechItem[];
  demo_url: string | null;
  repo_url: string | null;
  live_url: string | null;
};

export type ShowcaseToken = {
  id: string;
  user_id: string;
  name: string;
  token_prefix: string;
  last_used_at: string | null;
  created_at: string;
};

export type CreateShowcaseTokenResult = {
  token: string;
} & ShowcaseToken;

export type PublicShowcase = {
  owner_display_name: string;
  entries: ShowcaseEntry[];
};

function absoluteImageURL(url: string | null): string | null {
  if (!url) return null;
  return url.startsWith("http") ? url : `${BASE_URL}${url}`;
}

function withAbsoluteImages(e: ShowcaseEntry): ShowcaseEntry {
  return { ...e, logo_url: absoluteImageURL(e.logo_url), banner_url: absoluteImageURL(e.banner_url) };
}

// --- Entries ---

export function useShowcaseEntries() {
  return useQuery<ShowcaseEntry[]>({
    queryKey: ["showcase-entries"],
    queryFn: async () => (await api.get<ShowcaseEntry[]>("/showcase")).map(withAbsoluteImages),
  });
}

export function useCreateShowcaseEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: ShowcaseEntryPayload) => api.post<ShowcaseEntry>("/showcase", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["showcase-entries"] }),
  });
}

export function useUpdateShowcaseEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...body }: Partial<ShowcaseEntryPayload> & { id: string }) =>
      api.patch<ShowcaseEntry>(`/showcase/${id}`, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["showcase-entries"] }),
  });
}

export function useDeleteShowcaseEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/showcase/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["showcase-entries"] }),
  });
}

// --- Images ---
// Raw fetch (not api.ts) — multipart bodies must not be JSON-encoded, and the
// browser needs to set its own Content-Type with the multipart boundary.

async function uploadShowcaseImage(entryId: string, kind: "logo" | "banner", file: File): Promise<ShowcaseEntry> {
  const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;
  const form = new FormData();
  form.append("file", file);
  const res = await fetch(`${BASE_URL}/showcase/${entryId}/${kind}`, {
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

export function useUploadShowcaseLogo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, file }: { id: string; file: File }) => uploadShowcaseImage(id, "logo", file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["showcase-entries"] }),
  });
}

export function useUploadShowcaseBanner() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, file }: { id: string; file: File }) => uploadShowcaseImage(id, "banner", file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["showcase-entries"] }),
  });
}

export function useDeleteShowcaseLogo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/showcase/${id}/logo`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["showcase-entries"] }),
  });
}

export function useDeleteShowcaseBanner() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/showcase/${id}/banner`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["showcase-entries"] }),
  });
}

// --- Tokens ---

export function useShowcaseTokens() {
  return useQuery<ShowcaseToken[]>({
    queryKey: ["showcase-tokens"],
    queryFn: () => api.get<ShowcaseToken[]>("/settings/showcase-tokens"),
  });
}

export function useCreateShowcaseToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { name: string }) =>
      api.post<CreateShowcaseTokenResult>("/settings/showcase-tokens", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["showcase-tokens"] }),
  });
}

export function useDeleteShowcaseToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/settings/showcase-tokens/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["showcase-tokens"] }),
  });
}
