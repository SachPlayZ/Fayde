import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type TOTPSetup = {
  secret: string;
  qr_url: string;
};

export function useTOTPStatus() {
  return useQuery<{ enabled: boolean }>({
    queryKey: ["totp-status"],
    queryFn: () => api.get<{ enabled: boolean }>("/auth/totp/status"),
  });
}

export function useSetupTOTP() {
  return useMutation({
    mutationFn: () => api.post<TOTPSetup>("/auth/totp/setup", {}),
  });
}

export function useEnableTOTP() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { code: string }) =>
      api.post<void>("/auth/totp/enable", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["totp-status"] }),
  });
}

export function useDisableTOTP() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { code: string }) =>
      api.post<void>("/auth/totp/disable", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["totp-status"] }),
  });
}
