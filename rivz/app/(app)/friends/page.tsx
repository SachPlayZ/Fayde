"use client";
import { useState } from "react";
import {
  useFriends,
  useFriendRequests,
  useSendFriendRequest,
  useAcceptFriendRequest,
  useDeclineFriendRequest,
  useRemoveFriend,
  type Friend,
  type FriendRequest,
} from "@/lib/friends-hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";
import { Users, UserPlus, Check, X, Trash2, Search, Clock } from "lucide-react";

function Avatar({
  displayName,
  avatarUrl,
  email,
}: {
  displayName: string | null;
  avatarUrl: string | null;
  email: string;
}) {
  const initials = displayName
    ? displayName
        .trim()
        .split(/\s+/)
        .map((p) => p[0])
        .filter(Boolean)
        .join("")
        .slice(0, 2)
        .toUpperCase()
    : email.slice(0, 2).toUpperCase();

  return (
    <div className="size-9 rounded-full bg-primary/10 ring-1 ring-primary/20 flex items-center justify-center shrink-0 overflow-hidden">
      {avatarUrl ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img src={avatarUrl} alt={displayName || email} className="size-full object-cover" referrerPolicy="no-referrer" />
      ) : (
        <span className="text-primary text-xs font-bold">{initials}</span>
      )}
    </div>
  );
}

function FriendRow({ friend }: { friend: Friend }) {
  const remove = useRemoveFriend();
  return (
    <div className="flex items-center gap-3 px-4 py-3 hover:bg-muted/30 transition-colors">
      <Avatar displayName={friend.user.display_name} avatarUrl={friend.user.avatar_url} email={friend.user.email} />
      <div className="min-w-0 flex-1">
        <p className="font-medium text-sm truncate">{friend.user.display_name || friend.user.email}</p>
        {friend.user.display_name && (
          <p className="text-xs text-muted-foreground truncate">{friend.user.email}</p>
        )}
      </div>
      <Button
        size="icon"
        variant="ghost"
        className="size-8 text-muted-foreground hover:text-destructive shrink-0"
        onClick={() =>
          remove.mutate(friend.id, {
            onSuccess: () => toast.success("Friend removed"),
            onError: () => toast.error("Failed to remove friend"),
          })
        }
        disabled={remove.isPending}
      >
        <Trash2 className="size-4" />
      </Button>
    </div>
  );
}

function IncomingRequestRow({ req }: { req: FriendRequest }) {
  const accept = useAcceptFriendRequest();
  const decline = useDeclineFriendRequest();
  const name = req.user.display_name || req.user.email;

  return (
    <div className="flex items-center gap-3 px-4 py-3 hover:bg-muted/30 transition-colors">
      <Avatar
        displayName={req.user.display_name}
        avatarUrl={req.user.avatar_url}
        email={req.user.email}
      />
      <div className="min-w-0 flex-1">
        <p className="font-medium text-sm truncate">{name}</p>
        {req.user.display_name && (
          <p className="text-xs text-muted-foreground truncate">{req.user.email}</p>
        )}
      </div>
      <div className="flex items-center gap-1.5 shrink-0">
        <Button
          size="sm"
          className="h-7 px-2.5 text-xs gap-1"
          onClick={() =>
            accept.mutate(req.id, {
              onSuccess: () => toast.success(`Accepted ${name}`),
              onError: () => toast.error("Failed to accept"),
            })
          }
          disabled={accept.isPending}
        >
          <Check className="size-3.5" /> Accept
        </Button>
        <Button
          size="sm"
          variant="ghost"
          className="h-7 px-2.5 text-xs gap-1 text-muted-foreground hover:text-destructive"
          onClick={() =>
            decline.mutate(req.id, {
              onSuccess: () => toast.success("Declined"),
              onError: () => toast.error("Failed to decline"),
            })
          }
          disabled={decline.isPending}
        >
          <X className="size-3.5" /> Decline
        </Button>
      </div>
    </div>
  );
}

