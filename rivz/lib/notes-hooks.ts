"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type Note = {
  id: string;
  user_id: string;
  parent_id: string | null;
  title: string;
  body: string;
  is_folder: boolean;
  icon: string | null;
  position: number;
  archived: boolean;
  created_at: string;
  updated_at: string;
};

export type NoteRef = { id: string; title: string; icon: string | null };

export function useNotes() {
  return useQuery<Note[]>({
    queryKey: ["notes"],
    queryFn: () => api.get<Note[]>("/notes"),
  });
}

export function useNote(id: string | null) {
  return useQuery<Note>({
    queryKey: ["note", id],
    queryFn: () => api.get<Note>(`/notes/${id}`),
    enabled: !!id,
  });
}

export function useCreateNote() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { title?: string; parent_id?: string | null; is_folder?: boolean }) =>
      api.post<Note>("/notes", body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["notes"] }),
  });
}

export type NotePatch = {
  title?: string;
  body?: string;
  plain?: string;
  parent_id?: string | null;
  position?: number;
  archived?: boolean;
  icon?: string | null;
};

export function useUpdateNote() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: NotePatch }) =>
      api.patch<Note>(`/notes/${id}`, patch),
    onSuccess: (note) => {
      qc.setQueryData(["note", note.id], note);
      qc.invalidateQueries({ queryKey: ["notes"] });
    },
  });
}

export function useDeleteNote() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/notes/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["notes"] }),
  });
}

export function useReorderNotes() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (items: { id: string; position: number; parent_id: string | null }[]) =>
      api.put<void>("/notes/reorder", items),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["notes"] }),
  });
}

export function useNoteBacklinks(id: string | null) {
  return useQuery<NoteRef[]>({
    queryKey: ["note-backlinks", id],
    queryFn: () => api.get<NoteRef[]>(`/notes/${id}/backlinks`),
    enabled: !!id,
  });
}

export function useTaskNotes(taskId: string | null) {
  return useQuery<NoteRef[]>({
    queryKey: ["task-notes", taskId],
    queryFn: () => api.get<NoteRef[]>(`/tasks/${taskId}/notes`),
    enabled: !!taskId,
  });
}

export function useLinkTaskNote(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ noteId, link }: { noteId: string; link: boolean }) =>
      link
        ? api.post<void>(`/notes/${noteId}/tasks/${taskId}`, {})
        : api.delete<void>(`/notes/${noteId}/tasks/${taskId}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["task-notes", taskId] }),
  });
}
