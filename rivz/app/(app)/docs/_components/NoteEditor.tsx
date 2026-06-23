"use client";
import { useMemo, useRef, useState } from "react";
import { useCreateBlockNote } from "@blocknote/react";
import { BlockNoteView } from "@blocknote/mantine";
import type { Block, PartialBlock } from "@blocknote/core";
import "@blocknote/core/fonts/inter.css";
import "@blocknote/mantine/style.css";
import { useUpdateNote, type Note } from "@/lib/notes-hooks";
import { useTheme } from "next-themes";

// Flatten BlockNote blocks into plain text for search indexing.
function blocksToPlain(blocks: Block[]): string {
  const parts: string[] = [];
  const walk = (bs: Block[]) => {
    for (const b of bs) {
      if (Array.isArray(b.content)) {
        for (const c of b.content) {
          if (c && typeof c === "object" && "text" in c && typeof c.text === "string") {
            parts.push(c.text);
          }
        }
      }
      if (b.children?.length) walk(b.children);
    }
  };
  walk(blocks);
  return parts.join(" ");
}

function parseInitial(body: string): PartialBlock[] | undefined {
  if (!body) return undefined;
  try {
    const parsed = JSON.parse(body);
    return Array.isArray(parsed) && parsed.length ? parsed : undefined;
  } catch {
    return undefined;
  }
}

export default function NoteEditor({ note }: { note: Note }) {
  const update = useUpdateNote();
  const { resolvedTheme } = useTheme();
  const [title, setTitle] = useState(note.title);
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [saving, setSaving] = useState(false);

  // Recreate the editor when the selected note changes.
  const initialContent = useMemo(() => parseInitial(note.body), [note.id]); // eslint-disable-line react-hooks/exhaustive-deps
  const editor = useCreateBlockNote({ 
    initialContent,
    uploadFile: async (file: File) => {
      const body = new FormData();
      body.append("file", file);

      const token = localStorage.getItem("token");
      const apiURL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
      const res = await fetch(`${apiURL}/notes/images`, {
        method: "POST",
        headers: {
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: body,
      });

      if (!res.ok) {
        throw new Error("Failed to upload image");
      }

      const data = await res.json();
      return `${apiURL}${data.url}`;
    }
  }, [note.id]);

  const scheduleSave = (patch: { title?: string; body?: string; plain?: string }) => {
    setSaving(true);
    if (saveTimer.current) clearTimeout(saveTimer.current);
    saveTimer.current = setTimeout(() => {
      update.mutate(
        { id: note.id, patch },
        { onSettled: () => setSaving(false) }
      );
    }, 700);
  };

  const onBodyChange = () => {
    const blocks = editor.document;
    scheduleSave({
      body: JSON.stringify(blocks),
      plain: blocksToPlain(blocks),
    });
  };

  const onTitleChange = (v: string) => {
    setTitle(v);
    scheduleSave({ title: v || "Untitled" });
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-8 pt-8 pb-2">
        <input
          value={title}
          onChange={(e) => onTitleChange(e.target.value)}
          placeholder="Untitled"
          className="text-3xl font-bold tracking-tight bg-transparent outline-none w-full placeholder:text-muted-foreground/40"
        />
        <span className="text-xs text-muted-foreground shrink-0 ml-4">
          {saving ? "Saving…" : "Saved"}
        </span>
      </div>
      <div className="flex-1 overflow-y-auto pb-16">
        <BlockNoteView
          editor={editor}
          theme={resolvedTheme === "dark" ? "dark" : "light"}
          onChange={onBodyChange}
        />
      </div>
    </div>
  );
}
