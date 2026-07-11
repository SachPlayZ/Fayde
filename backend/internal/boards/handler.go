package boards

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for the boards endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new boards Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Create handles POST /boards.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	var input CreateBoardInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.Error(w, 400, "invalid request body")
		return
	}
	if input.Name == "" {
		httputil.Error(w, 400, "name is required")
		return
	}

	b, err := h.svc.CreateBoard(r.Context(), userID, input)
	if err != nil {
		httputil.Error(w, 500, "failed to create board")
		return
	}
	httputil.JSON(w, 201, b)
}

// List handles GET /boards.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boards, err := h.svc.ListBoards(r.Context(), userID)
	if err != nil {
		httputil.Error(w, 500, "failed to list boards")
		return
	}
	if boards == nil {
		boards = []*Board{}
	}
	httputil.JSON(w, 200, boards)
}

// Get handles GET /boards/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	detail, err := h.svc.GetBoard(r.Context(), userID, boardID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "board not found")
		case errors.Is(err, ErrNotMember):
			httputil.Error(w, 403, "not a member of this board")
		default:
			httputil.Error(w, 500, "failed to get board")
		}
		return
	}
	httputil.JSON(w, 200, detail)
}

// Update handles PATCH /boards/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	var input UpdateBoardInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.Error(w, 400, "invalid request body")
		return
	}

	if err := h.svc.UpdateBoard(r.Context(), userID, boardID, input); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "board not found")
		case errors.Is(err, ErrForbidden):
			httputil.Error(w, 403, "only the owner can update the board")
		default:
			httputil.Error(w, 500, "failed to update board")
		}
		return
	}
	httputil.JSON(w, 200, map[string]string{"status": "updated"})
}

// Delete handles DELETE /boards/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	if err := h.svc.DeleteBoard(r.Context(), userID, boardID); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "board not found")
		case errors.Is(err, ErrForbidden):
			httputil.Error(w, 403, "only the owner can delete the board")
		default:
			httputil.Error(w, 500, "failed to delete board")
		}
		return
	}
	httputil.JSON(w, 204, nil)
}

// AddTask handles POST /boards/{id}/tasks.
func (h *Handler) AddTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	var input AddTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.Error(w, 400, "invalid request body")
		return
	}
	if input.Title == "" {
		httputil.Error(w, 400, "title is required")
		return
	}

	task, err := h.svc.AddTask(r.Context(), userID, boardID, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotMember):
			httputil.Error(w, 403, "not a member of this board")
		default:
			httputil.Error(w, 500, "failed to add task")
		}
		return
	}
	httputil.JSON(w, 201, task)
}

// DeleteTask handles DELETE /boards/{id}/tasks/{taskId}.
func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	taskID := chi.URLParam(r, "taskId")

	if err := h.svc.DeleteTask(r.Context(), userID, boardID, taskID); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "task not found")
		case errors.Is(err, ErrForbidden):
			httputil.Error(w, 403, "only the owner or creator can delete this task")
		default:
			httputil.Error(w, 500, "failed to delete task")
		}
		return
	}
	httputil.JSON(w, 204, nil)
}

// Complete handles POST /boards/{id}/tasks/{taskId}/complete.
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	taskID := chi.URLParam(r, "taskId")

	c, err := h.svc.Complete(r.Context(), userID, boardID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotMember):
			httputil.Error(w, 403, "not a member of this board")
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "task not found")
		default:
			httputil.Error(w, 500, "failed to complete task")
		}
		return
	}
	httputil.JSON(w, 200, c)
}

// Uncomplete handles DELETE /boards/{id}/tasks/{taskId}/complete.
func (h *Handler) Uncomplete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	taskID := chi.URLParam(r, "taskId")

	if err := h.svc.Uncomplete(r.Context(), userID, boardID, taskID); err != nil {
		switch {
		case errors.Is(err, ErrNotMember):
			httputil.Error(w, 403, "not a member of this board")
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "task not found")
		default:
			httputil.Error(w, 500, "failed to uncomplete task")
		}
		return
	}
	httputil.JSON(w, 204, nil)
}

