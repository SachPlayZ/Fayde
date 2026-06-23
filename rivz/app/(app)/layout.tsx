"use client";
import { useAuth } from "@/lib/auth-context";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { ThemeToggle } from "@/app/(app)/_components/ThemeToggle";
import { ActivitySidebar } from "@/app/(app)/_components/ActivitySidebar";
import { NotificationsBell } from "@/components/NotificationsBell";
import { CommandPaletteProvider } from "@/components/CommandPalette";
import { QuickCaptureProvider } from "@/lib/quick-capture-context";
import { QuickCaptureDialog } from "@/components/QuickCaptureDialog";
import { PomodoroTimer } from "@/components/PomodoroTimer";
import { LogOut, Activity, ShieldCheck, ClipboardList, FolderKanban, Zap, Settings, FileText, LayoutDashboard, Flame, Target } from "lucide-react";
import { Button, buttonVariants } from "@/components/ui/button";
import Link from "next/link";
import Image from "next/image";
import { PageTracker } from "@/components/PageTracker";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const { user, logout, loading } = useAuth();
  const router = useRouter();
  const [activityOpen, setActivityOpen] = useState(false);

  useEffect(() => {
    if (!loading && !user) router.replace("/login");
  }, [user, loading, router]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="flex flex-col items-center gap-3">
          <div className="size-7 rounded-full border-2 border-foreground/30 border-t-foreground animate-spin" />
          <p className="text-sm text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  if (!user) return null;

  const initials = user.email.slice(0, 2).toUpperCase();

  return (
    <QuickCaptureProvider>
    <CommandPaletteProvider>
    <PageTracker userId={user?.id} />
    <div className="min-h-screen flex flex-col">
      <header className="sticky top-0 z-40 border-b border-border bg-background/90 backdrop-blur-md">
        <div className="max-w-5xl mx-auto px-4 h-14 flex items-center justify-between">
          <Link href="/dashboard" className="flex items-center gap-2">
            <Image src="/logo.png" alt="Fayde" width={24} height={24} className="size-6 rounded-md" priority />
            <span className="font-semibold text-sm tracking-tight">Fayde</span>
          </Link>

          <div className="flex items-center gap-1.5">
            {/* Nav links */}
            <nav className="hidden sm:flex items-center gap-0.5 mr-2">
              <Link
                href="/dashboard"
                className={buttonVariants({ variant: "ghost", size: "sm" }) + " gap-1.5 text-xs"}
              >
                <LayoutDashboard className="w-3.5 h-3.5" />
                Home
              </Link>
              <Link
                href="/tasks"
                className={buttonVariants({ variant: "ghost", size: "sm" }) + " gap-1.5 text-xs"}
              >
                <ClipboardList className="w-3.5 h-3.5" />
                Tasks
              </Link>
              <Link
                href="/habits"
                className={buttonVariants({ variant: "ghost", size: "sm" }) + " gap-1.5 text-xs"}
              >
                <Flame className="w-3.5 h-3.5" />
                Habits
              </Link>
              <Link
                href="/goals"
                className={buttonVariants({ variant: "ghost", size: "sm" }) + " gap-1.5 text-xs"}
              >
                <Target className="w-3.5 h-3.5" />
                Goals
              </Link>
              <Link
                href="/projects"
                className={buttonVariants({ variant: "ghost", size: "sm" }) + " gap-1.5 text-xs"}
              >
                <FolderKanban className="w-3.5 h-3.5" />
                Projects
              </Link>
              <Link
                href="/sprints"
                className={buttonVariants({ variant: "ghost", size: "sm" }) + " gap-1.5 text-xs"}
              >
                <Zap className="w-3.5 h-3.5" />
                Sprints
              </Link>
              <Link
                href="/docs"
                className={buttonVariants({ variant: "ghost", size: "sm" }) + " gap-1.5 text-xs"}
              >
                <FileText className="w-3.5 h-3.5" />
                Docs
              </Link>
              <Link
                href="/settings"
                className={buttonVariants({ variant: "ghost", size: "sm" }) + " gap-1.5 text-xs"}
              >
                <Settings className="w-3.5 h-3.5" />
                Settings
              </Link>
            </nav>

            <span className="text-xs text-muted-foreground hidden sm:block mr-1">
              {user.email}
            </span>
            <div className="flex items-center justify-center w-7 h-7 rounded-full bg-secondary text-secondary-foreground text-xs font-semibold select-none">
              {initials}
            </div>
            {user.role === "admin" && (
              <Link
                href="/admin"
                className={buttonVariants({ variant: "ghost", size: "sm" }) + " gap-1.5 text-xs"}
              >
                <ShieldCheck className="w-3.5 h-3.5" />
                Admin
              </Link>
            )}
            <NotificationsBell />
            <ThemeToggle />
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => setActivityOpen((v) => !v)}
              aria-label="Toggle activity log"
              className={activityOpen ? "bg-muted" : ""}
            >
              <Activity className="w-4 h-4" />
            </Button>
            <Button variant="ghost" size="icon-sm" onClick={logout} aria-label="Sign out">
              <LogOut className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-5xl w-full mx-auto px-4 py-8 animate-in fade-in-0 slide-in-from-bottom-3 duration-400">
        {children}
      </main>

      <ActivitySidebar open={activityOpen} onClose={() => setActivityOpen(false)} />
      <QuickCaptureDialog />
      <PomodoroTimer />
    </div>
    </CommandPaletteProvider>
    </QuickCaptureProvider>
  );
}
