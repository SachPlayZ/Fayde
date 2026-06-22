import { useQuery } from "@tanstack/react-query";
import { api } from "./api";

export type DailyCount = { date: string; count: number };
export type DailyViews = { date: string; views: number; unique: number };
export type PageCount = { path: string; count: number };

export type SiteMetricsResponse = {
  total_users: number;
  new_users: DailyCount[];
  total_views: number;
  page_views: DailyViews[];
  unique_visitors: number;
  top_pages: PageCount[];
  active_users_7d: number;
};

export type MetricsRange = "7d" | "30d" | "90d";

export function useSiteMetrics(range: MetricsRange) {
  return useQuery<SiteMetricsResponse>({
    queryKey: ["site-metrics", range],
    queryFn: () => api.get<SiteMetricsResponse>(`/admin/site-metrics?range=${range}`),
  });
}
