"use client";
import { useState } from "react";
import { toast } from "sonner";
import {
  useAutomations,
  useCreateAutomation,
  useUpdateAutomation,
  useDeleteAutomation,
  type Action,
} from "@/lib/automations-hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import { Zap, Plus, Trash2, ArrowRight } from "lucide-react";

const EVENTS = [
  { v: "created", label: "Task created" },
  { v: "status_changed", label: "Status changes to" },
  { v: "updated", label: "Task updated" },
];
const STATUSES = ["todo", "in_progress", "done", "failed"];
const PRIORITIES = ["low", "medium", "high"];
const ACTION_TYPES = [
  { v: "set_priority", label: "Set priority" },
  { v: "set_status", label: "Set status" },
  { v: "notify", label: "Notify me" },
  { v: "webhook", label: "Send to chat" },
];

function actionValueOptions(type: string): string[] | null {
  if (type === "set_priority") return PRIORITIES;
  if (type === "set_status") return STATUSES;
  return null;
}

export function AutomationsTab() {
  const { data: rules, isLoading } = useAutomations();
  const create = useCreateAutomation();
  const update = useUpdateAutomation();
  const del = useDeleteAutomation();

  const [name, setName] = useState("");
  const [event, setEvent] = useState("status_changed");
  const [toStatus, setToStatus] = useState("done");
  const [actions, setActions] = useState<Action[]>([{ type: "notify", value: "" }]);

  const addAction = () => setActions((a) => [...a, { type: "notify", value: "" }]);
  const setAction = (i: number, patch: Partial<Action>) =>
    setActions((a) => a.map((act, idx) => (idx === i ? { ...act, ...patch } : act)));
  const removeAction = (i: number) => setActions((a) => a.filter((_, idx) => idx !== i));

  const handleCreate = () => {
    if (!name.trim()) {
      toast.error("Name your rule");
      return;
    }
    create.mutate(
      {
        name: name.trim(),
        trigger: { event, to: event === "status_changed" ? toStatus : undefined },
        conditions: [],
        actions: actions.filter((a) => a.type),
      },
      {
        onSuccess: () => {
          setName("");
          setActions([{ type: "notify", value: "" }]);
          toast.success("Automation created");
        },
        onError: () => toast.error("Failed to create"),
      }
    );
  };

  return (
    <div className="mt-4 flex flex-col gap-5">
      {/* Builder */}
      <div className="rounded-xl border border-border bg-card p-5 flex flex-col gap-4">
        <div className="flex items-center gap-2">
          <Zap className="size-4 text-amber-500" />
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Rule name…"
            className="flex-1"
          />
        </div>

        <div className="flex flex-wrap items-center gap-2 text-sm">
          <span className="text-muted-foreground">When</span>
          <Select value={event} onValueChange={setEvent}>
            <SelectTrigger className="w-44"><SelectValue /></SelectTrigger>
            <SelectContent>
              {EVENTS.map((e) => <SelectItem key={e.v} value={e.v}>{e.label}</SelectItem>)}
            </SelectContent>
          </Select>
          {event === "status_changed" && (
            <Select value={toStatus} onValueChange={setToStatus}>
              <SelectTrigger className="w-32"><SelectValue /></SelectTrigger>
              <SelectContent>
                {STATUSES.map((s) => <SelectItem key={s} value={s}>{s}</SelectItem>)}
              </SelectContent>
            </Select>
          )}
        </div>

        <div className="flex flex-col gap-2">
          <span className="text-sm text-muted-foreground flex items-center gap-1">
            <ArrowRight className="size-3.5" /> Then
          </span>
          {actions.map((a, i) => {
            const opts = actionValueOptions(a.type);
            return (
              <div key={i} className="flex items-center gap-2">
                <Select value={a.type} onValueChange={(v) => setAction(i, { type: v, value: "" })}>
                  <SelectTrigger className="w-40"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {ACTION_TYPES.map((t) => <SelectItem key={t.v} value={t.v}>{t.label}</SelectItem>)}
                  </SelectContent>
                </Select>
                {opts ? (
                  <Select value={a.value} onValueChange={(v) => setAction(i, { value: v })}>
                    <SelectTrigger className="w-32"><SelectValue placeholder="value" /></SelectTrigger>
                    <SelectContent>
                      {opts.map((o) => <SelectItem key={o} value={o}>{o}</SelectItem>)}
                    </SelectContent>
                  </Select>
                ) : a.type === "webhook" ? (
                  <>
                    <Select value={a.kind ?? "slack"} onValueChange={(v) => setAction(i, { kind: v })}>
                      <SelectTrigger className="w-28"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="slack">Slack</SelectItem>
                        <SelectItem value="discord">Discord</SelectItem>
                      </SelectContent>
                    </Select>
                    <Input
                      value={a.value}
                      onChange={(e) => setAction(i, { value: e.target.value })}
                      placeholder="webhook URL"
                      className="flex-1"
                    />
                  </>
                ) : (
                  <Input
                    value={a.value}
                    onChange={(e) => setAction(i, { value: e.target.value })}
                    placeholder="message (optional)"
                    className="flex-1"
                  />
                )}
                {actions.length > 1 && (
                  <Trash2
                    className="size-4 text-muted-foreground hover:text-destructive cursor-pointer shrink-0"
                    onClick={() => removeAction(i)}
                  />
                )}
              </div>
            );
          })}
          <button
            onClick={addAction}
            className="self-start text-xs text-muted-foreground hover:text-foreground inline-flex items-center gap-1"
          >
            <Plus className="size-3.5" /> Add action
          </button>
        </div>

        <Button onClick={handleCreate} disabled={create.isPending} className="self-start">
          Create rule
        </Button>
      </div>

      {/* Existing rules */}
      {isLoading ? (
        <div className="h-24 rounded-xl border border-border bg-card animate-pulse" />
      ) : (rules?.length ?? 0) === 0 ? (
        <p className="text-sm text-muted-foreground text-center py-8">No automations yet.</p>
      ) : (
        <div className="flex flex-col divide-y divide-border rounded-xl border border-border bg-card overflow-hidden">
          {rules!.map((r) => (
            <div key={r.id} className="flex items-center gap-3 px-5 py-3">
              <button
                onClick={() => update.mutate({ id: r.id, patch: { enabled: !r.enabled } })}
                className={cn(
                  "relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors",
                  r.enabled ? "bg-primary" : "bg-muted-foreground/30"
                )}
              >
                <span className={cn("inline-block size-4 rounded-full bg-white shadow transition-transform", r.enabled ? "translate-x-4" : "translate-x-0.5")} />
              </button>
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium truncate">{r.name}</p>
                <p className="text-xs text-muted-foreground">
                  {r.trigger.event === "status_changed" ? `status → ${r.trigger.to}` : r.trigger.event}
                  {" · "}
                  {r.actions.map((a) => a.type).join(", ")}
                </p>
              </div>
              <Trash2
                className="size-4 text-muted-foreground hover:text-destructive cursor-pointer shrink-0"
                onClick={() => del.mutate(r.id, { onSuccess: () => toast.success("Deleted") })}
              />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
