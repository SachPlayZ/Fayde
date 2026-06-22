import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type Dependency = {
  task_id: string;
  depends_on_id: string;
  title?: string;
};

export type DependencyList = {
  blocked_by: Dependency[];
  blocking: Dependency[];
};

export function useTaskDependencies(taskId: string, enabled = true) {
  return useQuery<DependencyList>({
    queryKey: ["dependencies", taskId],
    queryFn: () => api.get<DependencyList>(`/tasks/${taskId}/dependencies`),
    enabled: enabled && !!taskId,
  });
}

export function useAddDependency(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (dependsOnId: string) =>
      api.post<void>(`/tasks/${taskId}/dependencies`, { depends_on_id: dependsOnId }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["dependencies", taskId] }),
  });
}

export function useRemoveDependency(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (depId: string) =>
      api.delete<void>(`/tasks/${taskId}/dependencies/${depId}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["dependencies", taskId] }),
  });
}
