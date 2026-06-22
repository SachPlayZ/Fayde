# Full Feature Implementation Plan

## Architecture Overview

- **Backend**: Go, chi, pgx/v5, postgres migrations numbered `000N_*.up/down.sql`, packages follow `model → repository → service → handler` pattern, SSE broker already wired
- **Frontend**: Next.js 19 app router, TanStack Query, zod + react-hook-form, shadcn/ui, lucide-react, sonner toasts, tailwind
- **DB**: Postgres enums (`task_status`, `task_priority`), UUID PKs, migrations auto-run on startup
- **Real-time**: SSE broker fans events per-user and to admins — already used for task CRUD

---

## Phase 1: Database Migrations

Single migration file `0003_features.up.sql` (and matching `.down.sql`) covering all new schema:

```sql
-- 1. Add 'failed' to task_status (already done)
ALTER TYPE task_status ADD VALUE IF NOT EXISTS 'failed';

-- 2. Subtasks
CREATE TABLE subtasks (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id    UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  title      TEXT NOT NULL,
  done       BOOLEAN NOT NULL DEFAULT false,
  position   INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_subtasks_task ON subtasks(task_id);

-- 3. Tags
CREATE TABLE tags (
  id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name    TEXT NOT NULL,
  color   TEXT NOT NULL DEFAULT '#6366f1',
  UNIQUE(user_id, name)
);
CREATE TABLE task_tags (
  task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  tag_id  UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  PRIMARY KEY (task_id, tag_id)
);

-- 4. Task dependencies
CREATE TABLE task_dependencies (
  task_id       UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  depends_on_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  PRIMARY KEY (task_id, depends_on_id)
);

-- 5. Recurring tasks
ALTER TABLE tasks
  ADD COLUMN recurrence      TEXT,         -- 'daily' | 'weekly' | 'monthly' | null
  ADD COLUMN recurrence_end  TIMESTAMPTZ,
  ADD COLUMN parent_task_id  UUID REFERENCES tasks(id) ON DELETE SET NULL;

-- 6. Comments
CREATE TABLE task_comments (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id    UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  user_id    UUID NOT NULL REFERENCES users(id),
  body       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_comments_task ON task_comments(task_id);

-- 7. Assignees
ALTER TABLE tasks ADD COLUMN assignee_id UUID REFERENCES users(id) ON DELETE SET NULL;

-- 8. Notifications
CREATE TABLE notifications (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type       TEXT NOT NULL, -- 'mention' | 'due_reminder' | 'dependency_unblocked' | 'assigned'
  task_id    UUID REFERENCES tasks(id) ON DELETE CASCADE,
  message    TEXT NOT NULL,
  read       BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notifications_user ON notifications(user_id);

-- 9. Manual sort order (for drag-to-reorder)
ALTER TABLE tasks ADD COLUMN sort_order FLOAT8 NOT NULL DEFAULT 0;

-- 10. User preferences
ALTER TABLE users
  ADD COLUMN theme          TEXT    NOT NULL DEFAULT 'system',
  ADD COLUMN digest_enabled BOOLEAN NOT NULL DEFAULT true;
```

---

## Phase 2: Backend — Core Task Features

### 2A. Subtasks

**Package**: `internal/subtasks/`

- `model.go`: `Subtask{ID, TaskID, Title, Done, Position}`
- `repository.go`: `List(taskID)`, `Create(taskID, title)`, `Update(id, done, title)`, `Delete(id)`, `Reorder(taskID, ids[])`
- `service.go`: thin wrapper, validates task ownership via tasks.Service
- `handler.go`:
  - `GET    /tasks/{id}/subtasks`
  - `POST   /tasks/{id}/subtasks`
  - `PATCH  /tasks/{id}/subtasks/{subId}`
  - `DELETE /tasks/{id}/subtasks/{subId}`
  - `PUT    /tasks/{id}/subtasks/order` — body `{ids: []}`

Wire into `server.go` under the auth group.

### 2B. Tags

