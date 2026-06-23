import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { api } from "./api";
import { useSSE } from "./sse-hook";

export type Notification = {
  id: string;
  user_id: string;
  type: string;
  task_id: string | null;
  message: string;
  read: boolean;
  created_at: string;
};

export function useNotifications(unreadOnly = false) {
  const qc = useQueryClient();
  const queryKey = ["notifications", unreadOnly];

  const query = useQuery<Notification[]>({
    queryKey,
    queryFn: () => api.get<Notification[]>(`/notifications${unreadOnly ? "?unread=true" : ""}`),
  });

  // Invalidate on SSE notification events.
  const { lastEvent } = useSSE();
  useEffect(() => {
    if (lastEvent?.type === "notification") {
      qc.invalidateQueries({ queryKey: ["notifications"] });
      qc.invalidateQueries({ queryKey: ["notifications", "unread-count"] });
    }
  }, [lastEvent, qc]);

  return query;
}

export function useUnreadCount() {
  const qc = useQueryClient();
  const query = useQuery<{ count: number }>({
    queryKey: ["notifications", "unread-count"],
    queryFn: () => api.get<{ count: number }>("/notifications/unread-count"),
  });

  const { lastEvent } = useSSE();
  useEffect(() => {
    if (lastEvent?.type === "notification") {
      qc.invalidateQueries({ queryKey: ["notifications", "unread-count"] });
    }
  }, [lastEvent, qc]);

  return query;
}

export function useMarkRead() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.patch<void>(`/notifications/${id}/read`, {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["notifications"] });
      qc.invalidateQueries({ queryKey: ["notifications", "unread-count"] });
    },
  });
}

export function useSnoozeNotification() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, until }: { id: string; until: string }) =>
      api.post<void>(`/notifications/${id}/snooze`, { until }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["notifications"] });
      qc.invalidateQueries({ queryKey: ["notifications", "unread-count"] });
    },
  });
}

export function useMarkAllRead() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api.post<void>("/notifications/read-all", {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["notifications"] });
      qc.invalidateQueries({ queryKey: ["notifications", "unread-count"] });
    },
  });
}
