"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type Trigger = { event: string; to?: string };
export type Condition = { field: string; op: string; value: string };
export type Action = { type: string; value: string; kind?: string };

export type Automation = {
  id: string;
  name: string;
  enabled: boolean;
  trigger: Trigger;
  conditions: Condition[];
  actions: Action[];
  created_at: string;
};

export function useAutomations() {
  return useQuery<Automation[]>({
    queryKey: ["automations"],
    queryFn: () => api.get<Automation[]>("/automations"),
  });
}

export function useCreateAutomation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: {
      name: string;
      trigger: Trigger;
      conditions: Condition[];
      actions: Action[];
    }) => api.post<Automation>("/automations", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["automations"] }),
  });
}

export function useUpdateAutomation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: Partial<Automation> }) =>
      api.patch<Automation>(`/automations/${id}`, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["automations"] }),
  });
}

export function useDeleteAutomation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/automations/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["automations"] }),
  });
}
