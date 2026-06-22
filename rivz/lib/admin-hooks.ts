import { useQuery } from "@tanstack/react-query";
import { api } from "./api";
import type { Tag } from "./tags-hooks";

export type AdminTask = {
  id: string;
  user_id: string;
  user_email: string;
  title: string;
  description: string;
  status: "todo" | "in_progress" | "done" | "failed";
  priority: "low" | "medium" | "high";
  due_date: string | null;
  recurrence: string | null;
  recurrence_end: string | null;
  assignee_id: string | null;
  assignee_email: string | null;
  sort_order: number;
  tags: Tag[];
  subtask_count: number;
  subtasks_done: number;
  created_at: string;
  updated_at: string;
};

export type AdminTasksResponse = {
  data: AdminTask[];
  page: number;
  limit: number;
  total: number;
};

export type AdminUser = {
  id: string;
  email: string;
  role: string;
  created_at: string;
  task_count: number;
};

export function useAdminTasks(params?: {
  status?: string;
  search?: string;
  sort?: string;
  order?: string;
  page?: number;
  limit?: number;
}) {
  const q = new URLSearchParams();
  if (params?.status) q.set("status", params.status);
  if (params?.search) q.set("search", params.search);
  if (params?.sort) q.set("sort", params.sort);
  if (params?.order) q.set("order", params.order);
  if (params?.page) q.set("page", String(params.page));
  if (params?.limit) q.set("limit", String(params.limit));

  const qs = q.toString();
  return useQuery<AdminTasksResponse>({
    queryKey: ["admin", "tasks", qs],
    queryFn: () => api.get<AdminTasksResponse>(`/admin/tasks${qs ? `?${qs}` : ""}`),
  });
}

export function useAdminUsers(enabled = true) {
  return useQuery<AdminUser[]>({
    queryKey: ["admin", "users"],
    queryFn: () => api.get<AdminUser[]>("/admin/users"),
    enabled,
  });
}
