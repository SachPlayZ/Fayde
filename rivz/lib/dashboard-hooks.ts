"use client";
import { useQuery } from "@tanstack/react-query";
import { startOfDay, addDays } from "date-fns";
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
  leetcode_due_today: number;
};

export function useDashboard() {
  return useQuery<DashboardSummary>({
    queryKey: ["dashboard"],
    queryFn: () => {
      const todayStart = startOfDay(new Date()).toISOString();
      const todayEnd = startOfDay(addDays(new Date(), 1)).toISOString();
      return api.get<DashboardSummary>(
        `/dashboard?today_start=${encodeURIComponent(todayStart)}&today_end=${encodeURIComponent(todayEnd)}`
      );
    },
  });
}
