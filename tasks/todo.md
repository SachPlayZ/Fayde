# Friends + Shared Boards + personal-tasks accordion + Focus cleanup

## Decisions
- Focus Mode (`/focus` page + `focusmode` backend pkg) is redundant with the
  always-on Pomodoro floating widget. REMOVE it (not fix).
- Friendship: `friendships(requester_id, addressee_id, status)`, status =
  pending|accepted|declined. No blocking, no multi-workspace — MVP.
- Board = owner + N members, each board has a fixed list of "board_tasks"
  (e.g. "Go to the Gym"), each member checks off their own completion per
  calendar day (UTC date, no timezone handling — MVP). Everyone on the board
  sees everyone's completions live via SSE + gets a lightweight notification.
- Personal-tasks accordion sits at the bottom of the board detail screen,
  collapsed by default, shows the signed-in user's own tasks due today
  (reuses existing `useTasks` hook — no new backend endpoint).

## Backend (Go)

### Migrations
- `0021_friends_and_boards.up/down.sql`
  - `friendships(id, requester_id, addressee_id, status, created_at, responded_at)`
    unique on (requester_id, addressee_id), check requester_id <> addressee_id
  - `boards(id, owner_id, name, description, created_at)`
  - `board_members(board_id, user_id, role, joined_at)` PK(board_id, user_id)
  - `board_tasks(id, board_id, title, sort_order, created_by, created_at)`
  - `board_task_completions(id, board_task_id, user_id, completion_date DATE, completed_at)`
    unique(board_task_id, user_id, completion_date)
  - `board_invites(id, board_id, token UNIQUE, created_by, expires_at NULL, created_at)`
- `0022_remove_focus_mode.up/down.sql` — drop `focus_sessions` (up), recreate (down)

### New packages (repo/service/handler each, follow existing 4-file pattern)
- `internal/friends/` — send/accept/decline requests, list friends, search users by email
  - repo queries `users` table directly (auth owns it, no cross-import of auth needed
    beyond a plain SQL SELECT via pgxpool — same as sharing does for tasks)
