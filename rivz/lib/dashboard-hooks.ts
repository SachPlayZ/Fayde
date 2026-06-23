"use client";
import { useQuery } from "@tanstack/react-query";
import { api } from "./api";
import type { Habit } from "./habits-hooks";

export type TaskBrief = {
  id: string;
  title: string;
  priority: string;
  status: string;
  due_date: string | null;
};

export type DashboardSummary = {
  due_today: TaskBrief[];
  overdue: TaskBrief[];
  upcoming: TaskBrief[];
  completed_this_week: number;
  created_this_week: number;
  time_this_week_minutes: number;
  pomodoros_today: number;
  habits: Habit[];
};

export function useDashboard() {
  return useQuery<DashboardSummary>({
    queryKey: ["dashboard"],
    queryFn: () => api.get<DashboardSummary>("/dashboard"),
  });
}
