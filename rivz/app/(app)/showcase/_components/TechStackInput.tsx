"use client";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Plus, X, Star } from "lucide-react";
import { cn } from "@/lib/utils";
import type { TechItem } from "@/lib/showcase-hooks";

type Props = {
  value: TechItem[];
  onChange: (items: TechItem[]) => void;
};

export function TechStackInput({ value, onChange }: Props) {
  const update = (i: number, patch: Partial<TechItem>) => {
    onChange(value.map((item, idx) => (idx === i ? { ...item, ...patch } : item)));
  };
  const remove = (i: number) => onChange(value.filter((_, idx) => idx !== i));
  const add = () => onChange([...value, { name: "", is_sponsor: false }]);

  return (
    <div className="flex flex-col gap-2">
      {value.map((item, i) => (
        <div key={i} className="flex items-center gap-2">
          <Input
            value={item.name}
            onChange={(e) => update(i, { name: e.target.value })}
            placeholder="e.g. Next.js"
            className="h-8 text-xs flex-1"
          />
          <button
            type="button"
            onClick={() => update(i, { is_sponsor: !item.is_sponsor })}
            title="Mark as sponsor tool (bonus prize eligibility)"
            className={cn(
              "flex items-center gap-1 px-2 py-1 rounded-full text-[10px] font-medium border transition-colors shrink-0",
              item.is_sponsor
                ? "bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30"
                : "border-border text-muted-foreground hover:border-amber-500/40"
            )}
          >
            <Star className={cn("size-3", item.is_sponsor && "fill-current")} />
            Sponsor
          </button>
          <Button type="button" variant="ghost" size="icon-sm" onClick={() => remove(i)}>
            <X className="size-3.5" />
          </Button>
        </div>
      ))}
      <Button type="button" variant="outline" size="sm" className="h-7 text-xs w-fit gap-1" onClick={add}>
        <Plus className="size-3" /> Add tech
      </Button>
    </div>
  );
}
