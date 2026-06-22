import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type OutboundWebhook = {
  id: string;
  user_id: string;
  name: string;
  url: string;
  events: string[];
  secret: string;
  enabled: boolean;
  created_at: string;
};

export type Webhook = OutboundWebhook;

export function useWebhooks() {
  return useQuery<Webhook[]>({
    queryKey: ["webhooks"],
    queryFn: () => api.get<Webhook[]>("/settings/webhooks"),
  });
}

export function useCreateWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: {
      name: string;
      url: string;
      events: string[];
      secret?: string;
    }) => api.post<Webhook>("/settings/webhooks", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["webhooks"] }),
  });
}

export function useUpdateWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...body }: { id: string } & Partial<Webhook>) =>
      api.patch<Webhook>(`/settings/webhooks/${id}`, body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["webhooks"] }),
  });
}

export function useDeleteWebhook() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/settings/webhooks/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["webhooks"] }),
  });
}