- `internal/boards/` — CRUD boards, board_tasks, join-by-token, complete/uncomplete,
  get board detail (members + tasks + today's completion matrix)
  - `SetNotificationsService`, `SetSSEBroker` cross-injection (mirror tasks.Service pattern)
  - on complete: SSE `Publish` to other members' userIDs with event type
    `board_task_completed`, payload {board_id, board_task_id, user_id, display_name, task_title, date}
  - also `notifications.Service.Create(...)` per other member, type `board_task_completed`

### Routes (server.go, inside AuthenticateAny group unless noted)
```
POST   /friends/requests            body {addressee_email}
POST   /friends/requests/{id}/accept
POST   /friends/requests/{id}/decline
DELETE /friends/{id}                (remove friend or cancel/decline own request)
GET    /friends
GET    /friends/requests            (incoming + outgoing pending)
GET    /friends/search?q=email

POST   /boards                      body {name, description}
GET    /boards                      (mine)
GET    /boards/{id}                 (detail: members, tasks, today completions)
PATCH  /boards/{id}                 (owner only)
DELETE /boards/{id}                 (owner only)
POST   /boards/{id}/tasks           body {title}  (any member)
DELETE /boards/{id}/tasks/{taskId}  (owner or creator)
POST   /boards/{id}/tasks/{taskId}/complete     (toggle on for today, for me)
DELETE /boards/{id}/tasks/{taskId}/complete     (toggle off for today, for me)
POST   /boards/{id}/invite          body {friend_user_id}  (direct add, must be friends)
POST   /boards/{id}/share           (create/rotate invite token)
DELETE /boards/{id}/share           (revoke)
GET    /boards/join/{token}         PUBLIC (no auth) — preview {board_name, member_count}
POST   /boards/join/{token}         (authed) — join board via token
```

### Wiring
- `main.go`: construct friendsRepo/Svc/Handler, boardsRepo/Svc/Handler; wire
  `boardsSvc.SetNotificationsService`, `boardsSvc.SetSSEBroker`; DELETE all
  focusmode construction lines; add friendsHandler/boardsHandler to `server.New(...)`.
- `server.go`: add routes above (public join-preview route outside auth group,
  same style as `/share/{token}`); add both handler params to `New(...)` signature;
  DELETE all focusmode routes.
- Delete `backend/internal/focusmode/` entirely.

### Tests (testcontainers, `go test -race`)
- `internal/friends/*_test.go` — send/accept/decline/list/search, duplicate-request rejection
- `internal/boards/*_test.go` — create board, add task, complete/uncomplete idempotency,
  join via token, non-member cannot see/complete, owner-only delete enforced

## Frontend (rivz/)

- `npx shadcn@latest add accordion` (none exists yet)
- DELETE `app/(app)/focus/`, `lib/focus-hooks.ts`, the `/focus` nav entry +
  unused `Brain` icon import in `AppSidebar.tsx`
- `lib/friends-hooks.ts` — useFriends, useFriendRequests, useSendFriendRequest,
  useAcceptFriendRequest, useDeclineFriendRequest, useRemoveFriend, useSearchUsers
- `lib/boards-hooks.ts` — useBoards, useBoard(id), useCreateBoard, useUpdateBoard,
  useDeleteBoard, useAddBoardTask, useDeleteBoardTask, useCompleteBoardTask,
  useUncompleteBoardTask, useInviteFriendToBoard, useBoardShareToken,
  useCreateBoardShareToken, useJoinBoardPreview(token), useJoinBoard
- `app/(app)/friends/page.tsx` — friends list, incoming/outgoing requests, search+add
- `app/(app)/boards/page.tsx` — my boards list + create dialog
- `app/(app)/boards/[id]/page.tsx` — matrix view (tasks × members with checkmarks,
  click own cell to toggle), add-task input, invite friend / copy-share-link UI
  (reuse the Input+Copy+Revoke pattern from TaskDetailClient sharing panel),
  live updates via `lastEvent.type === "board_task_completed"` from `useSSE()`,
  + bottom collapsed `<Accordion type="single" collapsible>` "My tasks today"
  pulling from existing `useTasks` filtered to today/todo+in_progress
- `app/join/[token]/page.tsx` — public preview + "sign in to join" / join action
  (mirrors `app/share/[token]/`)
- `AppSidebar.tsx` — add Friends + Boards nav entries, remove Focus entry

## Verification
- `cd backend && go build ./... && go vet ./... && go test -race ./internal/friends/... ./internal/boards/...`
- `cd rivz && pnpm lint && pnpm build`
- Manual browser pass: create 2 test users, friend each other, create board,
  add the 5 example tasks, complete one as user A, confirm user B sees it
  live (SSE) + gets notification, confirm accordion collapsed by default and
  shows personal tasks, confirm /focus route is gone and Pomodoro still works.

## Unresolved questions (none blocking — proceeding with defaults above)
- none — running autonomously per /goal directive

## Review (done)

Built via 2 parallel subagents (backend, frontend) against a shared contract, then
reconciled by hand and verified end-to-end with 2 real users through the actual API
+ browser (chrome-devtools), not just build/test green.

Bugs found and fixed during verification (agents' reports alone were not trustworthy):
- **Crash**: completing a board task → notification → email delivery panicked the
  whole backend process. Root cause: `main.go` passed a typed-nil `*email.Client`
  into the `EmailSender` interface param of `notifSvc.SetDeliverers` — a nil pointer
  boxed in an interface is non-nil, so `deliver()`'s `s.email != nil` check passed
  and it called a method on a nil receiver. Pre-existing bug, exposed by boards being
  the first feature to trigger a notification in this local run. Fixed by only
  assigning the interface var when the pointer is actually non-nil.
- **Contract mismatch (friends)**: backend resolves "other party" into a nested
  `user: {...}` object for both `/friends` and `/friends/requests`; frontend agent
  had written flat/prefixed fields (`user_id`, `requester_email`, etc.) independently.
  Adapted frontend types + 3 call sites to the backend's (cleaner, already-tested) shape.
- **Missing fields**: `Board.member_count` and `BoardDetail.share_token` were read by
  the frontend but never populated by the backend. Added both (COUNT subquery +
  invite lookup in `GetBoard`).
- **Wrong join response shape**: `POST /boards/join/{token}` returned the full board
  object; frontend read `data.board_id` for the redirect → would've redirected to
  `/boards/undefined`. Fixed handler to return `{"board_id": ...}`.
- **Dead code / stale cache key**: removed the unused `useBoardShareToken` GET hook,
  but that left `useCreateBoardShareToken`/`useRevokeBoardShareToken` invalidating an
  orphaned `["board-share", id]` query key instead of `["board", id]` (what the page
  actually reads `share_token` from) — share link create/revoke silently didn't
  update the UI until a manual reload. Fixed both mutations' `onSuccess`.

Verified live (2 real users, real Postgres, real browser):
- friend request → accept → both friends lists correct
- board create, 5 example tasks added, invite-by-friend
- completion toggle persists, matrix shows correct per-user state
- SSE push + notification delivered to the other member in real time
- share-link create/revoke updates instantly; link persists across reload
- accordion collapsed by default, expands to real personal tasks, updates label/count
- `/focus` 404s cleanly, nav has no Focus entry, Pomodoro widget still present/functional
- `go build/vet/test -race` clean, `pnpm lint/build` clean

Not covered (out of scope / would need more time): automated tests for the frontend,
join-via-link flow only checked via API not full second-browser UI session, mobile
viewport not checked.
