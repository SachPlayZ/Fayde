"use client";
import { useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { UploadCloud, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

type Props = {
  label: string;
  currentUrl: string | null;
  aspect?: "square" | "wide";
  onUpload: (file: File) => Promise<unknown>;
  onRemove: () => Promise<unknown>;
};

export function ImageUploadField({ label, currentUrl, aspect = "square", onUpload, onRemove }: Props) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [busy, setBusy] = useState(false);

  const handleFile = async (file: File | undefined) => {
    if (!file) return;
    if (!file.type.startsWith("image/")) {
      toast.error("File must be an image");
      return;
    }
    setBusy(true);
    try {
      await onUpload(file);
      toast.success(`${label} uploaded`);
    } catch {
      toast.error(`Failed to upload ${label.toLowerCase()}`);
    } finally {
      setBusy(false);
      if (inputRef.current) inputRef.current.value = "";
    }
  };

  const handleRemove = async () => {
    setBusy(true);
    try {
      await onRemove();
      toast.success(`${label} removed`);
    } catch {
      toast.error(`Failed to remove ${label.toLowerCase()}`);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex flex-col gap-1.5">
      <Label>{label}</Label>
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={(e) => handleFile(e.target.files?.[0])}
      />
      {currentUrl ? (
        <div className="relative w-fit">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={currentUrl}
            alt={label}
            className={cn(
              "rounded-lg border border-border object-cover",
              aspect === "square" ? "size-20" : "h-20 w-full max-w-sm"
            )}
          />
          <Button
            type="button"
            variant="destructive"
            size="icon-sm"
            className="absolute -top-2 -right-2 size-5 rounded-full"
            onClick={handleRemove}
            disabled={busy}
          >
            <X className="size-3" />
          </Button>
        </div>
      ) : (
        <div
          role="button"
          tabIndex={0}
          onClick={() => inputRef.current?.click()}
          onKeyDown={(e) => e.key === "Enter" && inputRef.current?.click()}
          className={cn(
            "flex flex-col items-center justify-center gap-1 rounded-lg border-2 border-dashed px-4 py-4 cursor-pointer transition-colors select-none border-border hover:border-primary/50 hover:bg-muted/40",
            aspect === "square" ? "size-20" : "h-20 w-full max-w-sm",
            busy && "pointer-events-none opacity-60"
          )}
        >
          <UploadCloud className="size-4 text-muted-foreground" />
          <p className="text-[10px] text-muted-foreground text-center">{busy ? "Uploading…" : "Click to upload"}</p>
        </div>
      )}
    </div>
  );
}
