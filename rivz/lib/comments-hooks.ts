import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type Comment = {
  id: string;
  task_id: string;
  user_id: string;
  user_email: string;
  body: string;
  created_at: string;
  updated_at: string;
};

export function useComments(taskId: string, enabled = true) {
  return useQuery<Comment[]>({
    queryKey: ["comments", taskId],
    queryFn: () => api.get<Comment[]>(`/tasks/${taskId}/comments`),
    enabled: enabled && !!taskId,
  });
}

export function useCreateComment(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: string) => api.post<Comment>(`/tasks/${taskId}/comments`, { body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["comments", taskId] }),
  });
}

export function useUpdateComment(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: string }) =>
      api.patch<Comment>(`/tasks/${taskId}/comments/${id}`, { body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["comments", taskId] }),
  });
}

export function useDeleteComment(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/tasks/${taskId}/comments/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["comments", taskId] }),
  });
}
