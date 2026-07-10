"use client";
import React, { useEffect, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Loader2, LayoutGrid, Users } from "lucide-react";
import { Button } from "@/components/ui/button";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

type Preview = {
  board_name: string;
  member_count: number;
};

export function JoinBoardClient({ token }: { token: string }) {
  const router = useRouter();
  const [preview, setPreview] = useState<Preview | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [joining, setJoining] = useState(false);
  const [joined, setJoined] = useState(false);

  // Check auth by looking at localStorage — lazy initializer runs once on client
  const [authToken] = useState<string | null>(
    () => (typeof window !== "undefined" ? localStorage.getItem("token") : null)
  );

  useEffect(() => {
    if (!token) return;
    fetch(`${BASE_URL}/boards/join/${token}`)
      .then((res) => {
        if (!res.ok)
          throw new Error(
            res.status === 404 ? "Board not found or link expired" : "Failed to load board"
          );
        return res.json() as Promise<Preview>;
      })
      .then((data) => {
        setPreview(data);
        setLoading(false);
      })
      .catch((err: Error) => {
        setError(err.message);
        setLoading(false);
      });
  }, [token]);

  const handleJoin = async () => {
    if (!authToken) return;
    setJoining(true);
    try {
      const res = await fetch(`${BASE_URL}/boards/join/${token}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${authToken}`,
        },
        body: JSON.stringify({}),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body?.error?.message ?? "Failed to join board");
      }
      const data = (await res.json()) as { board_id: string };
      setJoined(true);
      router.push(`/boards/${data.board_id}`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to join board";
      setError(msg);
    } finally {
      setJoining(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      {/* Header */}
      <header className="border-b border-border bg-background/90 backdrop-blur-md">
        <div className="max-w-2xl mx-auto px-4 h-14 flex items-center">
          <Link href="/" className="flex items-center gap-2">
            <Image src="/logo.png" alt="Fayde" width={24} height={24} className="size-6 rounded-md" />
            <span className="font-bold text-sm tracking-tight">Fayde</span>
          </Link>
        </div>
      </header>

      {/* Content */}
      <main className="flex-1 max-w-2xl w-full mx-auto px-4 py-12">
        {loading ? (
          <div className="flex flex-col items-center justify-center py-24 gap-3">
            <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
            <p className="text-sm text-muted-foreground">Loading board…</p>
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center py-24 gap-3 text-center">
            <LayoutGrid className="size-8 text-muted-foreground" />
            <p className="text-sm font-medium">{error}</p>
            <p className="text-xs text-muted-foreground">
              This link may have expired or be invalid.
            </p>
          </div>
        ) : preview ? (
          <div className="flex flex-col items-center gap-6 animate-in fade-in-0 slide-in-from-bottom-3 duration-400">
            <div className="size-16 rounded-2xl bg-primary/10 ring-2 ring-primary/20 flex items-center justify-center">
              <LayoutGrid className="size-8 text-primary" />
            </div>

            <div className="text-center flex flex-col gap-1">
              <h1 className="text-2xl font-bold tracking-tight">{preview.board_name}</h1>
              <p className="text-sm text-muted-foreground flex items-center justify-center gap-1.5">
                <Users className="size-4" />
                {preview.member_count} {preview.member_count === 1 ? "member" : "members"}
              </p>
            </div>

            <div className="w-full max-w-xs flex flex-col gap-3">
              {joined ? (
                <p className="text-center text-sm text-emerald-600 dark:text-emerald-400 font-medium">
                  You&apos;ve joined! Redirecting…
                </p>
              ) : authToken ? (
                <Button
                  onClick={handleJoin}
                  disabled={joining}
                  className="w-full gap-2"
                  size="lg"
                >
                  {joining && <Loader2 className="size-4 animate-spin" />}
                  Join Board
                </Button>
              ) : (
                <>
                  <p className="text-center text-sm text-muted-foreground">
                    Sign in to join this board.
                  </p>
                  <Button
                    render={<Link href={`/login?redirect=/join/${token}`} />}
                    size="lg"
                    className="w-full"
                  >
                    Log in to join
                  </Button>
                  <Button
                    render={<Link href={`/signup?redirect=/join/${token}`} />}
                    variant="outline"
                    size="lg"
                    className="w-full"
                  >
                    Create an account
                  </Button>
                </>
              )}
            </div>

            <div className="pt-4 border-t border-border w-full max-w-xs text-center">
              <p className="text-xs text-muted-foreground">
                Shared via{" "}
                <span className="font-semibold text-foreground">Fayde</span> — productivity suite
              </p>
            </div>
          </div>
        ) : null}
      </main>
    </div>
  );
}
