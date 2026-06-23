"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type Habit = {
  id: string;
  name: string;
  cadence: string;
  target_per_period: number;
  color: string | null;
  position: number;
  archived: boolean;
  created_at: string;
  current_streak: number;
  longest_streak: number;
  done_today: boolean;
};

export type HabitLog = { date: string; count: number };

export function useHabits() {
  return useQuery<Habit[]>({
    queryKey: ["habits"],
    queryFn: () => api.get<Habit[]>("/habits"),
  });
}

export function useCreateHabit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { name: string; cadence?: string; color?: string }) =>
      api.post<Habit>("/habits", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["habits"] }),
  });
}

export function useUpdateHabit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: Partial<Habit> }) =>
      api.patch<Habit>(`/habits/${id}`, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["habits"] }),
  });
}

export function useDeleteHabit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/habits/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["habits"] }),
  });
}

export function useToggleHabit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, date }: { id: string; date?: string }) =>
      api.post<{ done: boolean }>(`/habits/${id}/toggle`, date ? { date } : {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["habits"] });
      qc.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });
}

export function useHabitLogs(id: string | null, from: string, to: string) {
  return useQuery<HabitLog[]>({
    queryKey: ["habit-logs", id, from, to],
    queryFn: () => api.get<HabitLog[]>(`/habits/${id}/logs?from=${from}&to=${to}`),
    enabled: !!id,
  });
}
