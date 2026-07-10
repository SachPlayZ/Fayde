"use client";
import { useState } from "react";
import Link from "next/link";
import { useBoards, useCreateBoard, type Board } from "@/lib/boards-hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { toast } from "sonner";
import { LayoutGrid, Plus, Users, ArrowRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { format } from "date-fns";

function BoardCard({ board }: { board: Board }) {
  return (
    <Link
      href={`/boards/${board.id}`}
      className={cn(
        "group rounded-xl border border-border bg-card p-5 flex flex-col gap-3",
        "hover:border-primary/30 hover:bg-accent/30 hover:shadow-sm transition-all duration-200"
      )}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0 flex-1">
          <h3 className="font-semibold truncate group-hover:text-primary transition-colors">
            {board.name}
          </h3>
          {board.description && (
            <p className="text-sm text-muted-foreground mt-0.5 line-clamp-2">
              {board.description}
            </p>
          )}
        </div>
        <ArrowRight className="size-4 text-muted-foreground group-hover:text-primary group-hover:translate-x-0.5 transition-all duration-200 shrink-0 mt-0.5" />
      </div>
      <div className="flex items-center gap-3 text-xs text-muted-foreground">
        <span className="inline-flex items-center gap-1">
          <Users className="size-3.5" />
          {board.member_count} {board.member_count === 1 ? "member" : "members"}
        </span>
        <span>·</span>
        <span>{format(new Date(board.created_at), "MMM d, yyyy")}</span>
      </div>
    </Link>
  );
}

export default function BoardsPage() {
  const { data: boards, isLoading } = useBoards();
  const createBoard = useCreateBoard();
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  const handleCreate = () => {
    if (!name.trim()) return;
    createBoard.mutate(
      { name: name.trim(), description: description.trim() || undefined },
      {
        onSuccess: () => {
          toast.success("Board created");
          setOpen(false);
          setName("");
          setDescription("");
        },
        onError: (err) => toast.error(err.message || "Failed to create board"),
      }
    );
  };

  return (
    <div className="flex flex-col gap-6 max-w-3xl">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="text-xl font-bold tracking-tight">Boards</h2>
          <p className="text-sm text-muted-foreground mt-0.5">
            Shared accountability boards with friends.
          </p>
        </div>
        <Button onClick={() => setOpen(true)} className="gap-1.5 shrink-0">
          <Plus className="size-4" /> New Board
        </Button>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-28 rounded-xl border border-border bg-card animate-pulse" />
          ))}
        </div>
      ) : (boards?.length ?? 0) === 0 ? (
        <div className="flex flex-col items-center justify-center gap-3 py-16 rounded-xl border border-border bg-card text-muted-foreground">
          <LayoutGrid className="size-8" />
          <p className="text-sm">No boards yet. Create one to get started.</p>
          <Button variant="outline" onClick={() => setOpen(true)} className="gap-1.5">
            <Plus className="size-4" /> Create Board
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {boards!.map((b: Board) => (
            <BoardCard key={b.id} board={b} />
          ))}
        </div>
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Create Board</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-3 py-2">
            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium">Name</label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleCreate()}
                placeholder="Morning Routine"
                autoFocus
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <label className="text-sm font-medium text-muted-foreground">
                Description <span className="font-normal">(optional)</span>
              </label>
              <Input
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Daily habits we&apos;re tracking together"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!name.trim() || createBoard.isPending}
            >
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
