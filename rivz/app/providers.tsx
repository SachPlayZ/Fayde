"use client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "next-themes";
import { AuthProvider, useAuth } from "@/lib/auth-context";
import { useSSE } from "@/lib/sse-hook";
import { Toaster } from "sonner";
import { useState } from "react";

/**
 * SSEConnector always calls useSSE (satisfies Rules of Hooks) but the hook
 * itself checks for a token before opening a connection.
 */
function SSEConnector() {
  useSSE();
  return null;
}

function InnerProviders({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();
  return (
    <>
      {user && <SSEConnector />}
      {children}
    </>
  );
}

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => new QueryClient());
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
        <AuthProvider>
          <InnerProviders>{children}</InnerProviders>
          <Toaster />
        </AuthProvider>
      </ThemeProvider>
    </QueryClientProvider>
  );
}
