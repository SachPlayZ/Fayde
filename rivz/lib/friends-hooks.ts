"use client";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export type FriendUser = {
  id: string;
  email: string;
  display_name: string | null;
  avatar_url: string | null;
};

// The backend resolves "the other party" into `user` server-side for both
// /friends (status=accepted) and /friends/requests (status=pending) — same shape.
export type Friend = {
  id: string;
  status: "pending" | "accepted" | "declined";
  direction: "incoming" | "outgoing";
  created_at: string;
  responded_at: string | null;
  user: FriendUser;
};

export type FriendRequest = Friend;

export type UserSearchResult = {
  id: string;
  email: string;
  display_name: string | null;
  avatar_url: string | null;
};

export function useFriends() {
  return useQuery<Friend[]>({
    queryKey: ["friends"],
    queryFn: () => api.get<Friend[]>("/friends"),
  });
}

export function useFriendRequests() {
  return useQuery<FriendRequest[]>({
    queryKey: ["friend-requests"],
    queryFn: () => api.get<FriendRequest[]>("/friends/requests"),
  });
}

export function useSendFriendRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (addressee_email: string) =>
      api.post<FriendRequest>("/friends/requests", { addressee_email }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["friend-requests"] }),
  });
}

export function useAcceptFriendRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api.post<FriendRequest>(`/friends/requests/${id}/accept`, {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["friend-requests"] });
      qc.invalidateQueries({ queryKey: ["friends"] });
    },
  });
}

export function useDeclineFriendRequest() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api.post<void>(`/friends/requests/${id}/decline`, {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["friend-requests"] });
    },
  });
}

export function useRemoveFriend() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/friends/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["friends"] });
      qc.invalidateQueries({ queryKey: ["friend-requests"] });
    },
  });
}

export function useSearchUsers(q: string) {
  return useQuery<UserSearchResult[]>({
    queryKey: ["users-search", q],
    queryFn: () => api.get<UserSearchResult[]>(`/friends/search?q=${encodeURIComponent(q)}`),
    enabled: q.length > 0,
  });
}
