import { useMutation } from "@tanstack/react-query";
import { api } from "./api";

type AuthResult = { token: string; user: { id: string; email: string; role: string } };

export function useVerifyEmail() {
  return useMutation<AuthResult, Error, string>({
    mutationFn: (token: string) =>
      api.post<AuthResult>("/auth/verify-email", { token }),
  });
}

export function useResendVerification() {
  return useMutation<void, Error, string>({
    mutationFn: (email: string) =>
      api.post<void>("/auth/resend-verification", { email }),
  });
}
