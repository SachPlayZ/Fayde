import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type Subtask = {
  id: string;
  task_id: string;
  title: string;
  done: boolean;
  position: number;
  created_at: string;
};

export function useSubtasks(taskId: string, enabled = true) {
  return useQuery<Subtask[]>({
    queryKey: ["subtasks", taskId],
    queryFn: () => api.get<Subtask[]>(`/tasks/${taskId}/subtasks`),
    enabled: enabled && !!taskId,
  });
}

export function useCreateSubtask(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (title: string) => api.post<Subtask>(`/tasks/${taskId}/subtasks`, { title }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["subtasks", taskId] }),
  });
}

export function useUpdateSubtask(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, done, title }: { id: string; done: boolean; title: string }) =>
      api.patch<Subtask>(`/tasks/${taskId}/subtasks/${id}`, { done, title }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["subtasks", taskId] }),
  });
}

export function useDeleteSubtask(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/tasks/${taskId}/subtasks/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["subtasks", taskId] }),
  });
}

export function useReorderSubtasks(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (ids: string[]) =>
      api.put<void>(`/tasks/${taskId}/subtasks/order`, { ids }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["subtasks", taskId] }),
  });
}
