package showcase

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

const maxImageUploadSize = 5 << 20 // 5 MB

// Handler handles HTTP requests for showcase entries, images, tokens, and
// the public page/embed routes.
type Handler struct {
	svc      *Service
	s3Bucket string
}

// NewHandler creates a new showcase Handler. s3Bucket is empty when S3 isn't
// configured — image endpoints then respond 501 rather than panicking on a
// nil storage client.
func NewHandler(svc *Service, s3Bucket string) *Handler {
	return &Handler{svc: svc, s3Bucket: s3Bucket}
}

func (h *Handler) checkConfigured(w http.ResponseWriter) bool {
	if h.s3Bucket == "" {
		httputil.Error(w, http.StatusNotImplemented, "showcase image uploads not configured")
		return false
	}
	return true
}

func writeValidationErr(w http.ResponseWriter, err error) {
	fields := map[string]string{}
	if verrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range verrs {
			fields[strings.ToLower(e.Field())] = e.Tag()
		}
	}
	httputil.ValidationError(w, fields)
}

// --- Authenticated CRUD ---

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	entries, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list showcase entries")
		return
	}
	httputil.JSON(w, http.StatusOK, entries)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := validate.Struct(req); err != nil {
		writeValidationErr(w, err)
		return
	}
	e, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to create showcase entry")
		return
	}
	httputil.JSON(w, http.StatusCreated, e)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	e, err := h.svc.Get(r.Context(), id, userID)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "showcase entry not found")
		return
	}
	httputil.JSON(w, http.StatusOK, e)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	e, err := h.svc.Update(r.Context(), id, userID, req)
	if err != nil {
		if err == ErrNotFound {
			httputil.Error(w, http.StatusNotFound, "showcase entry not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to update showcase entry")
		return
	}
	httputil.JSON(w, http.StatusOK, e)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		if err == ErrNotFound {
			httputil.Error(w, http.StatusNotFound, "showcase entry not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete showcase entry")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Images ---

func parseImageUpload(w http.ResponseWriter, r *http.Request) (file interface {
	Read([]byte) (int, error)
	Close() error
}, filename, contentType string, size int64, ok bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxImageUploadSize)
	if err := r.ParseMultipartForm(maxImageUploadSize); err != nil {
		httputil.Error(w, http.StatusBadRequest, "file too large or invalid multipart form")
		return nil, "", "", 0, false
	}
	f, header, err := r.FormFile("file")
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "missing file field")
		return nil, "", "", 0, false
	}
	ct := header.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		f.Close()
		httputil.Error(w, http.StatusBadRequest, "file must be an image")
		return nil, "", "", 0, false
	}
	return f, header.Filename, ct, header.Size, true
}

func (h *Handler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	if !h.checkConfigured(w) {
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	file, filename, contentType, size, ok := parseImageUpload(w, r)
	if !ok {
		return
	}
	defer file.Close()
	e, err := h.svc.UploadLogo(r.Context(), id, userID, filename, contentType, file, size)
	if err != nil {
		if err == ErrNotFound {
			httputil.Error(w, http.StatusNotFound, "showcase entry not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to upload logo")
		return
	}
	httputil.JSON(w, http.StatusOK, e)
}

func (h *Handler) UploadBanner(w http.ResponseWriter, r *http.Request) {
	if !h.checkConfigured(w) {
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	file, filename, contentType, size, ok := parseImageUpload(w, r)
	if !ok {
		return
	}
	defer file.Close()
	e, err := h.svc.UploadBanner(r.Context(), id, userID, filename, contentType, file, size)
	if err != nil {
		if err == ErrNotFound {
			httputil.Error(w, http.StatusNotFound, "showcase entry not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to upload banner")
		return
	}
	httputil.JSON(w, http.StatusOK, e)
}

func (h *Handler) DeleteLogo(w http.ResponseWriter, r *http.Request) {
	if !h.checkConfigured(w) {
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteLogo(r.Context(), id, userID); err != nil {
		if err == ErrNotFound {
			httputil.Error(w, http.StatusNotFound, "showcase entry not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete logo")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	if !h.checkConfigured(w) {
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteBanner(r.Context(), id, userID); err != nil {
		if err == ErrNotFound {
			httputil.Error(w, http.StatusNotFound, "showcase entry not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete banner")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Public (no auth) ---

func (h *Handler) PublicList(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	entries, err := h.svc.PublicList(r.Context(), slug)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "not found")
		return
	}
	httputil.JSON(w, http.StatusOK, entries)
}

func (h *Handler) LogoRedirect(w http.ResponseWriter, r *http.Request) {
	if !h.checkConfigured(w) {
		return
	}
	id := chi.URLParam(r, "id")
	url, err := h.svc.PresignLogo(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "logo not found")
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) BannerRedirect(w http.ResponseWriter, r *http.Request) {
	if !h.checkConfigured(w) {
		return
	}
	id := chi.URLParam(r, "id")
	url, err := h.svc.PresignBanner(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "banner not found")
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

// --- Embed API (AuthenticateShowcaseToken only) ---

func (h *Handler) Embed(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	entries, err := h.svc.List(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list showcase entries")
		return
	}
	httputil.JSON(w, http.StatusOK, entries)
}

// --- Token management (JWT-only) ---

func (h *Handler) ListTokens(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	tokens, err := h.svc.ListTokens(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to list showcase tokens")
		return
	}
	httputil.JSON(w, http.StatusOK, tokens)
}

func (h *Handler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var req CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := validate.Struct(req); err != nil {
		writeValidationErr(w, err)
		return
	}
	result, err := h.svc.GenerateToken(r.Context(), userID, req.Name)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "failed to generate showcase token")
		return
	}
	httputil.JSON(w, http.StatusCreated, result)
}

func (h *Handler) DeleteToken(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteToken(r.Context(), id, userID); err != nil {
		if err == ErrNotFound {
			httputil.Error(w, http.StatusNotFound, "showcase token not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete showcase token")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
