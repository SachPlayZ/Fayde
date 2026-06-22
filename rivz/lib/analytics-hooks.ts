import { useQuery } from "@tanstack/react-query";
import { api } from "./api";

export type DailyStat = { date: string; done: number; created: number };
export type OverdueUserStat = { user_email: string; count: number; oldest_due: string };

export type AnalyticsResponse = {
  by_status: Record<string, number>;
  by_priority: Record<string, number>;
  completion_rate_7d: DailyStat[];
  overdue_by_user: OverdueUserStat[];
};

export function useAdminAnalytics() {
  return useQuery<AnalyticsResponse>({
    queryKey: ["admin", "analytics"],
    queryFn: () => api.get<AnalyticsResponse>("/admin/analytics"),
  });
}
