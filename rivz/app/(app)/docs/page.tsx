"use client";
import { useState, useMemo, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import dynamic from "next/dynamic";
import {
  useNotes,
  useNote,
  useCreateNote,
  useDeleteNote,
  useNoteBacklinks,
  type Note,
} from "@/lib/notes-hooks";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { FileText, Plus, Trash2, ChevronRight, Link2 } from "lucide-react";
import { toast } from "sonner";

const NoteEditor = dynamic(() => import("./_components/NoteEditor"), {
  ssr: false,
  loading: () => <div className="flex-1 animate-pulse bg-muted/30 m-8 rounded-xl" />,
});

export default function DocsPage() {
  return (
    <Suspense fallback={<div className="h-[calc(100vh-7rem)] animate-pulse rounded-xl bg-muted/30" />}>
      <DocsInner />
    </Suspense>
  );
}

function DocsInner() {
  const { data: notes, isLoading } = useNotes();
  const searchParams = useSearchParams();
  // Deep link from global search: /docs?note=<id> seeds the initial selection.
  const [selectedId, setSelectedId] = useState<string | null>(() => searchParams.get("note"));
  const create = useCreateNote();
  const del = useDeleteNote();

  const { data: selected } = useNote(selectedId);
  const { data: backlinks } = useNoteBacklinks(selectedId);

  // Build a parent→children tree from the flat list.
  const roots = useMemo(() => {
    const list = notes ?? [];
    const byParent = new Map<string | null, Note[]>();
    for (const n of list) {
      const key = n.parent_id;
      if (!byParent.has(key)) byParent.set(key, []);
      byParent.get(key)!.push(n);
    }
    return byParent;
  }, [notes]);

  const handleCreate = () => {
    create.mutate(
      { title: "Untitled" },
      { onSuccess: (n) => setSelectedId(n.id) }
    );
  };

  const handleDelete = (id: string) => {
    del.mutate(id, {
      onSuccess: () => {
        if (selectedId === id) setSelectedId(null);
        toast.success("Doc deleted");
      },
    });
  };

  const renderTree = (parentId: string | null, depth: number) => {
    const children = roots.get(parentId) ?? [];
    return children.map((n) => (
      <div key={n.id}>
        <button
          onClick={() => setSelectedId(n.id)}
          style={{ paddingLeft: 8 + depth * 14 }}
          className={cn(
            "group flex items-center gap-1.5 w-full text-left rounded-lg pr-2 py-1.5 text-sm hover:bg-muted transition-colors",
            selectedId === n.id && "bg-muted font-medium"
          )}
        >
          {n.icon ? (
            <span className="text-sm">{n.icon}</span>
          ) : (
            <FileText className="size-3.5 text-muted-foreground shrink-0" />
          )}
          <span className="truncate flex-1">{n.title}</span>
          <Trash2
            className="size-3.5 text-muted-foreground opacity-0 group-hover:opacity-100 hover:text-destructive shrink-0"
            onClick={(e) => {
              e.stopPropagation();
              handleDelete(n.id);
            }}
          />
        </button>
        {renderTree(n.id, depth + 1)}
      </div>
    ));
  };

  return (
    <div className="flex h-[calc(100vh-7rem)] gap-4">
      {/* Tree */}
      <aside className="w-64 shrink-0 flex flex-col rounded-xl border border-border bg-card overflow-hidden">
        <div className="flex items-center justify-between px-3 py-2.5 border-b border-border">
          <span className="text-sm font-semibold">Docs</span>
          <Button size="icon" variant="ghost" className="size-7" onClick={handleCreate}>
            <Plus className="size-4" />
          </Button>
        </div>
        <div className="flex-1 overflow-y-auto p-1.5">
          {isLoading ? (
            <div className="space-y-1 p-1">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="h-7 rounded-lg bg-muted/50 animate-pulse" />
              ))}
            </div>
          ) : (notes?.length ?? 0) === 0 ? (
            <button
              onClick={handleCreate}
              className="flex flex-col items-center gap-2 w-full py-10 text-muted-foreground hover:text-foreground"
            >
              <FileText className="size-6" />
              <span className="text-xs">Create your first doc</span>
            </button>
          ) : (
            renderTree(null, 0)
          )}
        </div>
      </aside>

      {/* Editor */}
      <main className="flex-1 rounded-xl border border-border bg-card overflow-hidden flex flex-col">
        {selected ? (
          <>
            <NoteEditor key={selected.id} note={selected} />
            {(backlinks?.length ?? 0) > 0 && (
              <div className="border-t border-border px-8 py-3">
                <div className="flex items-center gap-1.5 text-xs font-medium text-muted-foreground mb-2">
                  <Link2 className="size-3.5" /> Linked from
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {backlinks!.map((b) => (
                    <button
                      key={b.id}
                      onClick={() => setSelectedId(b.id)}
                      className="inline-flex items-center gap-1 rounded-full border border-border bg-muted/50 px-2.5 py-1 text-xs hover:bg-muted"
                    >
                      <ChevronRight className="size-3" />
                      {b.title}
                    </button>
                  ))}
                </div>
              </div>
            )}
          </>
        ) : (
          <div className="flex flex-1 flex-col items-center justify-center gap-3 text-muted-foreground">
            <FileText className="size-10" />
            <p className="text-sm">Select a doc or create a new one</p>
            <Button onClick={handleCreate} size="sm">
              <Plus className="size-4 mr-1.5" />
              New doc
            </Button>
          </div>
        )}
      </main>
    </div>
  );
}
