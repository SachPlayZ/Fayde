# Productivity Suite — Tier 1→4 Implementation Plan

## Progress
- ✅ **Step 0 — shared notif infra + web push** (DONE, verified): migration `0007_notif_delivery`, `internal/webpush/` pkg (VAPID send + subscriptions), unified `notifications.Deliver()` dispatcher fanning to in-app SSE + email (Resend) + web push + Slack/Discord, per-channel `users.notif_prefs` + chat target, `PATCH /auth/me/preferences` extended, Settings → Notifications tab, service worker push/click handlers. Endpoints smoke-tested green (pubkey, prefs roundtrip, subscribe persists). VAPID keys in `.env`.
- ✅ **Step 1 — Notes/Docs (1A) + Global Search (2C)** (DONE, verified): migrations `0008_notes` (notes tree + tsvector + note↔task + backlinks), `0009_search` (tsv on tasks/task_comments). `internal/notes/` full CRUD + reorder + backlinks ([[uuid]] regex) + task links; `internal/search/` websearch_to_tsquery UNION across tasks/notes/comments with ts_headline + ts_rank. Frontend: BlockNote editor (`/docs` two-pane tree+editor, debounced autosave, derived plaintext), `lib/notes-hooks.ts`, `lib/search-hooks.ts`, CommandPalette upgraded to live server search (tasks/docs/comments, deep-link `/docs?note=`), Docs nav link. Backend smoke green (search ranks+highlights, linking). `next build` green (also fixed pre-existing missing Suspense on /login,/verify-email,/oauth-callback).
- ✅ **Step 2 — Habits (1C) + Dashboard (2D)** (DONE, verified): migration `0010_habits` (habits + habit_logs). `internal/habits/` CRUD + toggle + streak calc (current/longest consecutive days, done_today). `internal/dashboard/` aggregate (due_today/overdue/upcoming/completed_this_week/time_this_week/pomodoros_today + habit streaks). Frontend: `/habits` (7-day toggle grid + streak flames + color picker), `/dashboard` (stat cards + task panels + habit widget), home redirects to /dashboard, nav + palette links. Backend smoke green (streak=2 after 2-day toggle). `next build` green.
- ✅ **Step 3 — Goals/OKRs (1B)** (DONE, verified): migration `0011_goals` (goals + key_results + tasks.goal_id). `internal/goals/` CRUD + KR CRUD + task linking (sets tasks.goal_id, avoids threading goal_id through tasks scans) + progress rollup (avg of KR %; task_completion KRs derive live from linked task done/total). Frontend: `/goals` (progress rings, status select, inline KR editor, create), nav + palette. Backend smoke green (progress=50%, task_completion auto-derives 1/2). `next build` green (18 routes).
- ✅ **Step 4 — Reminders+snooze (2B) + AI plan-day (3D)** (DONE, verified): migration `0012_reminders` (reminders + notifications.snoozed_until). `internal/reminders/` CRUD + DuePending(marks sent); scheduler gained a 1-min ticker firing due reminders via notifSvc.Create (rides the unified dispatcher). notifications: Snooze + list/count filter snoozed. groq `/ai/plan-day` (time-boxed schedule from open tasks). Frontend: reminders section in TaskForm, snooze buttons (1h/Tomorrow) in NotificationsBell, "Plan my day" dialog on dashboard. Smoke green (snooze 1→0, reminder create 201, due rows selectable). `next build` green.
- ✅ **Step 5a — Calendar view (1D phase 1)** (DONE, verified): `CalendarView.tsx` month grid (date-fns), tasks placed by due_date, dnd-kit drag-to-reschedule → useUpdateTask, month nav. Added `calendar` to ViewToggle, wired in TasksPageClient. `next build` green (18 routes).
- ⏳ **Step 5b — Google Calendar push (1D phase 2)**: DEFERRED — needs live Google OAuth consent + Calendar API which can't be verified in this env. Scaffolding pending.
- ✅ **Step 6 — Automations (2A) + Slack/Discord (3B)** (DONE, verified): migration `0013_automations`. `internal/automations/` rules (trigger/conditions/actions JSONB) + engine `OnTaskEvent` hooked into tasks.Service via primitive `AutomationEngine` interface (no import cycle); actions set_status/set_priority (direct SQL, loop-safe) / notify / webhook (Slack+Discord). Frontend: Automations settings tab (rule builder + enable toggle). Smoke green (status→done auto-set priority high→low + notify fired). 3B chat = notif channel (Step 0) + automation webhook action. `next build` green.
- ✅ **Step 7 — Email-to-task (3A)** (DONE, verified): migration `0014_inbox` (users.inbox_token, backfilled + unique). `internal/inbox/` public `POST /webhooks/email` parses Resend inbound (defensive: top-level or data-nested, to as string/array), matches `u+<token>@` → creates task (subject→title, body→desc). inbox_token exposed in /auth/me; settings shows the address w/ copy. Smoke green (webhook 200 → "Buy groceries" task created). `next build` green.
- ⏳ **Step 7b — Mobile audit (3E)**: light — responsive classes already pervasive; deep pass deferred.
- ⏳ **Step 8 — Tier 4** (whiteboard, voice capture, file hub, public roadmap, gamification, guest collab): NOT STARTED.
- ⏳ **Step 5b — Google Calendar push**: deferred (needs live OAuth).



