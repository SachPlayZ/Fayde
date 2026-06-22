import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type TimeEntry = {
  id: string;
  task_id: string;
  user_id: string;
  started_at: string;
  ended_at: string | null;
  duration_seconds: number | null;
  note: string;
  created_at: string;
};

export function useTimeEntries(taskId: string, enabled?: boolean) {
  return useQuery<TimeEntry[]>({
    queryKey: ["time", taskId],
    queryFn: () => api.get<TimeEntry[]>(`/tasks/${taskId}/time`),
    enabled: !!taskId && (enabled ?? true),
  });
}

export function useActiveTimeEntry(taskId: string, enabled?: boolean) {
  return useQuery<TimeEntry | null>({
    queryKey: ["time", taskId, "active"],
    queryFn: async () => {
      try {
        return await api.get<TimeEntry>(`/tasks/${taskId}/time/active`);
      } catch {
        return null;
      }
    },
    enabled: !!taskId && (enabled ?? true),
  });
}

export function useStartTimer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, note }: { taskId: string; note?: string }) =>
      api.post<TimeEntry>(`/tasks/${taskId}/time/start`, { note }),
    onSuccess: (_data, { taskId }) =>
      qc.invalidateQueries({ queryKey: ["time", taskId] }),
  });
}

export function useStopTimer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      taskId,
      entryId,
      note,
    }: {
      taskId: string;
      entryId: string;
      note?: string;
    }) => api.post<TimeEntry>(`/tasks/${taskId}/time/stop/${entryId}`, { note }),
    onSuccess: (_data, { taskId }) =>
      qc.invalidateQueries({ queryKey: ["time", taskId] }),
  });
}

export function useDeleteTimeEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, entryId }: { taskId: string; entryId: string }) =>
      api.delete<void>(`/tasks/${taskId}/time/${entryId}`),
    onSuccess: (_data, { taskId }) =>
      qc.invalidateQueries({ queryKey: ["time", taskId] }),
  });
}
