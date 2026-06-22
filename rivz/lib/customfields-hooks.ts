import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type FieldDefinition = {
  id: string;
  user_id: string;
  name: string;
  field_type: "text" | "number" | "date" | "select";
  options: string[] | null;
  created_at: string;
};

export type FieldValue = {
  id: string;
  task_id: string;
  field_id: string;
  name: string;
  value: string;
};

export function useCustomFieldDefs() {
  return useQuery<FieldDefinition[]>({
    queryKey: ["custom-fields"],
    queryFn: () => api.get<FieldDefinition[]>("/custom-fields"),
  });
}

export function useCreateFieldDef() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: {
      name: string;
      field_type: FieldDefinition["field_type"];
      options?: string[];
    }) => api.post<FieldDefinition>("/custom-fields", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["custom-fields"] }),
  });
}

export function useDeleteFieldDef() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/custom-fields/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["custom-fields"] }),
  });
}

export function useTaskFieldValues(taskId: string, enabled?: boolean) {
  return useQuery<FieldValue[]>({
    queryKey: ["custom-fields", "values", taskId],
    queryFn: () => api.get<FieldValue[]>(`/tasks/${taskId}/custom-fields`),
    enabled: !!taskId && (enabled ?? true),
  });
}

export function useSetFieldValue() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      taskId,
      fieldId,
      value,
    }: {
      taskId: string;
      fieldId: string;
      value: string;
    }) =>
      api.put<FieldValue>(`/tasks/${taskId}/custom-fields/${fieldId}`, {
        value,
      }),
    onSuccess: (_data, { taskId }) =>
      qc.invalidateQueries({ queryKey: ["custom-fields", "values", taskId] }),
  });
}
