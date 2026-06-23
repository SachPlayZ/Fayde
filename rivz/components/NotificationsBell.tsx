"use client";
import { useState } from "react";
import { formatDistanceToNow } from "date-fns";
import { Bell, CheckCheck, Layers, MessageSquare, AlertCircle, UserCheck, Clock, BellRing } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { useNotifications, useUnreadCount, useMarkRead, useMarkAllRead, useSnoozeNotification } from "@/lib/notifications-hooks";
import { cn } from "@/lib/utils";

const typeIcon: Record<string, React.ReactNode> = {
  mention: <MessageSquare className="size-3.5 text-blue-500" />,
  due_reminder: <AlertCircle className="size-3.5 text-amber-500" />,
  reminder: <BellRing className="size-3.5 text-amber-500" />,
  dependency_unblocked: <Layers className="size-3.5 text-emerald-500" />,
  assigned: <UserCheck className="size-3.5 text-violet-500" />,
};

const SNOOZE_OPTIONS: { label: string; ms: number }[] = [
  { label: "1h", ms: 60 * 60 * 1000 },
  { label: "Tomorrow", ms: 24 * 60 * 60 * 1000 },
];

export function NotificationsBell() {
  const [open, setOpen] = useState(false);
  const { data: countData } = useUnreadCount();
  const { data: notifications = [] } = useNotifications(false);
  const markRead = useMarkRead();
  const markAll = useMarkAllRead();
  const snooze = useSnoozeNotification();

  const unread = countData?.count ?? 0;
  const latest = notifications.slice(0, 10);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          <Bell className="size-4" />
          {unread > 0 && (
            <span className="absolute -top-0.5 -right-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-rose-500 text-[10px] font-bold text-white">
              {unread > 9 ? "9+" : unread}
            </span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-80 p-0">
        <div className="flex items-center justify-between px-3 py-2.5 border-b">
          <span className="text-sm font-semibold">Notifications</span>
          {unread > 0 && (
            <button
              onClick={() => markAll.mutate()}
              className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              <CheckCheck className="size-3" />
              Mark all read
            </button>
          )}
        </div>
        {latest.length === 0 ? (
          <p className="py-8 text-center text-sm text-muted-foreground">No notifications</p>
        ) : (
          <ul className="max-h-80 overflow-y-auto divide-y">
            {latest.map((n) => (
              <li
                key={n.id}
                className={cn(
                  "group/notif flex gap-2.5 px-3 py-2.5 cursor-pointer hover:bg-muted/50 transition-colors",
                  !n.read && "bg-blue-500/5"
                )}
                onClick={() => {
                  if (!n.read) markRead.mutate(n.id);
                }}
              >
                <span className="mt-0.5 shrink-0">{typeIcon[n.type] ?? <Bell className="size-3.5" />}</span>
                <div className="min-w-0 flex-1">
                  <p className="text-xs leading-snug">{n.message}</p>
                  <div className="flex items-center gap-2 mt-0.5">
                    <p className="text-[10px] text-muted-foreground">
                      {formatDistanceToNow(new Date(n.created_at), { addSuffix: true })}
                    </p>
                    <span className="flex items-center gap-1 opacity-0 group-hover/notif:opacity-100 transition-opacity">
                      <Clock className="size-2.5 text-muted-foreground" />
                      {SNOOZE_OPTIONS.map((o) => (
                        <button
                          key={o.label}
                          onClick={(e) => {
                            e.stopPropagation();
                            snooze.mutate({ id: n.id, until: new Date(Date.now() + o.ms).toISOString() });
                          }}
                          className="text-[10px] text-muted-foreground hover:text-foreground underline-offset-2 hover:underline"
                        >
                          {o.label}
                        </button>
                      ))}
                    </span>
                  </div>
                </div>
                {!n.read && <span className="mt-1.5 ml-auto shrink-0 size-1.5 rounded-full bg-blue-500" />}
              </li>
            ))}
          </ul>
        )}
      </PopoverContent>
    </Popover>
  );
}