// InviteFriend handles POST /boards/{id}/invite.
func (h *Handler) InviteFriend(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	var input InviteFriendInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.Error(w, 400, "invalid request body")
		return
	}
	if input.FriendUserID == "" {
		httputil.Error(w, 400, "friend_user_id is required")
		return
	}

	if err := h.svc.InviteFriend(r.Context(), userID, boardID, input.FriendUserID); err != nil {
		switch {
		case errors.Is(err, ErrNotMember):
			httputil.Error(w, 403, "not a member of this board")
		case errors.Is(err, ErrNotFriends):
			httputil.Error(w, 403, "target user is not your friend")
		case errors.Is(err, ErrAlreadyMember):
			httputil.Error(w, 409, "user is already a member")
		default:
			httputil.Error(w, 500, "failed to invite friend")
		}
		return
	}
	httputil.JSON(w, 200, map[string]string{"status": "invited"})
}

// CreateShareToken handles POST /boards/{id}/share.
func (h *Handler) CreateShareToken(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	inv, err := h.svc.CreateShareToken(r.Context(), userID, boardID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotMember):
			httputil.Error(w, 403, "not a member of this board")
		default:
			httputil.Error(w, 500, "failed to create share token")
		}
		return
	}
	httputil.JSON(w, 200, map[string]string{
		"token": inv.Token,
		"url":   "/boards/join/" + inv.Token,
	})
}

// RevokeShareToken handles DELETE /boards/{id}/share.
func (h *Handler) RevokeShareToken(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	if err := h.svc.RevokeShareToken(r.Context(), userID, boardID); err != nil {
		switch {
		case errors.Is(err, ErrNotMember):
			httputil.Error(w, 403, "not a member of this board")
		default:
			httputil.Error(w, 500, "failed to revoke share token")
		}
		return
	}
	httputil.JSON(w, 204, nil)
}

// JoinPreview handles GET /boards/join/{token} (public, no auth).
func (h *Handler) JoinPreview(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	preview, err := h.svc.JoinPreview(r.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			httputil.Error(w, 404, "invalid or expired invite token")
		default:
			httputil.Error(w, 500, "failed to fetch board preview")
		}
		return
	}
	httputil.JSON(w, 200, preview)
}

// Join handles POST /boards/join/{token}.
func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	token := chi.URLParam(r, "token")
	board, err := h.svc.JoinViaToken(r.Context(), userID, token)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken):
			httputil.Error(w, 404, "invalid or expired invite token")
		default:
			httputil.Error(w, 500, "failed to join board")
		}
		return
	}
	httputil.JSON(w, 200, map[string]string{"board_id": board.ID})
}

// ShareTask handles POST /boards/{id}/shared-tasks.
func (h *Handler) ShareTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	var input ShareTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.Error(w, 400, "invalid request body")
		return
	}
	if input.TaskID == "" {
		httputil.Error(w, 400, "task_id is required")
		return
	}

	if err := h.svc.ShareTask(r.Context(), userID, boardID, input.TaskID); err != nil {
		switch {
		case errors.Is(err, ErrNotMember):
			httputil.Error(w, 403, "not a member of this board")
		case errors.Is(err, ErrForbidden):
			httputil.Error(w, 403, "you do not own this task")
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "task not found")
		default:
			httputil.Error(w, 500, "failed to share task")
		}
		return
	}
	httputil.JSON(w, 200, map[string]string{"status": "shared"})
}

// UnshareTask handles DELETE /boards/{id}/shared-tasks/{taskId}.
func (h *Handler) UnshareTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	boardID := chi.URLParam(r, "id")
	taskID := chi.URLParam(r, "taskId")

	if err := h.svc.UnshareTask(r.Context(), userID, boardID, taskID); err != nil {
		switch {
		case errors.Is(err, ErrNotMember):
			httputil.Error(w, 403, "not a member of this board")
		case errors.Is(err, ErrForbidden):
			httputil.Error(w, 403, "you do not own this task")
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "task not found")
		default:
			httputil.Error(w, 500, "failed to unshare task")
		}
		return
	}
	httputil.JSON(w, 204, nil)
}

// ListBoardsForTask handles GET /tasks/{id}/boards ({id} here is the task ID).
func (h *Handler) ListBoardsForTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	taskID := chi.URLParam(r, "id")
	entries, err := h.svc.ListBoardsForTask(r.Context(), userID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			httputil.Error(w, 403, "you do not own this task")
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "task not found")
		default:
			httputil.Error(w, 500, "failed to list boards for task")
		}
		return
	}
	httputil.JSON(w, 200, entries)
}
