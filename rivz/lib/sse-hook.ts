"use client";
import { useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type SSEEvent = { type: string; payload: unknown };

export function useSSE() {
  const qc = useQueryClient();
  const esRef = useRef<EventSource | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [lastEvent, setLastEvent] = useState<SSEEvent | null>(null);

  useEffect(() => {
    let destroyed = false;

    function connect() {
      if (destroyed) return;

      const token =
        typeof window !== "undefined" ? localStorage.getItem("token") : null;
      if (!token) return;

      const url = `${BASE_URL}/events?token=${encodeURIComponent(token)}`;
      const es = new EventSource(url);
      esRef.current = es;

      es.onmessage = (e) => {
        try {
          const data = JSON.parse(e.data) as SSEEvent;
          setLastEvent(data);
          qc.invalidateQueries({ queryKey: ["tasks"] });
          qc.invalidateQueries({ queryKey: ["activity", "global"] });
          if (data.type === "notification") {
            qc.invalidateQueries({ queryKey: ["notifications"] });
            qc.invalidateQueries({ queryKey: ["notifications", "unread-count"] });
          }
        } catch {
          qc.invalidateQueries({ queryKey: ["tasks"] });
        }
      };

      es.onerror = () => {
        es.close();
        esRef.current = null;
        if (!destroyed) {
          timerRef.current = setTimeout(connect, 3000);
        }
      };
    }

    connect();

    return () => {
      destroyed = true;
      if (timerRef.current !== null) {
        clearTimeout(timerRef.current);
      }
      esRef.current?.close();
      esRef.current = null;
    };
  }, [qc]);

  return { lastEvent };
}
