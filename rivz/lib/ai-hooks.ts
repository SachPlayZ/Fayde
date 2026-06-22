import { useQuery, useMutation } from "@tanstack/react-query";
import { api } from "./api";

export function useParseTask() {
  return useMutation({
    mutationFn: (body: { text: string }) =>
      api.post<{
        title: string;
        description: string;
        priority: string;
        tags: string[];
        due_date: string | null;
      }>("/ai/parse-task", body),
  });
}

export function useBreakdownTask() {
  return useMutation({
    mutationFn: (body: { title: string; description: string }) =>
      api.post<{ subtasks: { title: string }[] }>("/ai/breakdown", body),
  });
}

export function useSuggestTags() {
  return useMutation({
    mutationFn: (body: { title: string; description: string }) =>
      api.post<{ tags: string[] }>("/ai/suggest-tags", body),
  });
}

export function useSuggestPriority() {
  return useMutation({
    mutationFn: (body: { title: string; description: string }) =>
      api.post<{ priority: string; reasoning: string }>(
        "/ai/suggest-priority",
        body
      ),
  });
}

export function useExpandDescription() {
  return useMutation({
    mutationFn: (body: { title: string; bullets: string }) =>
      api.post<{ description: string }>("/ai/expand-description", body),
  });
}

export function useEstimateTime() {
  return useMutation({
    mutationFn: (body: { title: string; description: string }) =>
      api.post<{ estimate_seconds: number; reasoning: string }>(
        "/ai/estimate-time",
        body
      ),
  });
}

export function useWeeklyDigest() {
  return useQuery<{ digest: string }>({
    queryKey: ["ai", "weekly-digest"],
    queryFn: () => api.get<{ digest: string }>("/ai/weekly-digest"),
    staleTime: 5 * 60 * 1000,
  });
}

export function useLoadAlert() {
  return useQuery<{ overloaded: boolean; message: string }>({
    queryKey: ["ai", "load-alert"],
    queryFn: () =>
      api.get<{ overloaded: boolean; message: string }>("/ai/load-alert"),
    staleTime: 5 * 60 * 1000,
  });
}
