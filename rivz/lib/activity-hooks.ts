import { useQuery } from "@tanstack/react-query";
import { api } from "./api";

export type ActivityLog = {
  id: string;
  task_id: string;
  user_id: string;
  action: string;
  changes: Record<string, unknown> | null;
  created_at: string;
};

/**
 * useTaskActivity fetches the activity log for a given task.
 * Only called when taskId is non-empty and the dialog is open.
 */
export function useTaskActivity(taskId: string, enabled = true) {
  return useQuery<ActivityLog[]>({
    queryKey: ["activity", taskId],
    queryFn: () => api.get<ActivityLog[]>(`/tasks/${taskId}/activity`),
    enabled: !!taskId && enabled,
  });
}