Conventions (must match existing code):
- **Backend**: Go, chi, pgx/v5. Each feature = package under `backend/internal/<name>/` with `model.go → repository.go → service.go → handler.go`. Migrations numbered `000N_*.up.sql`/`.down.sql` in `internal/migrations/`, auto-run on startup. Wire in `cmd/api/main.go` (`NewRepository → NewService → NewHandler`) then register routes in `internal/server/server.go` inside the `AuthenticateAny` group. SSE via `sseBroker`. Notifications via `notifSvc`.
- **Frontend**: Next.js 19 app router, TanStack Query, zod + react-hook-form, shadcn/ui, lucide, sonner. One `lib/<feature>-hooks.ts` per feature (mirror `pomodoro-hooks.ts`). Pages under `app/(app)/`. Realtime via `sse-hook.ts`.
- Next migration number: **0007**.

Legend: ◻ = task. Effort: S<1d, M 1-3d, L 3-7d, XL >1wk.

---

# TIER 1 — Missing Pillars

## 1A. Notes / Docs / Wiki  (XL) — top pick

Goal: rich markdown docs, doc↔task links, `[[backlinks]]`, foldering, sharing reuse.

### DB — `0007_notes.up.sql`
```sql
CREATE TABLE notes (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  parent_id   UUID REFERENCES notes(id) ON DELETE CASCADE,   -- folder tree
  title       TEXT NOT NULL DEFAULT 'Untitled',
  body        TEXT NOT NULL DEFAULT '',                       -- markdown
  is_folder   BOOLEAN NOT NULL DEFAULT false,
  icon        TEXT,
  position    DOUBLE PRECISION NOT NULL DEFAULT 0,
  archived    BOOLEAN NOT NULL DEFAULT false,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notes_user ON notes(user_id, parent_id);
-- full-text (feeds Tier 2 global search)
ALTER TABLE notes ADD COLUMN search_tsv tsvector
  GENERATED ALWAYS AS (to_tsvector('english', title || ' ' || body)) STORED;
CREATE INDEX idx_notes_tsv ON notes USING GIN(search_tsv);

-- doc <-> task links (also backs task "linked docs" panel)
CREATE TABLE note_links (
  note_id  UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  task_id  UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  PRIMARY KEY (note_id, task_id)
);
-- note <-> note backlinks resolved from [[id]] on save
CREATE TABLE note_backlinks (
  src_id  UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  dst_id  UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
  PRIMARY KEY (src_id, dst_id)
);
```

### Backend — `internal/notes/`
- ◻ `model.go`: `Note{ID,UserID,ParentID,Title,Body,IsFolder,Icon,Position,Archived,CreatedAt,UpdatedAt}`, `CreateRequest`, `UpdateRequest` (all ptr fields for patch).
- ◻ `repository.go`: `List(userID)` (tree, ordered parent+position), `Get`, `Create`, `Update`, `Delete`, `Reorder`, `LinkTask/UnlinkTask/ListTaskLinks`, `ListByTask(taskID)`, `SetBacklinks(srcID, dstIDs[])`, `Backlinks(noteID)`.
- ◻ `service.go`: on Update, regex `\[\[([0-9a-f-]{36})\]\]` from body → `SetBacklinks`. Validate ownership. Autosave-friendly (idempotent PATCH).
- ◻ `handler.go`: `List, Get, Create, Update, Delete, Reorder, LinkTask, UnlinkTask, ListTaskLinks, Backlinks`.
- ◻ Routes (server.go): `GET/POST /notes`, `GET/PATCH/DELETE /notes/{id}`, `PUT /notes/reorder`, `GET /notes/{id}/backlinks`, `POST/DELETE /notes/{id}/tasks/{taskId}`, `GET /tasks/{id}/notes`.
- ◻ Wire main.go. SSE event `note_updated` for multi-tab sync.