**Package**: `internal/tags/`

- `model.go`: `Tag{ID, UserID, Name, Color}`
- `repository.go`: `ListByUser`, `Create`, `Delete`, `AddToTask(taskID, tagID)`, `RemoveFromTask`, `ListByTask`
- `handler.go`:
  - `GET    /tags`
  - `POST   /tags`
  - `DELETE /tags/{id}`
  - `POST   /tasks/{id}/tags`   — body `{tag_id}`
  - `DELETE /tasks/{id}/tags/{tagId}`

Extend `tasks.Task` model to include `Tags []Tag` — join in `GetTask` and `ListTasks` via a lateral join or separate query per page.

### 2C. Task Dependencies

**Package**: `internal/dependencies/`

- `model.go`: `Dependency{TaskID, DependsOnID}`
- `repository.go`: `Add(taskID, dependsOnID)`, `Remove`, `ListBlockedBy(taskID)`, `ListBlocking(taskID)`
- `service.go`: cycle detection via BFS before inserting; publish SSE event `dependency_unblocked` + create notification when a blocking task reaches `done`
- `handler.go`:
  - `GET    /tasks/{id}/dependencies`
  - `POST   /tasks/{id}/dependencies` — body `{depends_on_id}`
  - `DELETE /tasks/{id}/dependencies/{depId}`

Extend `tasks.Service.UpdateTask` to check whether completing a task unblocks any dependents and fire SSE + notifications accordingly.

### 2D. Recurring Tasks

Extend existing `tasks` package — no new package needed:

- Add `Recurrence` and `RecurrenceEnd` fields to `CreateRequest` / `UpdateRequest` / `Task`
- In `tasks.Service.UpdateTask`: when status changes to `done` and `recurrence != ""`, clone the task with `due_date` advanced by the interval, `parent_task_id` set to the original, status reset to `todo`
- Cloned task inherits title, description, priority, assignee, tags, and recurrence fields

### 2E. Sort Order / Drag-to-Reorder

- Extend `UpdateRequest` with optional `SortOrder *float64`
- Extend `ListParams` to support `sort=sort_order`
- Add `PUT /tasks/reorder` — body `[{id, sort_order}]`, bulk-updates in a single transaction scoped to `user_id`

---

## Phase 3: Backend — Collaboration Features

### 3A. Comments

**Package**: `internal/comments/`

- `model.go`: `Comment{ID, TaskID, UserID, UserEmail, Body, CreatedAt, UpdatedAt}`
- `repository.go`: `List(taskID)`, `Create`, `Update(id, userID, body)`, `Delete(id, userID)`
- `service.go`: on create, scan body for `@email` patterns → look up user by email → insert notification row → publish SSE event `notification` to that user
- `handler.go`:
  - `GET    /tasks/{id}/comments`
  - `POST   /tasks/{id}/comments`       — body `{body}`
  - `PATCH  /tasks/{id}/comments/{cId}` — body `{body}`
  - `DELETE /tasks/{id}/comments/{cId}`

### 3B. Assignees

Extend `tasks.CreateRequest` and `tasks.UpdateRequest` with `AssigneeID *string`.

Extend task queries to `LEFT JOIN users assignee ON assignee.id = tasks.assignee_id` and return `assignee_email` in responses.

When a task is assigned or reassigned, create a notification for the new assignee and publish SSE.

### 3C. Notifications

**Package**: `internal/notifications/`

- `model.go`: `Notification{ID, UserID, Type, TaskID, Message, Read, CreatedAt}`
- `repository.go`: `ListByUser(userID, unreadOnly bool)`, `MarkRead(id, userID)`, `MarkAllRead(userID)`, `Create`, `UnreadCount(userID)`
- `service.go`: called by comments (mention), tasks (assignment, dependency unblocked), and scheduler (due reminder)
- `handler.go`:
  - `GET  /notifications`              — query param `?unread=true`
  - `PATCH /notifications/{id}/read`
  - `POST  /notifications/read-all`

