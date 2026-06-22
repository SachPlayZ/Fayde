"use client";
import { useState } from "react";
import { format } from "date-fns";
import { Users, Eye, MousePointerClick, Activity, TrendingUp } from "lucide-react";
import { cn } from "@/lib/utils";
import { useSiteMetrics, MetricsRange } from "@/lib/site-metrics-hooks";
import { Skeleton } from "@/components/ui/skeleton";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  ChartLegend,
  ChartLegendContent,
  type ChartConfig,
} from "@/components/ui/chart";
import {
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
} from "recharts";

const RANGES: { value: MetricsRange; label: string }[] = [
  { value: "7d", label: "7d" },
  { value: "30d", label: "30d" },
  { value: "90d", label: "90d" },
];

const pageViewsConfig: ChartConfig = {
  views: { label: "Page Views", color: "hsl(var(--chart-1))" },
  unique: { label: "Unique Visitors", color: "hsl(var(--chart-2))" },
};

const newUsersConfig: ChartConfig = {
  count: { label: "New Signups", color: "hsl(var(--chart-3))" },
};

const topPagesConfig: ChartConfig = {
  count: { label: "Views", color: "hsl(var(--chart-1))" },
};

function fmtDate(d: unknown) {
  if (typeof d !== "string") return String(d ?? "");
  try { return format(new Date(d), "MMM d"); } catch { return d; }
}

export function SiteMetricsTab() {
  const [range, setRange] = useState<MetricsRange>("30d");
  const { data, isLoading } = useSiteMetrics(range);

  const kpis = [
    {
      label: "Total Users",
      value: data?.total_users ?? 0,
      icon: Users,
      color: "text-foreground",
    },
    {
      label: `Unique Visitors (${range})`,
      value: data?.unique_visitors ?? 0,
      icon: MousePointerClick,
      color: "text-blue-500",
    },
    {
      label: `Page Views (${range})`,
      value: data?.total_views ?? 0,
      icon: Eye,
      color: "text-violet-500",
    },
    {
      label: "Active Users (7d)",
      value: data?.active_users_7d ?? 0,
      icon: Activity,
      color: "text-emerald-500",
    },
  ];

  return (
    <div className="flex flex-col gap-6 mt-4">
      {/* Range toggle */}
      <div className="flex items-center gap-1 rounded-xl bg-muted p-1 w-fit">
        {RANGES.map((r) => (
          <button
            key={r.value}
            onClick={() => setRange(r.value)}
            className={cn(
              "rounded-lg px-3 py-1 text-xs font-medium transition-all duration-150",
              range === r.value
                ? "bg-background text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground"
            )}
          >
            {r.label}
          </button>
        ))}
      </div>

      {/* KPI cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {kpis.map((k, i) => (
          <div
            key={k.label}
            className="rounded-xl border border-border bg-card px-4 py-3 flex items-center gap-3 animate-in fade-in-0 slide-in-from-bottom-3 duration-400"
            style={{ animationDelay: `${i * 60}ms` }}
          >
            <k.icon className={cn("size-4 shrink-0", k.color)} />
            <div>
              <p className="text-[11px] text-muted-foreground leading-tight">{k.label}</p>
              <p className="text-lg font-semibold leading-tight">
                {isLoading ? "—" : k.value.toLocaleString()}
              </p>
            </div>
          </div>
        ))}
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {[...Array(3)].map((_, i) => <Skeleton key={i} className="h-64 rounded-xl" />)}
        </div>
      ) : (
        <>
          {/* Page Views area chart */}
          <div className="rounded-xl border border-border bg-card p-4">
            <div className="flex items-center gap-2 mb-4">
              <TrendingUp className="size-4 text-muted-foreground" />
              <h3 className="text-sm font-semibold">Traffic Over Time</h3>
            </div>
            <ChartContainer config={pageViewsConfig} className="h-56 w-full">
              <AreaChart data={data?.page_views ?? []}>
                <defs>
                  <linearGradient id="fillViews" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="var(--color-views)" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="var(--color-views)" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="fillUnique" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="var(--color-unique)" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="var(--color-unique)" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" className="stroke-border/50" />
                <XAxis dataKey="date" tickFormatter={fmtDate} tick={{ fontSize: 11 }} tickLine={false} axisLine={false} />
                <YAxis tick={{ fontSize: 11 }} tickLine={false} axisLine={false} allowDecimals={false} />
                <ChartTooltip content={<ChartTooltipContent />} labelFormatter={fmtDate} />
                <ChartLegend content={<ChartLegendContent />} />
                <Area type="monotone" dataKey="views" stroke="var(--color-views)" fill="url(#fillViews)" strokeWidth={2} dot={false} />
                <Area type="monotone" dataKey="unique" stroke="var(--color-unique)" fill="url(#fillUnique)" strokeWidth={2} dot={false} />
              </AreaChart>
            </ChartContainer>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* New signups area chart */}
            <div className="rounded-xl border border-border bg-card p-4">
              <div className="flex items-center gap-2 mb-4">
                <Users className="size-4 text-muted-foreground" />
                <h3 className="text-sm font-semibold">New Signups</h3>
              </div>
              <ChartContainer config={newUsersConfig} className="h-48 w-full">
                <AreaChart data={data?.new_users ?? []}>
                  <defs>
                    <linearGradient id="fillCount" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="var(--color-count)" stopOpacity={0.3} />
                      <stop offset="95%" stopColor="var(--color-count)" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-border/50" />
                  <XAxis dataKey="date" tickFormatter={fmtDate} tick={{ fontSize: 11 }} tickLine={false} axisLine={false} />
                  <YAxis tick={{ fontSize: 11 }} tickLine={false} axisLine={false} allowDecimals={false} />
                  <ChartTooltip content={<ChartTooltipContent />} labelFormatter={fmtDate} />
                  <Area type="monotone" dataKey="count" stroke="var(--color-count)" fill="url(#fillCount)" strokeWidth={2} dot={false} />
                </AreaChart>
              </ChartContainer>
            </div>

            {/* Top pages bar chart */}
            <div className="rounded-xl border border-border bg-card p-4">
              <div className="flex items-center gap-2 mb-4">
                <Eye className="size-4 text-muted-foreground" />
                <h3 className="text-sm font-semibold">Top Pages</h3>
              </div>
              {!data?.top_pages?.length ? (
                <div className="flex items-center justify-center h-48 text-sm text-muted-foreground">
                  No page view data yet
                </div>
              ) : (
                <ChartContainer config={topPagesConfig} className="h-48 w-full">
                  <BarChart
                    data={data.top_pages}
                    layout="vertical"
                    margin={{ left: 8, right: 16 }}
                  >
                    <XAxis type="number" tick={{ fontSize: 11 }} tickLine={false} axisLine={false} allowDecimals={false} />
                    <YAxis
                      type="category"
                      dataKey="path"
                      tick={{ fontSize: 10 }}
                      tickLine={false}
                      axisLine={false}
                      width={90}
                      tickFormatter={(v: string) => v.length > 14 ? v.slice(0, 13) + "…" : v}
                    />
                    <ChartTooltip content={<ChartTooltipContent />} />
                    <Bar dataKey="count" fill="var(--color-count)" radius={[0, 4, 4, 0]} />
                  </BarChart>
                </ChartContainer>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
