"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type CardState = {
  id: string;
  problem_id: string;
  user_id: string;
  stability: number;
  difficulty: number;
  elapsed_days: number;
  scheduled_days: number;
  reps: number;
  lapses: number;
  card_state: "new" | "learning" | "review" | "relearning";
  due_date: string;
  last_review: string | null;
};

export type Problem = {
  id: string;
  user_id: string;
  lc_number: number | null;
  slug: string | null;
  title: string;
  url: string | null;
  difficulty: "easy" | "medium" | "hard";
  topics: string[];
  notes: string;
  solved_at: string | null;
  created_at: string;
  updated_at: string;
  card: CardState | null;
};

export type ReviewLog = {
  id: string;
  card_id: string;
  user_id: string;
  reviewed_at: string;
  rating: 1 | 2 | 3 | 4;
  scheduled_days: number;
  elapsed_days: number;
  stability: number;
  difficulty: number;
  card_state: string;
};

export type ProblemDetail = Problem & { reviews: ReviewLog[] };

export type LeetcodeStats = {
  total_problems: number;
  total_solved: number;
  due_today_count: number;
  overdue_count: number;
  total_reviews: number;
  review_streak: number;
  retention_rate: number;
};

export type ProblemFilter = {
  topic?: string;
  difficulty?: string;
  status?: "due" | "overdue" | "";
};

export type CreateProblemBody = {
  title: string;
  lc_number?: number | null;
  slug?: string | null;
  url?: string | null;
  difficulty?: string;
  topics?: string[];
  notes?: string;
};

function queryString(filter?: ProblemFilter) {
  if (!filter) return "";
  const params = new URLSearchParams();
  if (filter.topic) params.set("topic", filter.topic);
  if (filter.difficulty) params.set("difficulty", filter.difficulty);
  if (filter.status) params.set("status", filter.status);
  const qs = params.toString();
  return qs ? `?${qs}` : "";
}

export function useProblems(filter?: ProblemFilter) {
  return useQuery<Problem[]>({
    queryKey: ["leetcode-problems", filter],
    queryFn: () => api.get<Problem[]>(`/leetcode/problems${queryString(filter)}`),
  });
}

export function useProblem(id: string | null) {
  return useQuery<ProblemDetail>({
    queryKey: ["leetcode-problem", id],
    queryFn: () => api.get<ProblemDetail>(`/leetcode/problems/${id}`),
    enabled: !!id,
  });
}

export function useCreateProblem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateProblemBody) => api.post<Problem>("/leetcode/problems", body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["leetcode-problems"] });
      qc.invalidateQueries({ queryKey: ["leetcode-stats"] });
    },
  });
}

export function useUpdateProblem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: Partial<CreateProblemBody> }) =>
      api.patch<Problem>(`/leetcode/problems/${id}`, patch),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ["leetcode-problems"] });
      qc.invalidateQueries({ queryKey: ["leetcode-problem", variables.id] });
    },
  });
}

export function useDeleteProblem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/leetcode/problems/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["leetcode-problems"] });
      qc.invalidateQueries({ queryKey: ["leetcode-stats"] });
    },
  });
}

export function useReviewQueue() {
  return useQuery<Problem[]>({
    queryKey: ["leetcode-queue"],
    queryFn: () => api.get<Problem[]>("/leetcode/queue"),
  });
}

export function useReviewProblem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, rating }: { id: string; rating: 1 | 2 | 3 | 4 }) =>
      api.post<CardState>(`/leetcode/problems/${id}/review`, { rating }),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ["leetcode-problems"] });
      qc.invalidateQueries({ queryKey: ["leetcode-problem", variables.id] });
      qc.invalidateQueries({ queryKey: ["leetcode-queue"] });
      qc.invalidateQueries({ queryKey: ["leetcode-stats"] });
      qc.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export function useLeetcodeStats() {
  return useQuery<LeetcodeStats>({
    queryKey: ["leetcode-stats"],
    queryFn: () => api.get<LeetcodeStats>("/leetcode/stats"),
  });
}
