"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type Board = {
  id: string;
  owner_id: string;
  name: string;
  description: string;
  created_at: string;
  member_count: number;
};

export type BoardMember = {
  user_id: string;
  email: string;
  display_name: string | null;
  avatar_url: string | null;
  role: "owner" | "member";
  joined_at: string;
};

export type BoardTask = {
  id: string;
  board_id: string;
  title: string;
  sort_order: number;
  created_by: string;
  created_at: string;
};

export type BoardTaskCompletion = {
  board_task_id: string;
  user_id: string;
  completion_date: string;
  completed_at: string;
};

export type BoardSharedTask = {
  board_id: string;
  task_id: string;
  shared_by: string;
  shared_at: string;
  title: string;
  description: string;
  status: "todo" | "in_progress" | "done" | "failed";
  priority: "low" | "medium" | "high";
  due_date: string | null;
  owner_email: string;
  owner_display_name: string | null;
  owner_avatar_url: string | null;
};

export type BoardDetail = {
  id: string;
  owner_id: string;
  name: string;
  description: string;
  created_at: string;
  members: BoardMember[];
  tasks: BoardTask[];
  completions: BoardTaskCompletion[];
  shared_tasks: BoardSharedTask[];
  share_token: string | null;
};

export type BoardJoinPreview = {
  board_name: string;
  member_count: number;
};

export function useBoards() {
  return useQuery<Board[]>({
    queryKey: ["boards"],
    queryFn: () => api.get<Board[]>("/boards"),
  });
}

export function useBoard(id: string) {
  return useQuery<BoardDetail>({
    queryKey: ["board", id],
    queryFn: () => api.get<BoardDetail>(`/boards/${id}`),
    enabled: !!id,
  });
}

export function useCreateBoard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { name: string; description?: string }) =>
      api.post<Board>("/boards", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["boards"] }),
  });
}

export function useUpdateBoard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: { name?: string; description?: string } }) =>
      api.patch<Board>(`/boards/${id}`, patch),
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: ["boards"] });
      qc.invalidateQueries({ queryKey: ["board", id] });
    },
  });
}

export function useDeleteBoard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/boards/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["boards"] }),
  });
}

export function useAddBoardTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ boardId, title }: { boardId: string; title: string }) =>
      api.post<BoardTask>(`/boards/${boardId}/tasks`, { title }),
    onSuccess: (_data, { boardId }) =>
      qc.invalidateQueries({ queryKey: ["board", boardId] }),
  });
}

export function useDeleteBoardTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ boardId, taskId }: { boardId: string; taskId: string }) =>
      api.delete<void>(`/boards/${boardId}/tasks/${taskId}`),
    onSuccess: (_data, { boardId }) =>
      qc.invalidateQueries({ queryKey: ["board", boardId] }),
  });
}

export function useCompleteBoardTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ boardId, taskId }: { boardId: string; taskId: string }) =>
      api.post<void>(`/boards/${boardId}/tasks/${taskId}/complete`, {}),
    onSuccess: (_data, { boardId }) =>
      qc.invalidateQueries({ queryKey: ["board", boardId] }),
  });
}

export function useUncompleteBoardTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ boardId, taskId }: { boardId: string; taskId: string }) =>
      api.delete<void>(`/boards/${boardId}/tasks/${taskId}/complete`),
    onSuccess: (_data, { boardId }) =>
      qc.invalidateQueries({ queryKey: ["board", boardId] }),
  });
}

export function useInviteFriendToBoard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ boardId, friend_user_id }: { boardId: string; friend_user_id: string }) =>
      api.post<void>(`/boards/${boardId}/invite`, { friend_user_id }),
    onSuccess: (_data, { boardId }) =>
      qc.invalidateQueries({ queryKey: ["board", boardId] }),
  });
}

export function useCreateBoardShareToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (boardId: string) =>
      api.post<{ token: string }>(`/boards/${boardId}/share`, {}),
    onSuccess: (_data, boardId) =>
      qc.invalidateQueries({ queryKey: ["board", boardId] }),
  });
}

export function useRevokeBoardShareToken() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (boardId: string) => api.delete<void>(`/boards/${boardId}/share`),
    onSuccess: (_data, boardId) =>
      qc.invalidateQueries({ queryKey: ["board", boardId] }),
  });
}

export function useJoinBoardPreview(token: string) {
  return useQuery<BoardJoinPreview>({
    queryKey: ["board-join-preview", token],
    queryFn: () => api.get<BoardJoinPreview>(`/boards/join/${token}`),
    enabled: !!token,
  });
}

export function useJoinBoard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (token: string) =>
      api.post<{ board_id: string }>(`/boards/join/${token}`, {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["boards"] }),
  });
}

export function useShareTaskToBoard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ boardId, taskId }: { boardId: string; taskId: string }) =>
      api.post<{ status: string }>(`/boards/${boardId}/shared-tasks`, { task_id: taskId }),
    onSuccess: (_data, { boardId, taskId }) => {
      qc.invalidateQueries({ queryKey: ["board", boardId] });
      qc.invalidateQueries({ queryKey: ["task-boards", taskId] });
    },
  });
}

export function useUnshareTaskFromBoard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ boardId, taskId }: { boardId: string; taskId: string }) =>
      api.delete<void>(`/boards/${boardId}/shared-tasks/${taskId}`),
    onSuccess: (_data, { boardId, taskId }) => {
      qc.invalidateQueries({ queryKey: ["board", boardId] });
      qc.invalidateQueries({ queryKey: ["task-boards", taskId] });
    },
  });
}