### Frontend
- ◻ `lib/notes-hooks.ts`: `useNotes` (tree), `useNote(id)`, `useCreateNote`, `useUpdateNote` (debounced autosave — reuse existing autosave pattern from TaskForm), `useDeleteNote`, `useReorderNotes`, `useNoteBacklinks`, `useTaskNotes`.
- ◻ Nav: add "Docs" to `(app)/layout.tsx` sidebar.
- ◻ `app/(app)/docs/page.tsx` + `docs/[id]/page.tsx`: 2-pane — left tree (dnd-kit, collapsible folders), right editor.
- ◻ Editor: **BlockNote** (`@blocknote/react` + `@blocknote/mantine`) — Notion-style slash menu, drag blocks, out-of-box. Store BlockNote JSON in `notes.body` (jsonb-friendly text); also derive plaintext for `search_tsv`. Custom `[[` inline command → note-picker combobox inserting a link block to `[[id]]`, rendered as title chip resolving via `useNote`. Debounced autosave (existing pattern). SSR: dynamic import `ssr:false`.
- ◻ Backlinks footer panel; "Linked tasks" panel w/ add via task combobox.
- ◻ Task detail: add "Docs" tab listing `useTaskNotes`.
- ◻ Command palette: "New doc", jump-to-doc.
- ◻ Reuse `sharing` pkg pattern for public doc share (Tier 3 extends).

---

## 1B. Goals / OKRs  (L)

Goal: top layer above sprints/tasks. Goal → key results → linked tasks → auto progress %.

### DB — `0008_goals.up.sql`
```sql
CREATE TYPE goal_status AS ENUM ('on_track','at_risk','off_track','done','archived');
CREATE TABLE goals (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title       TEXT NOT NULL,
  description TEXT,
  status      goal_status NOT NULL DEFAULT 'on_track',
  target_date DATE,
  parent_id   UUID REFERENCES goals(id) ON DELETE SET NULL,  -- objective→sub-goal
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE key_results (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  goal_id     UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
  title       TEXT NOT NULL,
  metric_type TEXT NOT NULL DEFAULT 'percent',  -- percent|number|task_completion
  current_val DOUBLE PRECISION NOT NULL DEFAULT 0,
  target_val  DOUBLE PRECISION NOT NULL DEFAULT 100,
  position    DOUBLE PRECISION NOT NULL DEFAULT 0
);
-- link tasks to a goal (progress source when metric_type=task_completion)
ALTER TABLE tasks ADD COLUMN goal_id UUID REFERENCES goals(id) ON DELETE SET NULL;
```

### Backend — `internal/goals/`
- ◻ model/repo/service/handler. Repo: `List, Get, Create, Update, Delete` (goals); `AddKR, UpdateKR, DeleteKR, ReorderKR`; `ProgressFor(goalID)` — for `task_completion` KRs, compute `done/total` over `tasks WHERE goal_id=`.
- ◻ service: recompute KR `current_val` + roll up goal status when a linked task hits `done` (hook into `tasks.Service.UpdateTask` via callback or subscribe to SSE task events). Simplest: tasks.Service gets optional `goalProgress GoalRecalculator` dep, call on status change.
- ◻ Routes: `GET/POST /goals`, `GET/PATCH/DELETE /goals/{id}`, `POST /goals/{id}/key-results`, `PATCH/DELETE /goals/{id}/key-results/{krId}`, `GET /goals/{id}/progress`.
- ◻ Extend tasks model/handlers: accept `goal_id` in create/update; expose in Task JSON.

### Frontend
- ◻ `lib/goals-hooks.ts`. Sidebar "Goals".
- ◻ `app/(app)/goals/page.tsx`: cards w/ progress ring (status color), target date, KR list. Detail drawer: edit KRs, see linked tasks, sub-goals tree.
- ◻ TaskForm: add Goal combobox (reuse projects-combobox pattern).
- ◻ Dashboard widget (Tier 2 2D) shows goal progress.

---

## 1C. Habits + Streaks  (M)

Goal: recurring habit grid distinct from tasks; streak counts; heatmap.

