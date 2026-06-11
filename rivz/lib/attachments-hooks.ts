import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type Attachment = {
  id: string;
  task_id: string;
  user_id: string;
  filename: string;
  content_type: string;
  size_bytes: number;
  url: string;
  created_at: string;
};

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

/**
 * useAttachments fetches all attachments for a given task.
 */
export function useAttachments(taskId: string, enabled = true) {
  return useQuery<Attachment[]>({
    queryKey: ["attachments", taskId],
    queryFn: () => api.get<Attachment[]>(`/tasks/${taskId}/attachments`),
    enabled: !!taskId && enabled,
  });
}

/**
 * useUploadAttachment uploads a file to a task as a multipart/form-data POST.
 */
export function useUploadAttachment(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (file: File) => {
      const token =
        typeof window !== "undefined" ? localStorage.getItem("token") : null;
      const form = new FormData();
      form.append("file", file);

      const res = await fetch(`${BASE_URL}/tasks/${taskId}/attachments`, {
        method: "POST",
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        body: form,
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body?.error?.message ?? "Upload failed");
      }

      return res.json() as Promise<Attachment>;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["attachments", taskId] });
    },
  });
}

/**
 * useDeleteAttachment removes an attachment from a task.
 */
export function useDeleteAttachment(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (attId: string) =>
      api.delete<void>(`/tasks/${taskId}/attachments/${attId}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["attachments", taskId] });
    },
  });
}
