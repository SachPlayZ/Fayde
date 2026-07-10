package friends

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for the friends endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new friends Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// SendRequest handles POST /friends/requests.
func (h *Handler) SendRequest(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	var input SendRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.Error(w, 400, "invalid request body")
		return
	}
	if input.AddresseeEmail == "" {
		httputil.Error(w, 400, "addressee_email is required")
		return
	}

	f, err := h.svc.SendRequest(r.Context(), userID, input.AddresseeEmail)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "user not found")
		case errors.Is(err, ErrCannotSelfAdd):
			httputil.Error(w, 400, "cannot send request to yourself")
		case errors.Is(err, ErrAlreadyExists):
			httputil.Error(w, 409, "friend request already exists")
		default:
			httputil.Error(w, 500, "failed to send friend request")
		}
		return
	}

	httputil.JSON(w, 201, f)
}

// AcceptRequest handles POST /friends/requests/{id}/accept.
func (h *Handler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.AcceptRequest(r.Context(), userID, id); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "request not found")
		case errors.Is(err, ErrForbidden):
			httputil.Error(w, 403, "forbidden")
		default:
			httputil.Error(w, 500, "failed to accept request")
		}
		return
	}

	httputil.JSON(w, 200, map[string]string{"status": "accepted"})
}

// DeclineRequest handles POST /friends/requests/{id}/decline.
func (h *Handler) DeclineRequest(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.DeclineRequest(r.Context(), userID, id); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "request not found")
		case errors.Is(err, ErrForbidden):
			httputil.Error(w, 403, "forbidden")
		default:
			httputil.Error(w, 500, "failed to decline request")
		}
		return
	}

	httputil.JSON(w, 200, map[string]string{"status": "declined"})
}

// Remove handles DELETE /friends/{id}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.Remove(r.Context(), userID, id); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			httputil.Error(w, 404, "not found")
		case errors.Is(err, ErrForbidden):
			httputil.Error(w, 403, "forbidden")
		default:
			httputil.Error(w, 500, "failed to remove friend")
		}
		return
	}

	httputil.JSON(w, 204, nil)
}

// ListFriends handles GET /friends.
func (h *Handler) ListFriends(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	list, err := h.svc.ListFriends(r.Context(), userID)
	if err != nil {
		httputil.Error(w, 500, "failed to list friends")
		return
	}
	if list == nil {
		list = []*FriendRequest{}
	}
	httputil.JSON(w, 200, list)
}

// ListRequests handles GET /friends/requests.
func (h *Handler) ListRequests(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	list, err := h.svc.ListRequests(r.Context(), userID)
	if err != nil {
		httputil.Error(w, 500, "failed to list requests")
		return
	}
	if list == nil {
		list = []*FriendRequest{}
	}
	httputil.JSON(w, 200, list)
}

// Search handles GET /friends/search?q=email.
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, 401, "unauthorized")
		return
	}

	q := r.URL.Query().Get("q")
	results, err := h.svc.SearchUsers(r.Context(), q)
	if err != nil {
		httputil.Error(w, 500, "search failed")
		return
	}
	if results == nil {
		results = []*FriendUser{}
	}
	httputil.JSON(w, 200, results)
}