SSE event type `notification` is already supported by the broker — just needs the notification payload published on the user's channel.

---

## Phase 4: Backend — Notifications & Email

### 4A. Due Date Reminders + Daily Digest

**Package**: `internal/scheduler/`

- `scheduler.go`: goroutine started in `main.go`, runs a ticker every hour
- On each tick:
  1. **Due reminders**: `SELECT * FROM tasks WHERE due_date BETWEEN now() AND now() + interval '24 hours' AND status NOT IN ('done','failed')` — for each, if no `due_reminder` notification exists for that task in the last 23h, insert one and publish SSE
  2. **Daily digest** (gate on current hour == 8): query users with `digest_enabled = true`, for each build a summary of tasks due today / overdue, send email

**Email**: add `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `FROM_EMAIL` to `config.go`. Use Go stdlib `net/smtp`. Build HTML template as a Go `text/template` string.

Wire in `main.go`: `go scheduler.Start(ctx, pool, notificationsSvc, cfg)` after all services are wired.

---

## Phase 5: Backend — Admin Analytics

Extend `internal/admin/handlers.go`:

- `GET /admin/analytics` — returns:

```json
{
  "by_status":           { "todo": 0, "in_progress": 0, "done": 0, "failed": 0 },
  "by_priority":         { "low": 0, "medium": 0, "high": 0 },
  "completion_rate_7d":  [{ "date": "2026-06-15", "done": 0, "created": 0 }],
  "overdue_by_user":     [{ "user_email": "...", "count": 0, "oldest_due": "..." }]
}
```

All computed via SQL aggregation queries — no extra tables needed.

---

## Phase 6: Frontend — New Hooks & API Layer

All new hook files follow the existing pattern in `rivz/lib/`:

| File | Exports |
|------|---------|
| `lib/subtasks-hooks.ts` | `useSubtasks`, `useCreateSubtask`, `useUpdateSubtask`, `useDeleteSubtask`, `useReorderSubtasks` |
| `lib/tags-hooks.ts` | `useTags`, `useCreateTag`, `useDeleteTag`, `useAddTagToTask`, `useRemoveTagFromTask` |
| `lib/comments-hooks.ts` | `useComments`, `useCreateComment`, `useUpdateComment`, `useDeleteComment` |
| `lib/notifications-hooks.ts` | `useNotifications`, `useUnreadCount`, `useMarkRead`, `useMarkAllRead` |
| `lib/dependencies-hooks.ts` | `useTaskDependencies`, `useAddDependency`, `useRemoveDependency` |
| `lib/analytics-hooks.ts` | `useAdminAnalytics` |
| `lib/user-hooks.ts` | `useUpdatePreferences` |

`notifications-hooks.ts` also subscribes to the SSE stream and invalidates the query on `notification` events.

**Type extensions**:

- `lib/tasks-hooks.ts` `Task` type: add `tags`, `assignee_email`, `assignee_id`, `recurrence`, `recurrence_end`, `sort_order`, `subtask_count`, `subtasks_done`
- `lib/schemas.ts` `taskSchema`: add `recurrence`, `recurrence_end`, `assignee_id`
- `lib/admin-hooks.ts` `AdminTask` type: add same new fields

---

## Phase 7: Frontend — Task Form Expansions

All additions live inside the existing `TaskForm.tsx` modal.

### Subtasks section (edit mode only)

- Collapsible, same pattern as Attachments
- Checklist: each subtask row has checkbox + title text + delete button
- Inline "Add subtask" input at the bottom of the list
- Drag handle on each row using `@dnd-kit/sortable` — on drop calls `useReorderSubtasks`
- Progress bar at the top of the section: `{done}/{total}`

### Tags section

- Tag pills displayed in the banner header alongside the status badge
- "Add tag" button opens a `Popover` with:
  - List of existing user tags (color dot + name) — click to toggle on/off task
  - "New tag" inline form: name input + 6 color swatches
- Tag pills in edit mode show `×` to detach

### Assignee field

- Added to the `grid grid-cols-2` form grid as a third full-width row
- `Select` populated from `GET /admin/users` — shows email with avatar initials fallback
- Non-admin users see only themselves in the list

### Recurrence field

- `Select` in the form grid: None / Daily / Weekly / Monthly
- When not None, a `Repeat until` date picker appears below it

### Dependencies section (edit mode only)

- Collapsible section with two sub-lists: "Blocks" and "Blocked by"
- Combobox search (filter user's tasks by title) to add a dependency
- Warning callout at top of form when the task is currently blocked

---

## Phase 8: Frontend — Views & Navigation

### 8A. Kanban Board View

**New file**: `rivz/app/(app)/tasks/_components/KanbanView.tsx`

- 4 columns: Todo / In Progress / Done / Failed — each is a droppable zone via `@dnd-kit/core`
- Dragging a card to a different column fires `useUpdateTask` with the new `status`
- Within-column drag fires `useReorderSubtasks` bulk reorder
- Task card shows: title, priority dot, due date, tag pills, assignee initials, subtask progress `2/5`
- Clicking a card opens the existing `TaskForm` modal

**New file**: `rivz/app/(app)/tasks/_components/ViewToggle.tsx`

- Two icon buttons (`TableIcon` / `KanbanSquare` from lucide) in the filter bar
- View preference persisted to `localStorage`

Update `TasksPageClient.tsx` to switch between `<KanbanView>` and the existing table based on view state.

### 8B. Optimistic Task Creation

Extend `useCreateTask` in `lib/tasks-hooks.ts` with `onMutate`:

```ts
onMutate: async (newTask) => {
  await qc.cancelQueries({ queryKey: ['tasks'] });
  const prev = qc.getQueriesData<TasksResponse>({ queryKey: ['tasks'] });
  const optimistic: Task = {
    id: `optimistic-${Date.now()}`,
    tags: [], subtask_count: 0, subtasks_done: 0,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    ...newTask,
  };
  qc.setQueriesData<TasksResponse>({ queryKey: ['tasks'] }, (old) =>
    old ? { ...old, data: [optimistic, ...old.data], total: old.total + 1 } : old
  );
  return { prev };
},
onError: (_err, _vars, ctx) => {
  ctx?.prev?.forEach(([key, data]) => qc.setQueryData(key, data));
},
```

---

## Phase 9: Frontend — Global UX Features

### 9A. Command Palette (⌘K)

**New file**: `components/CommandPalette.tsx`

- Package: `pnpm add cmdk`
- Opens on `⌘K` / `Ctrl+K` via `useEffect` keydown in a `CommandPaletteProvider` context
- Sections:
  - **Tasks** — fuzzy search loaded tasks; Enter opens `TaskForm` for that task
  - **Actions** — "New task", "Go to Admin", "Toggle theme"
  - **Navigation** — Tasks, Admin
- Wired in `app/(app)/layout.tsx`

### 9B. Keyboard Shortcuts

**New file**: `lib/keyboard-hooks.ts` — `useKeyboardShortcuts(callbacks: Record<string, () => void>)`

Registered in `TasksPageClient.tsx`:

| Key | Action |
|-----|--------|
| `n` | Open new task modal |
| `/` | Focus search input |
| `Escape` | Close open modal |
| `e` (row focused) | Open edit modal |
| `Delete` (row focused) | Prompt delete |

### 9C. Bulk Actions

Add `selected: Set<string>` state to `TasksPageClient.tsx`.

- Checkbox column in each `TaskRow` (visible on hover or when any row is selected)
- "Select all" checkbox in `TableHead`
- Floating action bar fixed at screen bottom when `selected.size > 0`:
  - `{N} selected` label
  - Status `Select` → `POST /tasks/bulk-update`
  - Priority `Select` → same endpoint
  - Delete `Button` → `POST /tasks/bulk-delete`
  - Dismiss button

**Backend** (new routes):
- `POST /tasks/bulk-update` — body `{ids, status?, priority?}`, iterate in a transaction scoped to `user_id`
- `POST /tasks/bulk-delete` — body `{ids}`, same pattern

### 9D. Notifications Bell

**New file**: `components/NotificationsBell.tsx`

- Bell icon added to the app layout header (next to existing `ThemeToggle`)
- Red badge showing unread count from `useUnreadCount`
- Dropdown `Popover` lists latest 10 notifications: icon by type + task title + relative time
- "Mark all read" button
- Clicking a notification navigates to the tasks page and opens the edit modal for that task

### 9E. Theme Persistence Across Devices

`ThemeToggle.tsx` currently saves to `localStorage` only.

Extend: on toggle, also call `PATCH /auth/me/preferences` `{theme}` so the preference persists server-side.

On login, read `user.theme` from `/auth/me` and apply it before first render.

**Backend**: `PATCH /auth/me/preferences` — body `{theme?, digest_enabled?}`, updates `users` table.

---

## Phase 10: Frontend — Admin Analytics

**New file**: `rivz/app/(admin)/admin/_components/AnalyticsTab.tsx`

- Package: `pnpm add recharts`
- New "Analytics" tab added to the existing Tabs component in `admin/page.tsx`
- Charts:
  - **Donut**: tasks by status (4 segments using the existing status colors)
  - **Bar**: tasks by priority
  - **Line**: 7-day completion rate — done vs. created per day
  - **Table**: overdue tasks by user — email, count, oldest due date

All data from `useAdminAnalytics` hitting `GET /admin/analytics`.

---

## Phase 11: Frontend — Comments Panel

Fourth collapsible section in `TaskForm.tsx` (edit mode only), below Activity.

- List: avatar initials circle + user email + relative time + body text
- Markdown-lite rendering: `**bold**`, `_italic_`, `` `code` ``
- `@mention` autocomplete: typing `@` in the textarea opens a floating list of users filtered by the characters that follow; selecting inserts `@email` as text
- Own comments show pencil + trash icons on hover (edit in-place, confirm delete)
- Optimistic insert: comment appears immediately; rolls back on error

---

## Phase 12: Wiring & Integration

### Backend wiring order in `main.go`

1. Wire `notifications` (repo + svc)
2. Wire `subtasks` (repo + svc + handler)
3. Wire `tags` (repo + svc + handler)
4. Wire `comments` (repo + svc + handler, pass `notificationsSvc`)
5. Wire `dependencies` (repo + svc + handler, pass `tasksSvc` + `notificationsSvc`)
6. Pass `notificationsSvc` into `tasksSvc` constructor for assignment notifications
7. Start scheduler goroutine: `go scheduler.Start(ctx, pool, notificationsSvc, cfg)`
8. Register all new routes in `server.go`

### Frontend wiring order

1. Install packages: `pnpm add cmdk recharts @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities`
2. Wrap root layout with `CommandPaletteProvider`
3. Add `NotificationsBell` to `app/(app)/layout.tsx` header
4. Update `TaskForm.tsx` with all new sections (subtasks, tags, assignee, recurrence, dependencies, comments)
5. Add `KanbanView` + `ViewToggle` to tasks page client
6. Add bulk-select state + floating action bar to `TasksPageClient.tsx`
7. Add `AnalyticsTab` to admin page
8. Register keyboard shortcuts in `TasksPageClient.tsx`
9. Wire optimistic create in `useCreateTask`

---

## Unresolved Questions

1. **Assignee visibility** — can regular users assign tasks to any other user, or only to themselves? (Simplest default: only admins can assign to others)
2. **Email delivery** — raw SMTP or a transactional provider (Resend / Postmark)? Raw SMTP needs a real mail server; a provider is easier to test and more reliable in production
3. **@mention target** — by email or by display name? No display name column exists yet; would need adding to `users`
4. **Kanban pagination** — if a column has >50 tasks, load-more or virtual scroll?
5. **Recurring task on failure** — if a recurring task is marked `failed` (not `done`), should a new instance still spawn?
