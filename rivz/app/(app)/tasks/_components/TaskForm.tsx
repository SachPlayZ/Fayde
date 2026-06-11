"use client";
import { useRef, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { format, parseISO, formatDistanceToNow } from "date-fns";
import { taskSchema, type TaskInput } from "@/lib/schemas";
import { useCreateTask, useUpdateTask, type Task } from "@/lib/tasks-hooks";
import { useTaskActivity } from "@/lib/activity-hooks";
import {
  useAttachments,
  useUploadAttachment,
  useDeleteAttachment,
} from "@/lib/attachments-hooks";
import { ApiError } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { CalendarIcon, ChevronDown, ChevronUp, Paperclip, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

type TaskFormProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  task?: Task;
};

export function TaskForm({ open, onOpenChange, task }: TaskFormProps) {
  const isEdit = !!task;
  const createTask = useCreateTask();
  const updateTask = useUpdateTask();

  const initialDate = task?.due_date ? parseISO(task.due_date) : undefined;
  const [selectedDate, setSelectedDate] = useState<Date | undefined>(initialDate);
  const [calendarOpen, setCalendarOpen] = useState(false);
  const [attachmentsOpen, setAttachmentsOpen] = useState(false);
  const [activityOpen, setActivityOpen] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<TaskInput>({
    resolver: zodResolver(taskSchema),
    defaultValues: {
      title: task?.title ?? "",
      description: task?.description ?? "",
      status: task?.status ?? "todo",
      priority: task?.priority ?? "medium",
      due_date: task?.due_date ? task.due_date.slice(0, 10) : null,
    },
  });

  const statusValue = watch("status");
  const priorityValue = watch("priority");

  // Attachments (only fetched when section is open and editing).
  const { data: attachments = [], isLoading: attachmentsLoading } = useAttachments(
    task?.id ?? "",
    isEdit && attachmentsOpen
  );
  const uploadAttachment = useUploadAttachment(task?.id ?? "");
  const deleteAttachment = useDeleteAttachment(task?.id ?? "");

  // Activity log (only fetched when section is open and editing).
  const { data: activityLogs = [], isLoading: activityLoading } = useTaskActivity(
    task?.id ?? "",
    isEdit && activityOpen
  );

  const handleDateSelect = (date: Date | undefined) => {
    setSelectedDate(date);
    setValue("due_date", date ? format(date, "yyyy-MM-dd") : null);
    setCalendarOpen(false);
  };

  const onSubmit = async (data: TaskInput) => {
    try {
      const payload = {
        ...data,
        due_date: data.due_date ? `${data.due_date}T00:00:00Z` : null,
      };
      if (isEdit && task) {
        await updateTask.mutateAsync({ id: task.id, ...payload });
        toast.success("Task updated");
      } else {
        await createTask.mutateAsync(payload);
        toast.success("Task created");
      }
      handleClose();
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Something went wrong");
    }
  };

  const handleClose = () => {
    reset();
    setSelectedDate(initialDate);
    setAttachmentsOpen(false);
    setActivityOpen(false);
    onOpenChange(false);
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    try {
      await uploadAttachment.mutateAsync(file);
      toast.success("File uploaded");
    } catch {
      toast.error("Upload failed");
    }
    // Reset input so the same file can be re-uploaded if needed.
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const handleDeleteAttachment = async (attId: string) => {
    try {
      await deleteAttachment.mutateAsync(attId);
      toast.success("Attachment deleted");
    } catch {
      toast.error("Failed to delete attachment");
    }
  };

  const actionBadgeVariant = (action: string): "default" | "secondary" | "destructive" | "outline" => {
    switch (action) {
      case "created": return "default";
      case "updated": return "secondary";
      case "deleted": return "destructive";
      default: return "outline";
    }
  };

  const formatChanges = (changes: Record<string, unknown> | null): string => {
    if (!changes || Object.keys(changes).length === 0) return "";
    return Object.entries(changes)
      .map(([key, val]) => {
        const pair = val as [unknown, unknown];
        return `${key}: ${pair[0]} → ${pair[1]}`;
      })
      .join(", ");
  };

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) handleClose(); else onOpenChange(true); }}>
      <DialogContent className="sm:max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit task" : "New task"}</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4 mt-1">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="title">
              Title <span className="text-destructive">*</span>
            </Label>
            <Input
              id="title"
              placeholder="What needs to be done?"
              aria-invalid={!!errors.title}
              {...register("title")}
            />
            {errors.title && (
              <p className="text-xs text-destructive">{errors.title.message}</p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              placeholder="Add more details (optional)"
              rows={3}
              {...register("description")}
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <Label>Status</Label>
              <Select
                value={statusValue}
                onValueChange={(val) => setValue("status", val as TaskInput["status"])}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value="todo">Todo</SelectItem>
                    <SelectItem value="in_progress">In Progress</SelectItem>
                    <SelectItem value="done">Done</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <div className="flex flex-col gap-1.5">
              <Label>Priority</Label>
              <Select
                value={priorityValue}
                onValueChange={(val) => setValue("priority", val as TaskInput["priority"])}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value="low">Low</SelectItem>
                    <SelectItem value="medium">Medium</SelectItem>
                    <SelectItem value="high">High</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label>Due Date</Label>
            <div className="flex gap-2">
              <Popover open={calendarOpen} onOpenChange={setCalendarOpen}>
                <PopoverTrigger asChild>
                  <Button
                    type="button"
                    variant="outline"
                    className={cn(
                      "flex-1 justify-start text-left font-normal",
                      !selectedDate && "text-muted-foreground"
                    )}
                  >
                    <CalendarIcon className="mr-2 h-4 w-4 shrink-0" />
                    {selectedDate ? format(selectedDate, "PPP") : "Pick a date"}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    mode="single"
                    selected={selectedDate}
                    onSelect={handleDateSelect}
                  />
                </PopoverContent>
              </Popover>
              {selectedDate && (
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => handleDateSelect(undefined)}
                  aria-label="Clear date"
                >
                  <X className="h-4 w-4" />
                </Button>
              )}
            </div>
          </div>

          {/* Attachments section — only in edit mode */}
          {isEdit && (
            <div className="flex flex-col gap-2 border rounded-md overflow-hidden">
              <button
                type="button"
                className="flex items-center justify-between px-3 py-2 text-sm font-medium bg-muted/50 hover:bg-muted transition-colors w-full text-left"
                onClick={() => setAttachmentsOpen((v) => !v)}
              >
                <span className="flex items-center gap-2">
                  <Paperclip className="h-3.5 w-3.5" />
                  Attachments
                  {attachments.length > 0 && (
                    <Badge variant="secondary" className="text-xs">{attachments.length}</Badge>
                  )}
                </span>
                {attachmentsOpen ? (
                  <ChevronUp className="h-4 w-4" />
                ) : (
                  <ChevronDown className="h-4 w-4" />
                )}
              </button>

              {attachmentsOpen && (
                <div className="px-3 pb-3 flex flex-col gap-2">
                  {attachmentsLoading ? (
                    <p className="text-xs text-muted-foreground">Loading...</p>
                  ) : attachments.length === 0 ? (
                    <p className="text-xs text-muted-foreground">No attachments yet.</p>
                  ) : (
                    <ul className="flex flex-col gap-1">
                      {attachments.map((att) => (
                        <li key={att.id} className="flex items-center justify-between gap-2 text-sm">
                          <a
                            href={att.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="truncate text-primary hover:underline"
                          >
                            {att.filename}
                          </a>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon-sm"
                            onClick={() => handleDeleteAttachment(att.id)}
                            aria-label="Delete attachment"
                            className="shrink-0 text-muted-foreground hover:text-destructive"
                          >
                            <X className="h-3.5 w-3.5" />
                          </Button>
                        </li>
                      ))}
                    </ul>
                  )}
                  <div>
                    <input
                      ref={fileInputRef}
                      type="file"
                      className="hidden"
                      onChange={handleFileChange}
                    />
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => fileInputRef.current?.click()}
                      disabled={uploadAttachment.isPending}
                    >
                      {uploadAttachment.isPending ? "Uploading..." : "Upload file"}
                    </Button>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Activity section — only in edit mode */}
          {isEdit && (
            <div className="flex flex-col gap-2 border rounded-md overflow-hidden">
              <button
                type="button"
                className="flex items-center justify-between px-3 py-2 text-sm font-medium bg-muted/50 hover:bg-muted transition-colors w-full text-left"
                onClick={() => setActivityOpen((v) => !v)}
              >
                <span>Activity</span>
                {activityOpen ? (
                  <ChevronUp className="h-4 w-4" />
                ) : (
                  <ChevronDown className="h-4 w-4" />
                )}
              </button>

              {activityOpen && (
                <div className="px-3 pb-3">
                  {activityLoading ? (
                    <p className="text-xs text-muted-foreground">Loading...</p>
                  ) : activityLogs.length === 0 ? (
                    <p className="text-xs text-muted-foreground">No activity yet.</p>
                  ) : (
                    <ul className="flex flex-col gap-2">
                      {activityLogs.map((log) => {
                        const changesStr = formatChanges(log.changes);
                        return (
                          <li key={log.id} className="flex flex-col gap-0.5 text-sm">
                            <div className="flex items-center gap-2">
                              <Badge variant={actionBadgeVariant(log.action)} className="capitalize text-xs">
                                {log.action}
                              </Badge>
                              <span className="text-xs text-muted-foreground">
                                {formatDistanceToNow(new Date(log.created_at), { addSuffix: true })}
                              </span>
                            </div>
                            {changesStr && (
                              <p className="text-xs text-muted-foreground pl-1">{changesStr}</p>
                            )}
                          </li>
                        );
                      })}
                    </ul>
                  )}
                </div>
              )}
            </div>
          )}

          <DialogFooter className="mt-2 gap-2">
            <Button type="button" variant="outline" onClick={handleClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting
                ? isEdit ? "Saving..." : "Creating..."
                : isEdit ? "Save changes" : "Create task"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
