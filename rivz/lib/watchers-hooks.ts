import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type Watcher = {
  task_id: string;
  user_id: string;
  user_email: string;
};

export function useWatchers(taskId: string, enabled?: boolean) {
  return useQuery<Watcher[]>({
    queryKey: ["watchers", taskId],
    queryFn: () => api.get<Watcher[]>(`/tasks/${taskId}/watchers`),
    enabled: !!taskId && (enabled ?? true),
  });
}

export function useWatchStatus(taskId: string, enabled?: boolean) {
  return useQuery<{ watching: boolean }>({
    queryKey: ["watchers", "status", taskId],
    queryFn: () =>
      api.get<{ watching: boolean }>(`/tasks/${taskId}/watchers/status`),
    enabled: !!taskId && (enabled ?? true),
  });
}

export function useAddWatcher() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (taskId: string) =>
      api.post<void>(`/tasks/${taskId}/watchers`, {}),
    onSuccess: (_data, taskId) => {
      qc.invalidateQueries({ queryKey: ["watchers", taskId] });
      qc.invalidateQueries({ queryKey: ["watchers", "status", taskId] });
    },
  });
}

export function useRemoveWatcher() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (taskId: string) =>
      api.delete<void>(`/tasks/${taskId}/watchers`),
    onSuccess: (_data, taskId) => {
      qc.invalidateQueries({ queryKey: ["watchers", taskId] });
      qc.invalidateQueries({ queryKey: ["watchers", "status", taskId] });
    },
  });
}
