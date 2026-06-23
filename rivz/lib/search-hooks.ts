"use client";
import { useQuery } from "@tanstack/react-query";
import { api } from "./api";

export type SearchResult = {
  type: "task" | "note" | "comment";
  id: string;
  title: string;
  snippet: string;
  task_id?: string;
  rank: number;
};

/** Global full-text search. Disabled until the query is non-trivial. */
export function useSearch(query: string) {
  const q = query.trim();
  return useQuery<SearchResult[]>({
    queryKey: ["search", q],
    queryFn: () => api.get<SearchResult[]>(`/search?q=${encodeURIComponent(q)}`),
    enabled: q.length >= 2,
    staleTime: 10_000,
  });
}
