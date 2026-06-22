import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type TaskTemplate = {
  id: string;
  user_id: string;
  name: string;
  title: string;
  description: string;
  status: string;
  priority: string;
  effort_points: number | null;
  created_at: string;
};

export function useTemplates() {
  return useQuery<TaskTemplate[]>({
    queryKey: ["templates"],
    queryFn: () => api.get<TaskTemplate[]>("/templates"),
  });
}

export function useCreateTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: Partial<TaskTemplate>) =>
      api.post<TaskTemplate>("/templates", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["templates"] }),
  });
}

export function useDeleteTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/templates/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["templates"] }),
  });
}
