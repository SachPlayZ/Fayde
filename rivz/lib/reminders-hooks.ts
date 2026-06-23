"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type Reminder = {
  id: string;
  task_id: string | null;
  remind_at: string;
  note: string;
  sent: boolean;
  created_at: string;
};

export function useReminders(taskId: string, enabled = true) {
  return useQuery<Reminder[]>({
    queryKey: ["reminders", taskId],
    queryFn: () => api.get<Reminder[]>(`/tasks/${taskId}/reminders`),
    enabled: enabled && !!taskId,
  });
}

export function useCreateReminder(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { remind_at: string; note?: string }) =>
      api.post<Reminder>(`/tasks/${taskId}/reminders`, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["reminders", taskId] }),
  });
}

export function useDeleteReminder(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/reminders/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["reminders", taskId] }),
  });
}