### DB — `0009_habits.up.sql`
```sql
CREATE TABLE habits (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  cadence     TEXT NOT NULL DEFAULT 'daily',  -- daily|weekly|custom
  target_per_period INT NOT NULL DEFAULT 1,
  color       TEXT,
  archived    BOOLEAN NOT NULL DEFAULT false,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE habit_logs (
  habit_id  UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
  log_date  DATE NOT NULL,
  count     INT NOT NULL DEFAULT 1,
  PRIMARY KEY (habit_id, log_date)
);
```

### Backend — `internal/habits/`
- ◻ Repo: `List, Create, Update, Archive, Delete`; `Toggle(habitID, date)` (upsert/delete log); `Logs(habitID, from, to)`; `Streak(habitID)` — current+longest computed in SQL window or Go.
- ◻ Routes: `GET/POST /habits`, `PATCH/DELETE /habits/{id}`, `POST /habits/{id}/toggle` (body `{date}`), `GET /habits/{id}/logs?from=&to=`.
- ◻ Optional: scheduler creates `due_reminder` notif for uncompleted daily habits at user-set hour.

### Frontend
- ◻ `lib/habits-hooks.ts`. `app/(app)/habits/page.tsx`: row per habit w/ 7-day toggle dots + current streak flame; GitHub-style year heatmap on detail. Reuse existing daily-review for "today's habits".

---

## 1D. Calendar View + 2-way Google Sync  (L→XL)

Only ICS export exists today. Add in-app calendar grid + optional Google Calendar pull/push (OAuth already wired in `auth/oauth`).

### Phase 1 — Calendar view (L, no external dep)
- ◻ Frontend only: add `calendar` to `ViewToggle.tsx`. New `tasks/_components/CalendarView.tsx` — month grid (date-fns), tasks placed by `due_date`, drag task to date → `PATCH /tasks/{id}` due_date (reuse kanban dnd-kit). Day cells show count + overflow popover. Week/day sub-modes (weekly view already exists — unify).

### Phase 2 — Google 2-way sync (XL)
- ◻ DB `0010_calendar_sync.up.sql`: `calendar_connections(id,user_id,provider,access_token,refresh_token,expiry,calendar_id,sync_token,created_at)`; add `tasks.external_event_id TEXT`.
- ◻ `internal/calendarsync/`: OAuth scope `calendar.events` (extend existing Google OAuth consent). Repo stores tokens (encrypt at rest — reuse any existing crypto; else AES via env key). Service:
  - push: on task create/update with due_date → upsert Google event, store `external_event_id`.
  - pull: scheduler tick (extend `scheduler.go`) every 15min uses incremental `sync_token`; map events→tasks.
- ◻ Routes: `GET /calendar/connect` (oauth redirect), `GET /calendar/callback`, `DELETE /calendar/disconnect`, `GET /calendar/status`.
- ◻ Settings page: connect/disconnect card.
- Risk: token refresh, conflict resolution (last-write-wins by `updated_at`), rate limits. Gate behind feature flag (nil-handler pattern like groq).

---

# TIER 2 — Extend Existing

## 2A. Automation Rules Engine  (L)

Webhooks fire outward only. Add inbound if-this-then-that on task events.

