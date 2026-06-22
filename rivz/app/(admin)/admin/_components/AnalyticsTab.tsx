"use client";
import { useAdminAnalytics } from "@/lib/analytics-hooks";
import {
  PieChart,
  Pie,
  Cell,
  BarChart,
  Bar,
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { Skeleton } from "@/components/ui/skeleton";
import { format } from "date-fns";

const STATUS_COLORS: Record<string, string> = {
  todo: "#94a3b8",
  in_progress: "#3b82f6",
  done: "#10b981",
  failed: "#ef4444",
};

const PRIORITY_COLORS: Record<string, string> = {
  low: "#10b981",
  medium: "#f59e0b",
  high: "#ef4444",
};

export function AnalyticsTab() {
  const { data, isLoading } = useAdminAnalytics();

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {[...Array(4)].map((_, i) => <Skeleton key={i} className="h-64 rounded-xl" />)}
      </div>
    );
  }

  if (!data) return <p className="text-sm text-muted-foreground">No data available.</p>;

  const statusData = Object.entries(data.by_status).map(([name, value]) => ({ name, value }));
  const priorityData = Object.entries(data.by_priority).map(([name, value]) => ({ name, value }));

  return (
    <div className="flex flex-col gap-8">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Tasks by status — donut */}
        <div className="rounded-xl border border-border bg-card p-4">
          <h3 className="text-sm font-semibold mb-4">Tasks by Status</h3>
          <ResponsiveContainer width="100%" height={220}>
            <PieChart>
              <Pie data={statusData} dataKey="value" nameKey="name" cx="50%" cy="50%" innerRadius={55} outerRadius={80} paddingAngle={3}>
                {statusData.map((entry) => (
                  <Cell key={entry.name} fill={STATUS_COLORS[entry.name] ?? "#6366f1"} />
                ))}
              </Pie>
              <Tooltip formatter={(v) => [v, ""]} />
              <Legend formatter={(v) => v.replace("_", " ")} />
            </PieChart>
          </ResponsiveContainer>
        </div>

        {/* Tasks by priority — bar */}
        <div className="rounded-xl border border-border bg-card p-4">
          <h3 className="text-sm font-semibold mb-4">Tasks by Priority</h3>
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={priorityData} barCategoryGap="40%">
              <XAxis dataKey="name" tick={{ fontSize: 12 }} />
              <YAxis tick={{ fontSize: 12 }} allowDecimals={false} />
              <Tooltip />
              <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                {priorityData.map((entry) => (
                  <Cell key={entry.name} fill={PRIORITY_COLORS[entry.name] ?? "#6366f1"} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* 7-day completion rate — line */}
      <div className="rounded-xl border border-border bg-card p-4">
        <h3 className="text-sm font-semibold mb-4">7-Day Completion Rate</h3>
        <ResponsiveContainer width="100%" height={220}>
          <LineChart data={data.completion_rate_7d}>
            <XAxis
              dataKey="date"
              tick={{ fontSize: 11 }}
              tickFormatter={(d) => {
                try { return format(new Date(d), "MMM d"); } catch { return d; }
              }}
            />
            <YAxis tick={{ fontSize: 12 }} allowDecimals={false} />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="done" stroke="#10b981" strokeWidth={2} dot={false} name="Done" />
            <Line type="monotone" dataKey="created" stroke="#6366f1" strokeWidth={2} dot={false} name="Created" />
          </LineChart>
        </ResponsiveContainer>
      </div>

      {/* Overdue by user — table */}
      {data.overdue_by_user.length > 0 && (
        <div className="rounded-xl border border-border bg-card p-4">
          <h3 className="text-sm font-semibold mb-4">Overdue Tasks by User</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="text-left py-2 pr-4 text-xs text-muted-foreground font-medium">User</th>
                  <th className="text-left py-2 pr-4 text-xs text-muted-foreground font-medium">Count</th>
                  <th className="text-left py-2 text-xs text-muted-foreground font-medium">Oldest Due</th>
                </tr>
              </thead>
              <tbody>
                {data.overdue_by_user.map((row) => (
                  <tr key={row.user_email} className="border-b last:border-0">
                    <td className="py-2 pr-4 text-xs">{row.user_email}</td>
                    <td className="py-2 pr-4 text-xs font-medium text-rose-500">{row.count}</td>
                    <td className="py-2 text-xs text-muted-foreground">{row.oldest_due}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
