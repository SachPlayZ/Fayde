package notes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list notes")
		return
	}
	httputil.JSON(w, http.StatusOK, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	n, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"), userID)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "note not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get note")
		return
	}
	httputil.JSON(w, http.StatusOK, n)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	n, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create note")
		return
	}
	httputil.JSON(w, http.StatusCreated, n)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	n, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), userID, req)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "note not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to update note")
		return
	}
	httputil.JSON(w, http.StatusOK, n)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	err := h.svc.Delete(r.Context(), chi.URLParam(r, "id"), userID)
	if errors.Is(err, ErrNotFound) {
		httputil.Error(w, http.StatusNotFound, "note not found")
		return
	}
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to delete note")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var items []ReorderItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.svc.Reorder(r.Context(), userID, items); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to reorder notes")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Backlinks(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	refs, err := h.svc.Backlinks(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to get backlinks")
		return
	}
	httputil.JSON(w, http.StatusOK, refs)
}

func (h *Handler) ListTaskLinks(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	ids, err := h.svc.ListTaskLinks(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list task links")
		return
	}
	httputil.JSON(w, http.StatusOK, ids)
}

func (h *Handler) LinkTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	noteID := chi.URLParam(r, "id")
	taskID := chi.URLParam(r, "taskId")
	if err := h.svc.LinkTask(r.Context(), noteID, taskID, userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to link task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UnlinkTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	noteID := chi.URLParam(r, "id")
	taskID := chi.URLParam(r, "taskId")
	if err := h.svc.UnlinkTask(r.Context(), noteID, taskID, userID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to unlink task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListByTask returns notes linked to a task (mounted under /tasks/{id}/notes).
func (h *Handler) ListByTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	refs, err := h.svc.ListByTask(r.Context(), chi.URLParam(r, "id"), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list notes")
		return
	}
	httputil.JSON(w, http.StatusOK, refs)
}

// UploadImage handles POST /notes/images
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		httputil.Error(w, http.StatusBadRequest, "file too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	filename, err := h.svc.UploadImage(r.Context(), header.Filename, contentType, file, header.Size)
	if err != nil {
		if strings.Contains(err.Error(), "not configured") {
			httputil.Error(w, http.StatusNotImplemented, "S3 storage not configured")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to upload image")
		return
	}

	imageURL := fmt.Sprintf("/notes/images/%s", filename)
	httputil.JSON(w, http.StatusCreated, map[string]string{"url": imageURL})
}

// GetImage handles GET /notes/images/{filename}
func (h *Handler) GetImage(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		httputil.Error(w, http.StatusBadRequest, "missing filename")
		return
	}

	s3Key := "notes/images/" + filename
	body, contentType, err := h.svc.DownloadImage(r.Context(), s3Key)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "image not found")
		return
	}
	defer body.Close()

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	_, _ = io.Copy(w, body)
}
