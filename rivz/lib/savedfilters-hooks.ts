import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type SavedFilter = {
  id: string;
  user_id: string;
  name: string;
  params: Record<string, string>;
  created_at: string;
};

export function useSavedFilters() {
  return useQuery<SavedFilter[]>({
    queryKey: ["saved-filters"],
    queryFn: () => api.get<SavedFilter[]>("/saved-filters"),
  });
}

export function useCreateSavedFilter() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { name: string; params: Record<string, string> }) =>
      api.post<SavedFilter>("/saved-filters", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["saved-filters"] }),
  });
}

export function useDeleteSavedFilter() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/saved-filters/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["saved-filters"] }),
  });
}
