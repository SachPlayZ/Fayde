"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type KeyResult = {
  id: string;
  goal_id: string;
  title: string;
  metric_type: string;
  current_val: number;
  target_val: number;
  position: number;
};

export type Goal = {
  id: string;
  user_id: string;
  title: string;
  description: string;
  status: string;
  target_date: string | null;
  parent_id: string | null;
  position: number;
  created_at: string;
  updated_at: string;
  key_results: KeyResult[];
  progress: number;
};

export type LinkedTask = { id: string; title: string; status: string };

export function useGoals() {
  return useQuery<Goal[]>({ queryKey: ["goals"], queryFn: () => api.get<Goal[]>("/goals") });
}

export function useCreateGoal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { title: string; description?: string; target_date?: string }) =>
      api.post<Goal>("/goals", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["goals"] }),
  });
}

export function useUpdateGoal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: Partial<Goal> }) =>
      api.patch<Goal>(`/goals/${id}`, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["goals"] }),
  });
}

export function useDeleteGoal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/goals/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["goals"] }),
  });
}

export function useAddKR(goalId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { title: string; metric_type?: string; target_val?: number }) =>
      api.post<KeyResult>(`/goals/${goalId}/key-results`, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["goals"] }),
  });
}

export function useUpdateKR(goalId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ krId, patch }: { krId: string; patch: Partial<KeyResult> }) =>
      api.patch<KeyResult>(`/goals/${goalId}/key-results/${krId}`, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["goals"] }),
  });
}

export function useDeleteKR(goalId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (krId: string) => api.delete<void>(`/goals/${goalId}/key-results/${krId}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["goals"] }),
  });
}

export function useGoalTasks(goalId: string | null) {
  return useQuery<LinkedTask[]>({
    queryKey: ["goal-tasks", goalId],
    queryFn: () => api.get<LinkedTask[]>(`/goals/${goalId}/tasks`),
    enabled: !!goalId,
  });
}
