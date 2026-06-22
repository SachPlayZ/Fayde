"use client";
import { useState } from "react";
import { format, formatDistanceToNow } from "date-fns";
import { useAdminUsers } from "@/lib/admin-hooks";
import { useSiteMetrics } from "@/lib/site-metrics-hooks";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import { Users, ShieldCheck, TrendingUp } from "lucide-react";
import { SiteMetricsTab } from "./_components/SiteMetricsTab";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart";
import { AreaChart, Area, XAxis, YAxis, CartesianGrid } from "recharts";
import { format as dateFmt } from "date-fns";

const growthConfig: ChartConfig = {
  count: { label: "New Users", color: "hsl(var(--chart-2))" },
};

function fmtDate(d: unknown) {
  if (typeof d !== "string") return String(d ?? "");
  try { return dateFmt(new Date(d), "MMM d"); } catch { return d; }
}

function UsersTab() {
  const { data: users, isLoading } = useAdminUsers();
  const { data: metrics } = useSiteMetrics("30d");
  const list = users ?? [];

  const admins = list.filter((u) => u.role === "admin").length;

  return (
    <div className="flex flex-col gap-6 mt-4">
      {/* Stats */}
      <div className="grid grid-cols-2 sm:grid-cols-2 gap-3">
        {[
          { label: "Total users", value: list.length, icon: Users, color: "text-foreground" },
          { label: "Admins", value: admins, icon: ShieldCheck, color: "text-primary" },
        ].map((s, i) => (
          <div
            key={s.label}
            className="rounded-xl border border-border bg-card px-4 py-3 flex items-center gap-3 animate-in fade-in-0 slide-in-from-bottom-3 duration-400"
            style={{ animationDelay: `${i * 60}ms` }}
          >
            <s.icon className={cn("size-4 shrink-0", s.color)} />
            <div>
              <p className="text-[11px] text-muted-foreground leading-tight">{s.label}</p>
              <p className="text-lg font-semibold leading-tight">{isLoading ? "—" : s.value}</p>
            </div>
          </div>
        ))}
      </div>

      {/* User growth chart */}
      {metrics?.new_users && metrics.new_users.length > 0 && (
        <div className="rounded-xl border border-border bg-card p-4">
          <p className="text-sm font-semibold mb-4">User Growth (30d)</p>
          <ChartContainer config={growthConfig} className="h-40 w-full">
            <AreaChart data={metrics.new_users}>
              <defs>
                <linearGradient id="fillGrowth" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--color-count)" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="var(--color-count)" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" className="stroke-border/50" />
              <XAxis dataKey="date" tickFormatter={fmtDate} tick={{ fontSize: 10 }} tickLine={false} axisLine={false} />
              <YAxis tick={{ fontSize: 10 }} tickLine={false} axisLine={false} allowDecimals={false} />
              <ChartTooltip content={<ChartTooltipContent />} labelFormatter={fmtDate} />
              <Area type="monotone" dataKey="count" stroke="var(--color-count)" fill="url(#fillGrowth)" strokeWidth={2} dot={false} />
            </AreaChart>
          </ChartContainer>
        </div>
      )}

      {/* Users table */}
      {isLoading ? (
        <div className="flex flex-col gap-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full rounded-xl" />
          ))}
        </div>
      ) : list.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 gap-3 rounded-xl border border-border bg-card">
          <Users className="size-8 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">No users found</p>
        </div>
      ) : (
        <div className="rounded-xl border border-border overflow-hidden bg-card shadow-sm animate-in fade-in-0 slide-in-from-bottom-2 duration-400">
          <Table>
            <TableHeader>
              <TableRow className="bg-muted/40 hover:bg-muted/40">
                <TableHead>Email</TableHead>
                <TableHead className="w-24">Role</TableHead>
                <TableHead className="w-36">Joined</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {list.map((u, i) => (
                <TableRow
                  key={u.id}
                  className="animate-in fade-in-0 duration-300"
                  style={{ animationDelay: `${i * 30}ms` }}
                >
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <div className="flex size-7 items-center justify-center rounded-full bg-secondary text-secondary-foreground text-xs font-semibold shrink-0 select-none">
                        {u.email.slice(0, 2).toUpperCase()}
                      </div>
                      <span className="text-sm font-medium">{u.email}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span
                      className={cn(
                        "inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium",
                        u.role === "admin"
                          ? "bg-primary/10 text-primary"
                          : "bg-muted text-muted-foreground"
                      )}
                    >
                      {u.role === "admin" && <ShieldCheck className="size-2.5" />}
                      {u.role}
                    </span>
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-col">
                      <span className="text-xs text-muted-foreground">
                        {format(new Date(u.created_at), "MMM d, yyyy")}
                      </span>
                      <span className="text-[10px] text-muted-foreground/60">
                        {formatDistanceToNow(new Date(u.created_at), { addSuffix: true })}
                      </span>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}

export default function AdminPage() {
  return (
    <div className="flex flex-col gap-6">
      <div className="animate-in fade-in-0 slide-in-from-bottom-3 duration-400">
        <h2 className="text-xl font-bold tracking-tight">Admin Dashboard</h2>
        <p className="text-sm text-muted-foreground mt-0.5">Site metrics and user analytics</p>
      </div>
      <Tabs defaultValue="metrics">
        <TabsList className="bg-muted p-1 rounded-xl">
          <TabsTrigger value="metrics" className="rounded-lg text-xs">
            <TrendingUp className="size-3.5 mr-1.5" />
            Site Metrics
          </TabsTrigger>
          <TabsTrigger value="users" className="rounded-lg text-xs">
            <Users className="size-3.5 mr-1.5" />
            Users
          </TabsTrigger>
        </TabsList>
        <TabsContent value="metrics">
          <SiteMetricsTab />
        </TabsContent>
        <TabsContent value="users">
          <UsersTab />
        </TabsContent>
      </Tabs>
    </div>
  );
}
