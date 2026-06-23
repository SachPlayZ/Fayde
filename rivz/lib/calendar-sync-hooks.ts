"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type CalendarStatus = {
  connected: boolean;
  email?: string;
};

export function useCalendarStatus() {
  return useQuery<CalendarStatus>({
    queryKey: ["calendar-status"],
    queryFn: () => api.get<CalendarStatus>("/calendar/status"),
  });
}

export function useDisconnectCalendar() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api.delete<void>("/calendar/disconnect"),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["calendar-status"] });
    },
  });
}