### DB — `0011_automations.up.sql`
```sql
CREATE TABLE automations (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name      TEXT NOT NULL,
  enabled   BOOLEAN NOT NULL DEFAULT true,
  trigger   JSONB NOT NULL,  -- {event:'status_changed', to:'done'} | {event:'created'} | {event:'due_soon'}
  conditions JSONB NOT NULL DEFAULT '[]',  -- [{field:'priority',op:'eq',val:'high'}]
  actions   JSONB NOT NULL,  -- [{type:'set_tag',val:..}, {type:'notify_watchers'}, {type:'set_status',val:..}, {type:'move_project',val:..}, {type:'webhook',val:url}]
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Backend — `internal/automations/`
- ◻ model/repo/service/handler. Service exposes `Evaluate(ctx, userID, event TaskEvent)`.
- ◻ Engine: `tasks.Service` emits domain events (already emits SSE) — add an `automationEngine` dep; call `Evaluate` after Create/Update. Match trigger+conditions → run actions (reuse tags/projects/notif/webhooks services injected into automations.Service).
- ◻ Guard against loops (action that re-triggers): max depth 1, skip events caused by automations.
- ◻ Routes: `GET/POST /automations`, `PATCH/DELETE /automations/{id}`, `POST /automations/{id}/test`.

### Frontend
- ◻ `lib/automations-hooks.ts`. `app/(app)/settings` new tab: rule builder (trigger select → condition rows → action rows). Show run log (optional `automation_runs` table).

---

## 2B. Reminders + Snooze  (M)

Notifications exist but no custom per-task reminder time / snooze.

### DB — `0012_reminders.up.sql`
```sql
CREATE TABLE reminders (
  id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  task_id   UUID REFERENCES tasks(id) ON DELETE CASCADE,
  remind_at TIMESTAMPTZ NOT NULL,
  sent      BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_reminders_due ON reminders(remind_at) WHERE NOT sent;
ALTER TABLE notifications ADD COLUMN snoozed_until TIMESTAMPTZ;
```

### Backend — `internal/reminders/`
- ◻ Repo: `Create, List(taskID), Delete, DuePending(now)`. Scheduler tick → `DuePending` → notif + SSE + mark sent.
- ◻ Notifications: add `Snooze(id, until)` — hide until time; scheduler resurfaces.
- ◻ Routes: `POST /tasks/{id}/reminders`, `GET /tasks/{id}/reminders`, `DELETE /reminders/{id}`, `POST /notifications/{id}/snooze`.

### Frontend
- ◻ `lib/reminders-hooks.ts`. TaskForm: "Add reminder" (datetime). NotificationsBell: snooze menu (1h/tomorrow/custom).

---

## 2C. Global Full-Text Search  (M)

Today only `/tasks?search=`. Add unified search over tasks + comments + notes + attachments(name).

### DB — `0013_search.up.sql`
- ◻ Add `search_tsv` generated column + GIN index to `tasks` (title+description) and `comments` (body). (`notes` already has it from 1A.)

### Backend — `internal/search/`
- ◻ `GET /search?q=` → `Service.Search` runs UNION across tasks/comments/notes with `ts_rank`, returns typed results `{type, id, title, snippet, task_id?}`. ts_headline for snippet. Limit 30, grouped.

### Frontend
- ◻ Upgrade `CommandPalette.tsx`: debounced query → `/search`, grouped sections (Tasks/Docs/Comments), keyboard nav, jump on enter. Keep existing static commands above results.

---

## 2D. Personal Dashboard  (M)

Analytics are admin-only. Give users a home dashboard.

### Backend
- ◻ `internal/dashboard/` or extend tasks: `GET /dashboard` → aggregate `{due_today[], overdue[], completed_this_week, time_this_week_minutes, streaks[], goal_progress[], upcoming[]}`. Single endpoint, parallel queries.

### Frontend
- ◻ Make `(app)/page.tsx` a widget dashboard (bento grid): Due Today, Overdue, This Week chart (reuse admin chart lib), Active Goals, Habit streaks, AI weekly-digest card (existing `/ai/weekly-digest`), Pomodoro today. Each widget links into its module.

---

## 2E. Multi-user Workspaces / Teams — ❌ DROPPED

Decision: personal single-user product. No teams, no assignees. Existing `assignee_id` column stays dormant (no UI). Guest collaboration handled link-scoped in Tier 4, not via workspaces.

---

# TIER 3 — Integrations + Polish

## 3A. Email-to-Task / Inbox  (M)
- ◻ Inbound via **Resend Inbound** webhook → `POST /webhooks/email` (public, verify Resend signature). Parse `to` = `u+<userInboxToken>@inbox.<domain>`. DB `0014`: `users.inbox_token TEXT UNIQUE`. Create task from subject(title)+body(description); attachments→S3 (attachments pkg exists). Settings shows user's inbox address + copy button. Configure Resend inbound domain MX + webhook URL.

## 3B. Native Slack/Discord Notify  (S→M)
- ◻ Extend `webhooks` pkg: add `target_type` (generic|slack|discord). Format payload per target (Slack blocks / Discord embeds). Settings: pick type + paste incoming-webhook URL. Reuse automation action `webhook` to target these.

## 3C. PWA + Offline + Web Push  (M) — promoted: web-push is a chosen delivery channel
- ◻ `manifest.json` + icons, service worker (`next-pwa` or hand-rolled). Cache shell + GET /tasks. Offline mutation queue via TanStack persist + replay on reconnect (SSE already detects reconnect). Install prompt.
- ◻ **Web push** (powers reminders/notif channel): generate VAPID keys (env `VAPID_PUBLIC/PRIVATE`). DB `0015`: `push_subscriptions(id,user_id,endpoint,p256dh,auth,created_at)`. Routes `POST/DELETE /push/subscribe`. Backend `internal/webpush/` sends via VAPID (`SherClockHolmes/webpush-go`). `notifications.Service` + `scheduler` + `reminders` fan out to: in-app SSE (done) + email (Resend) + web-push + Slack/Discord webhook — unified `Deliver(userID, notif)` dispatcher honoring user prefs.

## 3D. More AI (groq)  (M, incremental)
Extend existing `internal/groq/handler.go` (nil-guard pattern already there):
- ◻ `POST /ai/plan-day` — input today's open tasks + working hours → time-boxed schedule (slots) → optionally create reminders/calendar blocks.
- ◻ `POST /ai/search` — natural-language → structured filter for task list ("high prio overdue in projX").
- ◻ `POST /ai/meeting-notes` — paste notes → extract action items → bulk-create tasks (reuse `/tasks/bulk`... add bulk-create).
- ◻ `POST /ai/prioritize` — reorder backlog by urgency/impact, return suggested `sort_order` + rationale.
- ◻ Frontend: add to `ai-hooks.ts`; surface in dashboard / daily-review / command palette.

## 3E. Mobile-responsive Audit  (S)
- ◻ Kanban/Gantt/Calendar horizontal-scroll + touch-drag pass. Sidebar → drawer on mobile. Test 360px. Checklist in `tasks/lessons.md`.

---

# TIER 4 — Nice-to-have

- ◻ **Whiteboard / Mindmap** (L): `tldraw` or excalidraw embed per board; persist JSON in `boards` table; link nodes→tasks. Heavy; standalone module.
- ◻ **Voice quick-capture** (S): Web Speech API in `QuickCaptureDialog` → transcript → existing `/ai/parse-task`.
- ◻ **File hub** (M): aggregate all attachments across tasks into one browsable view (attachments pkg already stores rows) — `GET /attachments?all=1`, grid UI, filter by type/project.
- ◻ **Public roadmap** (M): mark projects/tasks `public_roadmap`; reuse `sharing` for a read-only kanban at `/roadmap/{slug}`; upvotes table.
- ◻ **Gamification** (M): XP from completed tasks + habit streaks; `user_stats` table; levels/badges; dashboard widget. Derive from existing activity log.
- ◻ **Guest collaborators on share links** (M): extend `sharing` — token grants comment/edit scope, not just view; ties into workspace `guest` role.

---

# Sequencing (recommended)

0. **Shared infra first**: `users.notif_prefs` + unified `notifications.Deliver()` dispatcher, then **3C web-push/VAPID + PWA** (delivery channel everything else uses).
1. **2C Search** + **1A Notes** together (notes adds tsvector; search consumes it). Highest leverage.
2. **1C Habits** + **2D Dashboard** (dashboard surfaces habits/goals).
3. **1B Goals** (dashboard already has a slot).
4. **2B Reminders** + **3D AI plan-day** (reminders back the AI schedule, deliver via channels).
5. **1D Calendar** phase 1 (view), then phase 2 (Google push).
6. **2A Automations** (after notif/tags/webhooks stable) + **3B Slack/Discord** action.
7. **3A Resend inbound** + **3E mobile audit**.
8. **Tier 4**: whiteboard, voice, file hub, public roadmap, gamification, guest collab.

Build order per feature: migration → backend pkg → wire main.go → routes → curl-test → hooks → UI → SSE → verify (run app, prove behavior) before marking done.

---

# Locked decisions
- **Scope**: build everything (real product), Tiers 1→4. Workspaces/teams cut.
- **No teams/assignees** — personal product. `assignee_id` stays dormant.
- **Notes editor**: BlockNote (Notion-style blocks).
- **Calendar (1D)**: view + **one-way push** tasks→Google (existing Google OAuth in .env). No pull-back v1.
- **Notif delivery**: all 4 channels — in-app SSE + email (Resend) + web-push (VAPID) + Slack/Discord webhook. User prefs toggle per channel (`users.notif_prefs jsonb`).
- **Inbound email (3A)**: Resend Inbound.
- **Tier 4**: keep all (public roadmap + guest collab included; guest collab = share-link-scoped, not workspace).

# Remaining minor calls (non-blocking, will default)
- Google token storage: encrypt-at-rest via new `ENCRYPTION_KEY` env (AES-GCM) — default yes.
- Migration numbering may shift as features land in real order; renumber sequentially at build time.
