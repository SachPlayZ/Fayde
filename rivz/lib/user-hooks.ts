import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export function useUpdatePreferences() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (prefs: { theme?: string; digest_enabled?: boolean }) =>
      api.patch<void>("/auth/me/preferences", prefs),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["me"] }),
  });
}
