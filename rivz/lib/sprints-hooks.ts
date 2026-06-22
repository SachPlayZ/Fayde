import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type Sprint = {
  id: string;
  user_id: string;
  name: string;
  start_date: string;
  end_date: string;
  goal: string;
  created_at: string;
};

export function useSprints() {
  return useQuery<Sprint[]>({
    queryKey: ["sprints"],
    queryFn: () => api.get<Sprint[]>("/sprints"),
  });
}

export function useCreateSprint() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: {
      name: string;
      start_date: string;
      end_date: string;
      goal: string;
    }) => api.post<Sprint>("/sprints", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sprints"] }),
  });
}

export function useUpdateSprint() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...body }: { id: string } & Partial<Sprint>) =>
      api.patch<Sprint>(`/sprints/${id}`, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sprints"] }),
  });
}

export function useDeleteSprint() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/sprints/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sprints"] }),
  });
}

export function useSprintTaskIDs(sprintId: string) {
  return useQuery<string[]>({
    queryKey: ["sprints", sprintId, "tasks"],
    queryFn: () => api.get<string[]>(`/sprints/${sprintId}/tasks`),
    enabled: !!sprintId,
  });
}

export function useAddTaskToSprint() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ sprintId, task_id }: { sprintId: string; task_id: string }) =>
      api.post<void>(`/sprints/${sprintId}/tasks`, { task_id }),
    onSuccess: (_data, { sprintId }) =>
      qc.invalidateQueries({ queryKey: ["sprints", sprintId, "tasks"] }),
  });
}

export function useRemoveTaskFromSprint() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ sprintId, taskId }: { sprintId: string; taskId: string }) =>
      api.delete<void>(`/sprints/${sprintId}/tasks/${taskId}`),
    onSuccess: (_data, { sprintId }) =>
      qc.invalidateQueries({ queryKey: ["sprints", sprintId, "tasks"] }),
  });
}
