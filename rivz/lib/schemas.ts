import { z } from "zod";

export const loginSchema = z.object({
  email: z.string().email("Invalid email address"),
  password: z.string().min(6),
});

export const signupSchema = z.object({
  email: z.string().email("Invalid email address"),
  password: z.string().min(8, "Password must be at least 8 characters"),
  display_name: z.string().min(1, "Display name is required"),
});

export const taskSchema = z.object({
  title: z.string().min(1, "Title required"),
  description: z.string().optional(),
  status: z.enum(["todo", "in_progress", "done", "failed"]).optional(),
  priority: z.enum(["low", "medium", "high"]).optional(),
  due_date: z.string().optional().nullable(),
  recurrence: z.enum(["daily", "weekly", "monthly"]).optional().nullable(),
  recurrence_end: z.string().optional().nullable(),
  assignee_id: z.string().optional().nullable(),
});

export type LoginInput = z.infer<typeof loginSchema>;
export type SignupInput = z.infer<typeof signupSchema>;
export type TaskInput = z.infer<typeof taskSchema>;
