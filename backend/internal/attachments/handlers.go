package attachments

import (
	"net/http"
	"strings"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/SachPlayZ/rivz-asn/backend/internal/tasks"
	"github.com/go-chi/chi/v5"
)

const maxUploadSize = 10 << 20 // 10 MB

// Handler handles HTTP requests for attachment endpoints.
type Handler struct {
	svc      *Service
	tasksSvc *tasks.Service
	s3Bucket string
}

// NewHandler creates a new attachments Handler.
func NewHandler(svc *Service, tasksSvc *tasks.Service, s3Bucket string) *Handler {
	return &Handler{svc: svc, tasksSvc: tasksSvc, s3Bucket: s3Bucket}
}

// checkConfigured returns false and writes a 501 if S3 is not configured.
func (h *Handler) checkConfigured(w http.ResponseWriter) bool {
	if h.s3Bucket == "" {
		httputil.Error(w, http.StatusNotImplemented, "attachments not configured")
		return false
	}
	return true
}

// Upload handles POST /tasks/{id}/attachments.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if !h.checkConfigured(w) {
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	taskID := chi.URLParam(r, "id")

	// Verify task ownership.
	if _, err := h.tasksSvc.GetTask(r.Context(), taskID, userID); err != nil {
		if strings.Contains(err.Error(), "no rows") {
			httputil.Error(w, http.StatusNotFound, "task not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to verify task ownership")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
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

	att, err := h.svc.Upload(r.Context(), taskID, userID, header.Filename, contentType, file, header.Size)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to upload attachment")
		return
	}

	httputil.JSON(w, http.StatusCreated, att)
}

// List handles GET /tasks/{id}/attachments.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if !h.checkConfigured(w) {
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	taskID := chi.URLParam(r, "id")

	// Verify task ownership.
	if _, err := h.tasksSvc.GetTask(r.Context(), taskID, userID); err != nil {
		if strings.Contains(err.Error(), "no rows") {
			httputil.Error(w, http.StatusNotFound, "task not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to verify task ownership")
		return
	}

	list, err := h.svc.List(r.Context(), taskID, userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list attachments")
		return
	}

	httputil.JSON(w, http.StatusOK, list)
}

// Delete handles DELETE /tasks/{id}/attachments/{attId}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if !h.checkConfigured(w) {
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	taskID := chi.URLParam(r, "id")
	attID := chi.URLParam(r, "attId")

	// Verify task ownership.
	if _, err := h.tasksSvc.GetTask(r.Context(), taskID, userID); err != nil {
		if strings.Contains(err.Error(), "no rows") {
			httputil.Error(w, http.StatusNotFound, "task not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to verify task ownership")
		return
	}

	if err := h.svc.Delete(r.Context(), attID, taskID, userID); err != nil {
		if strings.Contains(err.Error(), "no rows") {
			httputil.Error(w, http.StatusNotFound, "attachment not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete attachment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