function OutgoingRequestRow({ req }: { req: FriendRequest }) {
  const remove = useRemoveFriend();
  const name = req.user.display_name || req.user.email;

  return (
    <div className="flex items-center gap-3 px-4 py-3 hover:bg-muted/30 transition-colors">
      <Avatar
        displayName={req.user.display_name}
        avatarUrl={req.user.avatar_url}
        email={req.user.email}
      />
      <div className="min-w-0 flex-1">
        <p className="font-medium text-sm truncate">{name}</p>
        {req.user.display_name && (
          <p className="text-xs text-muted-foreground truncate">{req.user.email}</p>
        )}
        <p className="text-xs text-muted-foreground flex items-center gap-1 mt-0.5">
          <Clock className="size-3" /> Pending
        </p>
      </div>
      <Button
        size="sm"
        variant="ghost"
        className="h-7 px-2.5 text-xs gap-1 text-muted-foreground hover:text-destructive shrink-0"
        onClick={() =>
          remove.mutate(req.id, {
            onSuccess: () => toast.success("Request cancelled"),
            onError: () => toast.error("Failed to cancel"),
          })
        }
        disabled={remove.isPending}
      >
        <X className="size-3.5" /> Cancel
      </Button>
    </div>
  );
}

export default function FriendsPage() {
  const { data: friends, isLoading: loadingFriends } = useFriends();
  const { data: requests, isLoading: loadingRequests } = useFriendRequests();
  const sendRequest = useSendFriendRequest();
  const [email, setEmail] = useState("");

  const incoming = requests?.filter((r) => r.direction === "incoming") ?? [];
  const outgoing = requests?.filter((r) => r.direction === "outgoing") ?? [];

  const handleSend = () => {
    const trimmed = email.trim();
    if (!trimmed) return;
    sendRequest.mutate(trimmed, {
      onSuccess: () => {
        toast.success(`Friend request sent to ${trimmed}`);
        setEmail("");
      },
      onError: (err) => toast.error(err.message || "Failed to send request"),
    });
  };

  const isLoading = loadingFriends || loadingRequests;

  return (
    <div className="flex flex-col gap-6 max-w-2xl">
      <div>
        <h2 className="text-xl font-bold tracking-tight">Friends</h2>
        <p className="text-sm text-muted-foreground mt-0.5">
          Connect with others and collaborate on boards.
        </p>
      </div>

      {/* Search + Send */}
      <div className="flex items-center gap-2 rounded-xl border border-border bg-card p-2">
        <Search className="size-4 text-muted-foreground ml-1 shrink-0" />
        <Input
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSend()}
          placeholder="Search by email to add a friend…"
          className="flex-1 border-0 shadow-none focus-visible:ring-0"
        />
        <Button
          onClick={handleSend}
          disabled={!email.trim() || sendRequest.isPending}
          className="gap-1.5"
        >
          <UserPlus className="size-4" /> Add
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-16 rounded-xl border border-border bg-card animate-pulse" />
          ))}
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          {/* Incoming requests */}
          {incoming.length > 0 && (
            <div className="rounded-xl border border-border bg-card overflow-hidden">
              <div className="px-4 py-2.5 border-b border-border bg-muted/30">
                <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                  Incoming requests ({incoming.length})
                </p>
              </div>
              <div className="divide-y divide-border">
                {incoming.map((req) => (
                  <IncomingRequestRow key={req.id} req={req} />
                ))}
              </div>
            </div>
          )}

          {/* Outgoing requests */}
          {outgoing.length > 0 && (
            <div className="rounded-xl border border-border bg-card overflow-hidden">
              <div className="px-4 py-2.5 border-b border-border bg-muted/30">
                <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                  Sent requests ({outgoing.length})
                </p>
              </div>
              <div className="divide-y divide-border">
                {outgoing.map((req) => (
                  <OutgoingRequestRow key={req.id} req={req} />
                ))}
              </div>
            </div>
          )}

          {/* Friends list */}
          <div className="rounded-xl border border-border bg-card overflow-hidden">
            <div className="px-4 py-2.5 border-b border-border bg-muted/30">
              <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                Friends ({friends?.length ?? 0})
              </p>
            </div>
            {(friends?.length ?? 0) === 0 ? (
              <div className="flex flex-col items-center justify-center gap-3 py-10 text-muted-foreground">
                <Users className="size-7" />
                <p className="text-sm">No friends yet. Send a request above.</p>
              </div>
            ) : (
              <div className="divide-y divide-border">
                {friends!.map((f: Friend) => (
                  <FriendRow key={f.id} friend={f} />
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
