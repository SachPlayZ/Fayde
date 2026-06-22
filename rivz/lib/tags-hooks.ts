import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type Tag = {
  id: string;
  user_id: string;
  name: string;
  color: string;
};

export function useTags() {
  return useQuery<Tag[]>({
    queryKey: ["tags"],
    queryFn: () => api.get<Tag[]>("/tags"),
  });
}

export function useCreateTag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; color: string }) => api.post<Tag>("/tags", data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tags"] }),
  });
}

export function useDeleteTag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/tags/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tags"] }),
  });
}

export function useAddTagToTask(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (tagId: string) => api.post<void>(`/tasks/${taskId}/tags`, { tag_id: tagId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tasks"] });
      qc.invalidateQueries({ queryKey: ["tags", "task", taskId] });
    },
  });
}

export function useRemoveTagFromTask(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (tagId: string) => api.delete<void>(`/tasks/${taskId}/tags/${tagId}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tasks"] });
      qc.invalidateQueries({ queryKey: ["tags", "task", taskId] });
    },
  });
}
