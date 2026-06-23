"use client";
import { useEffect, useRef, useState, Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { useVerifyEmail, useResendVerification } from "@/lib/verify-hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { CheckCircle2, Mail, XCircle, Loader2 } from "lucide-react";
import { toast } from "sonner";
import Link from "next/link";

function VerifyWithToken({ token }: { token: string }) {
  const router = useRouter();
  const { login } = useAuth();
  const { mutate: verify, isPending, isError, isSuccess } = useVerifyEmail();
  const ran = useRef(false);

  useEffect(() => {
    if (ran.current) return;
    ran.current = true;
    verify(token, {
      onSuccess(data) {
        login(data.token, { id: data.user.id, email: data.user.email, role: data.user.role ?? "user" });
        setTimeout(() => router.replace("/tasks"), 1500);
      },
    });
  }, [token, verify, login, router]);

  if (isPending) {
    return (
      <div className="flex flex-col items-center gap-3">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
        <p className="text-sm text-muted-foreground">Verifying your email…</p>
      </div>
    );
  }

  if (isSuccess) {
    return (
      <div className="flex flex-col items-center gap-3 text-center">
        <CheckCircle2 className="size-10 text-emerald-500" />
        <h2 className="text-lg font-semibold">Email verified!</h2>
        <p className="text-sm text-muted-foreground">Redirecting you to the app…</p>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex flex-col items-center gap-4 text-center">
        <XCircle className="size-10 text-destructive" />
        <div>
          <h2 className="text-lg font-semibold">Link expired or invalid</h2>
          <p className="text-sm text-muted-foreground mt-1">
            This verification link has already been used or has expired.
          </p>
        </div>
        <Link href="/verify-email" className="text-sm font-medium underline underline-offset-4">
          Request a new link
        </Link>
      </div>
    );
  }

  return null;
}

function RequestResend() {
  const [email, setEmail] = useState("");
  const { mutate: resend, isPending, isSuccess } = useResendVerification();

  const handleResend = () => {
    if (!email) return;
    resend(email, {
      onSuccess() {
        toast.success("Verification email sent — check your inbox");
      },
      onError() {
        toast.error("Something went wrong. Please try again.");
      },
    });
  };

  return (
    <div className="flex flex-col items-center gap-6 text-center">
      <Mail className="size-10 text-muted-foreground" />
      <div>
        <h2 className="text-lg font-semibold">Check your inbox</h2>
        <p className="text-sm text-muted-foreground mt-1 max-w-xs">
          We sent a verification link to your email address. Click it to activate your account.
        </p>
      </div>

      <div className="w-full flex flex-col gap-3">
        <p className="text-xs text-muted-foreground">Didn&apos;t receive it? Resend below.</p>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="resend-email" className="text-left text-xs">Email address</Label>
          <Input
            id="resend-email"
            type="email"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>
        <Button onClick={handleResend} disabled={isPending || !email} variant="outline" className="w-full">
          {isPending ? "Sending…" : "Resend verification email"}
        </Button>
      </div>

      <Link href="/login" className="text-sm text-muted-foreground underline underline-offset-4">
        Back to sign in
      </Link>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense fallback={null}>
      <VerifyEmailInner />
    </Suspense>
  );
}

function VerifyEmailInner() {
  const params = useSearchParams();
  const token = params.get("token");

  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <div className="w-full max-w-sm">
        <div className="bg-card border border-border rounded-xl p-8 shadow-sm animate-in fade-in-0 slide-in-from-bottom-4 duration-500">
          {token ? <VerifyWithToken token={token} /> : <RequestResend />}
        </div>
      </div>
    </div>
  );
}
