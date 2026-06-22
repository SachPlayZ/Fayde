import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type APIToken = {
  id: string;
  user_id: string;
  name: string;
  token_prefix: string;
  last_used_at: string | null;
  created_at: string;
};

export type CreateTokenResult = {
  token: string;
} & APIToken;

export function useAPITokens() {
  return useQuery<APIToken[]>({
    queryKey: ["api-tokens"],
    queryFn: () => api.get<APIToken[]>("/settings/api-tokens"),
  });
}

export function useCreateAPIToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { name: string }) =>
      api.post<CreateTokenResult>("/settings/api-tokens", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["api-tokens"] }),
  });
}

export function useDeleteAPIToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api.delete<void>(`/settings/api-tokens/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["api-tokens"] }),
  });
}
