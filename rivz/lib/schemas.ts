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

export const projectSchema = z.object({
  title: z.string().min(1, "Title required"),
  tagline: z.string().max(200, "Keep it to one sentence"),
  problem: z.string(),
  solution: z.string(),
  tech_stack: z.array(
    z.object({ name: z.string().min(1, "Required"), is_sponsor: z.boolean() })
  ),
  demo_url: z.union([z.string().url("Must be a valid URL"), z.literal("")]).nullable(),
  repo_url: z.union([z.string().url("Must be a valid URL"), z.literal("")]).nullable(),
  live_url: z.union([z.string().url("Must be a valid URL"), z.literal("")]).nullable(),
});

export const leetcodeProblemSchema = z.object({
  title: z.string().min(1, "Title required"),
  lc_number: z.number().int().positive().optional().nullable(),
  slug: z.string().optional().nullable(),
  url: z.union([z.string().url("Must be a valid URL"), z.literal("")]).optional().nullable(),
  difficulty: z.enum(["easy", "medium", "hard"]),
  topics: z.array(z.string()),
  notes: z.string().optional(),
});

export const usernameSchema = z
  .string()
  .min(3, "At least 3 characters")
  .max(50, "At most 50 characters")
  .regex(/^[a-z0-9_-]+$/, "Lowercase letters, numbers, - and _ only");

export type LoginInput = z.infer<typeof loginSchema>;
export type SignupInput = z.infer<typeof signupSchema>;
export type TaskInput = z.infer<typeof taskSchema>;
export type ProjectInput = z.infer<typeof projectSchema>;
export type LeetcodeProblemInput = z.infer<typeof leetcodeProblemSchema>;
