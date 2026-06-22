import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type ShareToken = {
  id: string;
  task_id: string;
  token: string;
  url: string;
  created_at: string;
};

export function useShareToken(taskId: string, enabled?: boolean) {
  return useQuery<ShareToken | null>({
    queryKey: ["share", taskId],
    queryFn: async () => {
      try {
        return await api.get<ShareToken>(`/tasks/${taskId}/share`);
      } catch {
        return null;
      }
    },
    enabled: !!taskId && (enabled ?? true),
  });
}

export function useCreateShareToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (taskId: string) =>
      api.post<{ token: string; url: string }>(`/tasks/${taskId}/share`, {}),
    onSuccess: (_data, taskId) =>
      qc.invalidateQueries({ queryKey: ["share", taskId] }),
  });
}

export function useDeleteShareToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (taskId: string) =>
      api.delete<void>(`/tasks/${taskId}/share`),
    onSuccess: (_data, taskId) =>
      qc.invalidateQueries({ queryKey: ["share", taskId] }),
  });
}
