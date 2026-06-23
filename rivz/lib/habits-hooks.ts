"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { format } from "date-fns";
import { api } from "./api";
import type { DashboardSummary } from "./dashboard-hooks";

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
    onMutate: async ({ id, date }) => {
      const targetDate = date || format(new Date(), "yyyy-MM-dd");

      // Cancel outgoing refetches so they don't overwrite our optimistic update
      await qc.cancelQueries({ queryKey: ["habit-logs", id] });
      await qc.cancelQueries({ queryKey: ["habits"] });
      await qc.cancelQueries({ queryKey: ["dashboard"] });

      // Snapshot the previous values
      const previousLogs = qc.getQueriesData<HabitLog[]>({ queryKey: ["habit-logs", id] });
      const previousHabits = qc.getQueryData<Habit[]>(["habits"]);
      const previousDashboard = qc.getQueryData<DashboardSummary>(["dashboard"]);

      // Optimistically update the logs queries
      qc.setQueriesData<HabitLog[]>({ queryKey: ["habit-logs", id] }, (old) => {
        const currentLogs = old || [];
        const exists = currentLogs.some((log) => log.date === targetDate);
        if (exists) {
          return currentLogs.filter((log) => log.date !== targetDate);
        } else {
          return [...currentLogs, { date: targetDate, count: 1 }];
        }
      });

      // Optimistically update the habits list
      if (previousHabits) {
        qc.setQueryData<Habit[]>(["habits"], (old) => {
          if (!old) return [];
          return old.map((h) => {
            if (h.id !== id) return h;
            
            const isTodayDate = targetDate === format(new Date(), "yyyy-MM-dd");
            const wasDoneToday = h.done_today;
            const newDoneToday = isTodayDate ? !wasDoneToday : wasDoneToday;
            
            let newCurrentStreak = h.current_streak;
            if (isTodayDate) {
              if (newDoneToday) {
                newCurrentStreak = h.current_streak + 1;
              } else {
                newCurrentStreak = Math.max(0, h.current_streak - 1);
              }
            }

            return {
              ...h,
              done_today: newDoneToday,
              current_streak: newCurrentStreak,
              longest_streak: Math.max(h.longest_streak, newCurrentStreak),
            };
          });
        });
      }

      // Optimistically update the dashboard habits list
      if (previousDashboard) {
        qc.setQueryData<DashboardSummary>(["dashboard"], (old) => {
          if (!old || !old.habits) return old;
          return {
            ...old,
            habits: old.habits.map((h: Habit) => {
              if (h.id !== id) return h;
              
              const isTodayDate = targetDate === format(new Date(), "yyyy-MM-dd");
              const wasDoneToday = h.done_today;
              const newDoneToday = isTodayDate ? !wasDoneToday : wasDoneToday;

              let newCurrentStreak = h.current_streak;
              if (isTodayDate) {
                if (newDoneToday) {
                  newCurrentStreak = h.current_streak + 1;
                } else {
                  newCurrentStreak = Math.max(0, h.current_streak - 1);
                }
              }

              return {
                ...h,
                done_today: newDoneToday,
                current_streak: newCurrentStreak,
                longest_streak: Math.max(h.longest_streak, newCurrentStreak),
              };
            }),
          };
        });
      }

      return { previousLogs, previousHabits, previousDashboard };
    },
    onError: (err, variables, context) => {
      // Revert to snapshot
      if (context?.previousLogs) {
        context.previousLogs.forEach(([queryKey, value]) => {
          qc.setQueryData(queryKey, value);
        });
      }
      if (context?.previousHabits) {
        qc.setQueryData(["habits"], context.previousHabits);
      }
      if (context?.previousDashboard) {
        qc.setQueryData(["dashboard"], context.previousDashboard);
      }
    },
    onSettled: (data, error, variables) => {
      // Invalidate to sync with DB
      qc.invalidateQueries({ queryKey: ["habit-logs", variables.id] });
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
