import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type GitHubLink = {
  id: string;
  task_id: string;
  repo: string;
  issue_number: number | null;
  pr_number: number | null;
  issue_url: string | null;
  pr_url: string | null;
  created_at: string;
};

export function useGitHubLinks(taskId: string, enabled?: boolean) {
  return useQuery<GitHubLink[]>({
    queryKey: ["github", taskId],
    queryFn: () => api.get<GitHubLink[]>(`/tasks/${taskId}/github`),
    enabled: !!taskId && (enabled ?? true),
  });
}

export function useAddGitHubLink() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      taskId,
      ...data
    }: {
      taskId: string;
      repo: string;
      issue_number?: number;
      pr_number?: number;
      issue_url?: string;
      pr_url?: string;
    }) => api.post<GitHubLink>(`/tasks/${taskId}/github`, data),
    onSuccess: (_data, { taskId }) =>
      qc.invalidateQueries({ queryKey: ["github", taskId] }),
  });
}

export function useDeleteGitHubLink() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, linkId }: { taskId: string; linkId: string }) =>
      api.delete<void>(`/tasks/${taskId}/github/${linkId}`),
    onSuccess: (_data, { taskId }) =>
      qc.invalidateQueries({ queryKey: ["github", taskId] }),
  });
}
