"use client";
import { use, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { format } from "date-fns";
import {
  useBoard,
  useAddBoardTask,
  useDeleteBoardTask,
  useCompleteBoardTask,
  useUncompleteBoardTask,
  useInviteFriendToBoard,
  useCreateBoardShareToken,
  useRevokeBoardShareToken,
  useShareTaskToBoard,
  useUnshareTaskFromBoard,
  type BoardMember,
  type BoardTask,
} from "@/lib/boards-hooks";
import { useFriends } from "@/lib/friends-hooks";
import { useTasks } from "@/lib/tasks-hooks";
import { useAuth } from "@/lib/auth-context";
import { useSSE } from "@/lib/sse-hook";
import { useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import {
  Check,
  Plus,
  Trash2,
  Loader2,
  UserPlus,
  Link as LinkIcon,
  Copy,
  X,
  LayoutGrid,
  ListTodo,
  Share2,
} from "lucide-react";
import { format as fmtDate } from "date-fns";

const statusConfig: Record<string, { label: string; className: string }> = {
  todo: { label: "Todo", className: "bg-muted text-muted-foreground border-muted" },
  in_progress: {
    label: "In Progress",
    className: "bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20",
  },
  done: {
    label: "Done",
    className: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20",
  },
  failed: {
    label: "Failed",
    className: "bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20",
  },
};

function MemberAvatar({ member }: { member: BoardMember }) {
  const initials = member.display_name
    ? member.display_name
        .trim()
        .split(/\s+/)
        .map((p) => p[0])
        .filter(Boolean)
        .join("")
        .slice(0, 2)
        .toUpperCase()
    : member.email.slice(0, 2).toUpperCase();

  return (
    <div
      className="size-8 rounded-full bg-primary/10 ring-1 ring-primary/20 flex items-center justify-center overflow-hidden shrink-0"
      title={member.display_name || member.email}
    >
      {member.avatar_url ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={member.avatar_url}
          alt={member.display_name || member.email}
          className="size-full object-cover"
          referrerPolicy="no-referrer"
        />
      ) : (
        <span className="text-primary text-[10px] font-bold">{initials}</span>
      )}
    </div>
  );
}

export default function BoardDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const router = useRouter();
  const { user } = useAuth();
  const qc = useQueryClient();
  const { lastEvent } = useSSE();

  const { data: board, isLoading } = useBoard(id);
  const { data: friends } = useFriends();

  const addTask = useAddBoardTask();
  const deleteTask = useDeleteBoardTask();
  const completeTask = useCompleteBoardTask();
  const uncompleteTask = useUncompleteBoardTask();
  const inviteFriend = useInviteFriendToBoard();
  const createShareToken = useCreateBoardShareToken();
  const revokeShareToken = useRevokeBoardShareToken();
  const shareTask = useShareTaskToBoard();
  const unshareTask = useUnshareTaskFromBoard();

  const [newTaskTitle, setNewTaskTitle] = useState("");

  // Today's date for completion lookup
  const today = format(new Date(), "yyyy-MM-dd");

  // Personal tasks for accordion
  const { data: myTasksData } = useTasks({
    status: "todo",
    limit: 50,
  });

  // Personal tasks (any status) for the "Share a Task" picker
  const { data: myAllTasksData } = useTasks({ limit: 50 });

  // SSE: refresh board when a board_task_completed or shared_task_updated event arrives
  useEffect(() => {
    if (
      lastEvent?.type === "board_task_completed" ||
      lastEvent?.type === "shared_task_updated"
    ) {
      qc.invalidateQueries({ queryKey: ["board", id] });
    }
  }, [lastEvent, id, qc]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-24">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (!board) {
    return (
      <div className="flex flex-col items-center justify-center py-24 gap-3 text-muted-foreground">
        <LayoutGrid className="size-8" />
        <p className="text-sm">Board not found.</p>
        <Button variant="outline" onClick={() => router.push("/boards")}>
          Back to Boards
        </Button>
      </div>
    );
  }

  const myUserId = user?.id;
  const members = board.members ?? [];
  const tasks = board.tasks ?? [];
  const completions = board.completions ?? [];
  const sharedTasks = board.shared_tasks ?? [];
  const sharedTaskIds = new Set(sharedTasks.map((t) => t.task_id));

  // Build a set for quick lookup: "taskId:userId" -> true if completed today
  const completedSet = new Set(
    completions
      .filter((c) => c.completion_date === today)
      .map((c) => `${c.board_task_id}:${c.user_id}`)
  );

  // Friends not already on the board
  const memberIds = new Set(members.map((m) => m.user_id));
  const friendsNotOnBoard = (friends ?? []).filter(
    (f) => !memberIds.has(f.user.id)
  );

  const handleAddTask = () => {
    const title = newTaskTitle.trim();
    if (!title) return;
    addTask.mutate(
      { boardId: id, title },
      {
        onSuccess: () => {
          setNewTaskTitle("");
          toast.success("Task added");
        },
        onError: (err) => toast.error(err.message || "Failed to add task"),
      }
    );
  };

  const handleToggleCell = (task: BoardTask, memberId: string) => {
    if (memberId !== myUserId) return; // read-only for others
    const key = `${task.id}:${memberId}`;
    const isDone = completedSet.has(key);
    if (isDone) {
      uncompleteTask.mutate(
        { boardId: id, taskId: task.id },
        { onError: (err) => toast.error(err.message || "Failed") }
      );
    } else {
      completeTask.mutate(
        { boardId: id, taskId: task.id },
        { onError: (err) => toast.error(err.message || "Failed") }
      );
    }
  };

  const handleToggleShare = (taskId: string, isShared: boolean) => {
    if (isShared) {
      unshareTask.mutate(
        { boardId: id, taskId },
        { onError: (err) => toast.error(err.message || "Failed to unshare task") }
      );
    } else {
      shareTask.mutate(
        { boardId: id, taskId },
        {
          onSuccess: () => toast.success("Task shared to board"),
          onError: (err) => toast.error(err.message || "Failed to share task"),
        }
      );
    }
  };

  const handleInvite = (friendUserId: string) => {
    inviteFriend.mutate(
      { boardId: id, friend_user_id: friendUserId },
      {
        onSuccess: () => toast.success("Friend added to board"),
        onError: (err) => toast.error(err.message || "Failed to invite"),
      }
    );
  };

  const shareToken = board.share_token;
  const shareUrl = shareToken
    ? `${typeof window !== "undefined" ? window.location.origin : ""}/join/${shareToken}`
    : null;

  return (
    <div className="flex flex-col gap-6 max-w-4xl">
      {/* Header */}
      <div>
        <h2 className="text-xl font-bold tracking-tight">{board.name}</h2>
        {board.description && (
          <p className="text-sm text-muted-foreground mt-0.5">{board.description}</p>
        )}
      </div>

      {/* Completion Matrix */}
      <div className="rounded-xl border border-border bg-card overflow-hidden">
        <div className="px-4 py-3 border-b border-border bg-muted/30 flex items-center justify-between gap-2">
          <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
            Today — {fmtDate(new Date(), "MMM d, yyyy")}
          </p>
        </div>

        {tasks.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-2 py-10 text-muted-foreground">
            <ListTodo className="size-6" />
            <p className="text-sm">No tasks yet. Add one below.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left px-4 py-2.5 font-medium text-muted-foreground text-xs min-w-[160px]">
                    Task
                  </th>
                  {members.map((m) => (
                    <th
                      key={m.user_id}
                      className="px-3 py-2.5 text-center"
                    >
                      <div className="flex flex-col items-center gap-1">
                        <MemberAvatar member={m} />
                        <span className="text-[10px] text-muted-foreground max-w-16 truncate">
                          {m.display_name || m.email.split("@")[0]}
                        </span>
                      </div>
                    </th>
                  ))}
                  <th className="px-3 py-2.5 w-10" />
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {tasks.map((task) => (
                  <tr
                    key={task.id}
                    className="hover:bg-muted/20 transition-colors"
                  >
                    <td className="px-4 py-3 font-medium">{task.title}</td>
                    {members.map((m) => {
                      const key = `${task.id}:${m.user_id}`;
                      const done = completedSet.has(key);
                      const isMe = m.user_id === myUserId;
                      return (
                        <td key={m.user_id} className="px-3 py-3 text-center">
                          <button
                            onClick={() => handleToggleCell(task, m.user_id)}
                            disabled={!isMe}
                            className={cn(
                              "size-7 rounded-md mx-auto flex items-center justify-center transition-all duration-150",
                              done
                                ? "bg-emerald-500 text-white shadow-sm"
                                : isMe
                                ? "border-2 border-dashed border-muted-foreground/40 hover:border-primary/60 hover:bg-primary/5 text-muted-foreground/50 cursor-pointer"
                                : "border border-border text-muted-foreground/20 cursor-default"
                            )}
                            aria-label={
                              isMe
                                ? done
                                  ? "Unmark complete"
                                  : "Mark complete"
                                : undefined
                            }
                          >
                            {done && <Check className="size-3.5 stroke-[3]" />}
                          </button>
                        </td>
                      );
                    })}
                    <td className="px-3 py-3">
                      <Button
                        size="icon"
                        variant="ghost"
                        className="size-6 text-muted-foreground/50 hover:text-destructive"
                        onClick={() =>
                          deleteTask.mutate(
                            { boardId: id, taskId: task.id },
                            {
                              onSuccess: () => toast.success("Task removed"),
                              onError: (err) =>
                                toast.error(err.message || "Failed"),
                            }
                          )
                        }
                      >
                        <Trash2 className="size-3.5" />
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Add task row */}
        <div className="px-4 py-3 border-t border-border flex items-center gap-2">
          <Input
            value={newTaskTitle}
            onChange={(e) => setNewTaskTitle(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleAddTask()}
            placeholder="Add a task (e.g. Go to the Gym)…"
            className="flex-1 h-8 text-sm"
          />
          <Button
            size="sm"
            onClick={handleAddTask}
            disabled={!newTaskTitle.trim() || addTask.isPending}
            className="h-8 gap-1.5"
          >
            <Plus className="size-3.5" /> Add
          </Button>
        </div>
      </div>

      {/* Shared Tasks (read-only for other members) */}
      <div className="rounded-xl border border-border bg-card overflow-hidden">
        <div className="px-4 py-2.5 border-b border-border bg-muted/30">
          <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
            Shared Tasks
          </p>
        </div>
        {sharedTasks.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-1.5 py-6 text-muted-foreground text-center px-4">
            <Share2 className="size-5" />
            <p className="text-xs">No one has shared a personal task to this board yet.</p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {sharedTasks.map((st) => {
              const cfg = statusConfig[st.status];
              const isOwner = st.shared_by === myUserId;
              return (
                <div key={`${st.board_id}:${st.task_id}`} className="flex items-start gap-3 px-4 py-2.5">
                  <div className="size-6 rounded-full bg-primary/10 ring-1 ring-primary/20 flex items-center justify-center overflow-hidden shrink-0 mt-0.5">
                    {st.owner_avatar_url ? (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img
                        src={st.owner_avatar_url}
                        alt={st.owner_display_name || st.owner_email}
                        className="size-full object-cover"
                        referrerPolicy="no-referrer"
                      />
                    ) : (
                      <span className="text-primary text-[9px] font-bold">
                        {(st.owner_display_name || st.owner_email).slice(0, 2).toUpperCase()}
                      </span>
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium truncate">{st.title}</span>
                      {cfg && (
                        <Badge className={cn("text-[10px] px-1.5 py-0 shrink-0", cfg.className)}>
                          {cfg.label}
                        </Badge>
                      )}
                    </div>
                    {st.description && (
                      <p className="text-xs text-muted-foreground truncate mt-0.5">{st.description}</p>
                    )}
                    <p className="text-[10px] text-muted-foreground mt-0.5">
                      Shared by {st.owner_display_name || st.owner_email}
                      {st.due_date && ` · Due ${fmtDate(new Date(st.due_date), "MMM d, yyyy")}`}
                    </p>
                  </div>
                  {isOwner && (
                    <Button
                      size="icon"
                      variant="ghost"
                      className="size-6 text-muted-foreground/50 hover:text-destructive shrink-0"
                      onClick={() => handleToggleShare(st.task_id, true)}
                    >
                      <X className="size-3.5" />
                    </Button>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Share a Task */}
      <div className="rounded-xl border border-border bg-card overflow-hidden">
        <div className="px-4 py-2.5 border-b border-border bg-muted/30">
          <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
            Share a Task
          </p>
        </div>
        {(myAllTasksData?.data?.length ?? 0) === 0 ? (
          <div className="flex flex-col items-center justify-center gap-1.5 py-6 text-muted-foreground text-center px-4">
            <p className="text-xs">You have no tasks to share yet.</p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {myAllTasksData!.data.map((task) => {
              const isShared = sharedTaskIds.has(task.id);
              const cfg = statusConfig[task.status];
              return (
                <label
                  key={task.id}
                  className="flex items-center gap-3 px-4 py-2.5 cursor-pointer hover:bg-muted/20 transition-colors"
                >
                  <input
                    type="checkbox"
                    checked={isShared}
                    onChange={() => handleToggleShare(task.id, isShared)}
                    className="rounded"
                  />
                  <span className="flex-1 text-sm truncate">{task.title}</span>
                  {cfg && (
                    <Badge className={cn("text-[10px] px-1.5 py-0 shrink-0", cfg.className)}>
                      {cfg.label}
                    </Badge>
                  )}
                </label>
              );
            })}
          </div>
        )}
      </div>

      {/* Invite & Share */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Invite friends */}
        <div className="rounded-xl border border-border bg-card overflow-hidden">
          <div className="px-4 py-2.5 border-b border-border bg-muted/30">
            <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
              Invite Friends
            </p>
          </div>
          {friendsNotOnBoard.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-1.5 py-6 text-muted-foreground text-center px-4">
              <UserPlus className="size-5" />
              <p className="text-xs">All your friends are already on this board, or you have no friends yet.</p>
            </div>
          ) : (
            <div className="divide-y divide-border">
              {friendsNotOnBoard.map((f) => (
                <div key={f.user.id} className="flex items-center gap-3 px-4 py-2.5">
                  <div className="size-7 rounded-full bg-primary/10 ring-1 ring-primary/20 flex items-center justify-center overflow-hidden shrink-0">
                    {f.user.avatar_url ? (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img
                        src={f.user.avatar_url}
                        alt={f.user.display_name || f.user.email}
                        className="size-full object-cover"
                        referrerPolicy="no-referrer"
                      />
                    ) : (
                      <span className="text-primary text-[10px] font-bold">
                        {(f.user.display_name || f.user.email).slice(0, 2).toUpperCase()}
                      </span>
                    )}
                  </div>
                  <span className="text-sm truncate flex-1">{f.user.display_name || f.user.email}</span>
                  <Button
                    size="sm"
                    variant="outline"
                    className="h-7 px-2.5 text-xs gap-1 shrink-0"
                    onClick={() => handleInvite(f.user.id)}
                    disabled={inviteFriend.isPending}
                  >
                    <UserPlus className="size-3" /> Invite
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Share link */}
        <div className="rounded-xl border border-border bg-card overflow-hidden">
          <div className="px-4 py-2.5 border-b border-border bg-muted/30">
            <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
              Share Link
            </p>
          </div>
          <div className="p-4 flex flex-col gap-2">
            {shareUrl ? (
              <>
                <div className="flex items-center gap-1.5">
                  <Input
                    readOnly
                    value={shareUrl}
                    className="h-7 text-[10px] font-mono select-all flex-1 bg-background"
                  />
                  <Button
                    type="button"
                    size="sm"
                    className="h-7 px-2 text-xs gap-1 shrink-0"
                    onClick={() => {
                      navigator.clipboard.writeText(shareUrl);
                      toast.success("Copied share link!");
                    }}
                  >
                    <Copy className="size-3" /> Copy
                  </Button>
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-6 text-[10px] text-destructive hover:bg-destructive/10 self-start"
                  onClick={() =>
                    revokeShareToken.mutate(id, {
                      onSuccess: () => toast.success("Share link revoked"),
                      onError: () => toast.error("Failed to revoke"),
                    })
                  }
                  disabled={revokeShareToken.isPending}
                >
                  <X className="size-3 mr-1" /> Revoke Link
                </Button>
              </>
            ) : (
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="h-7 text-xs w-full justify-center gap-1.5"
                onClick={() =>
                  createShareToken.mutate(id, {
                    onSuccess: () => toast.success("Share link created!"),
                    onError: () => toast.error("Failed to create link"),
                  })
                }
                disabled={createShareToken.isPending}
              >
                <LinkIcon className="size-3 text-primary" />
                Create share link
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* Members list */}
      <div className="rounded-xl border border-border bg-card overflow-hidden">
        <div className="px-4 py-2.5 border-b border-border bg-muted/30">
          <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
            Members ({members.length})
          </p>
        </div>
        <div className="divide-y divide-border">
          {members.map((m) => (
            <div key={m.user_id} className="flex items-center gap-3 px-4 py-2.5">
              <MemberAvatar member={m} />
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium truncate">
                  {m.display_name || m.email}
                  {m.user_id === myUserId && (
                    <span className="ml-1.5 text-xs text-muted-foreground font-normal">(you)</span>
                  )}
                </p>
                {m.display_name && (
                  <p className="text-xs text-muted-foreground truncate">{m.email}</p>
                )}
              </div>
              {m.role === "owner" && (
                <Badge variant="secondary" className="text-[10px] px-1.5 py-0.5">
                  Owner
                </Badge>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Personal tasks accordion — collapsed by default (defaultValue=[]) */}
      <Accordion defaultValue={[]} multiple={false} className="rounded-xl border border-border bg-card overflow-hidden">
        <AccordionItem value="my-tasks" className="border-none">
          <AccordionTrigger className="px-4 py-3 hover:no-underline hover:bg-muted/30 transition-colors [&>svg]:text-muted-foreground">
            <div className="flex items-center gap-2">
              <ListTodo className="size-4 text-muted-foreground" />
              <span className="text-sm font-medium">My tasks today</span>
              {(myTasksData?.data?.length ?? 0) > 0 && (
                <Badge variant="secondary" className="text-[10px] px-1.5 py-0 ml-1">
                  {myTasksData!.data.length}
                </Badge>
              )}
            </div>
          </AccordionTrigger>
          <AccordionContent className="px-0 pb-0">
            <div className="border-t border-border">
              {(myTasksData?.data?.length ?? 0) === 0 ? (
                <p className="text-sm text-muted-foreground px-4 py-4 text-center">
                  No pending tasks — you&apos;re all caught up!
                </p>
              ) : (
                <div className="divide-y divide-border">
                  {myTasksData!.data.map((task) => {
                    const cfg = statusConfig[task.status];
                    return (
                      <div key={task.id} className="flex items-center gap-3 px-4 py-2.5">
                        <span className="flex-1 text-sm truncate">{task.title}</span>
                        {cfg && (
                          <Badge className={cn("text-[10px] px-1.5 py-0 shrink-0", cfg.className)}>
                            {cfg.label}
                          </Badge>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  );
}
