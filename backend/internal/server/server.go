package server

import (
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/admin"
	"github.com/SachPlayZ/rivz-asn/backend/internal/attachments"
	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/comments"
	"github.com/SachPlayZ/rivz-asn/backend/internal/dependencies"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
	"github.com/SachPlayZ/rivz-asn/backend/internal/notifications"
	"github.com/SachPlayZ/rivz-asn/backend/internal/sse"
	"github.com/SachPlayZ/rivz-asn/backend/internal/subtasks"
	"github.com/SachPlayZ/rivz-asn/backend/internal/tags"
	"github.com/SachPlayZ/rivz-asn/backend/internal/tasks"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// New builds and returns a configured chi router with all routes registered.
func New(
	cfg ServerConfig,
	authHandler *auth.Handler,
	tasksHandler *tasks.Handler,
	adminHandler *admin.Handler,
	sseHandler *sse.Handler,
	attachmentsHandler *attachments.Handler,
	subtasksHandler *subtasks.Handler,
	tagsHandler *tags.Handler,
	commentsHandler *comments.Handler,
	depsHandler *dependencies.Handler,
	notifHandler *notifications.Handler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Cache-Control"},
		ExposedHeaders:   []string{"Link", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth routes.
	r.Post("/auth/signup", authHandler.Signup)
	r.Post("/auth/login", authHandler.Login)
	r.With(auth.Authenticate(cfg.JWTSecret)).Get("/auth/me", authHandler.Me)
	r.With(auth.Authenticate(cfg.JWTSecret)).Patch("/auth/me/preferences", authHandler.UpdatePreferences)

	// SSE.
	r.Get("/events", sseHandler.ServeSSE)

	// JWT-protected routes.
	r.Group(func(r chi.Router) {
		r.Use(auth.Authenticate(cfg.JWTSecret))

		r.Get("/activity", tasksHandler.GetUserActivity)

		// Tasks.
		r.Post("/tasks/bulk-update", tasksHandler.BulkUpdate)
		r.Post("/tasks/bulk-delete", tasksHandler.BulkDelete)
		r.Put("/tasks/reorder", tasksHandler.Reorder)
		r.Post("/tasks", tasksHandler.Create)
		r.Get("/tasks", tasksHandler.List)
		r.Get("/tasks/{id}", tasksHandler.Get)
		r.Patch("/tasks/{id}", tasksHandler.Update)
		r.Delete("/tasks/{id}", tasksHandler.Delete)
		r.Get("/tasks/{id}/activity", tasksHandler.GetActivity)

		// Attachments.
		r.Post("/tasks/{id}/attachments", attachmentsHandler.Upload)
		r.Get("/tasks/{id}/attachments", attachmentsHandler.List)
		r.Delete("/tasks/{id}/attachments/{attId}", attachmentsHandler.Delete)

		// Subtasks.
		r.Get("/tasks/{id}/subtasks", subtasksHandler.List)
		r.Post("/tasks/{id}/subtasks", subtasksHandler.Create)
		r.Patch("/tasks/{id}/subtasks/{subId}", subtasksHandler.Update)
		r.Delete("/tasks/{id}/subtasks/{subId}", subtasksHandler.Delete)
		r.Put("/tasks/{id}/subtasks/order", subtasksHandler.Reorder)

		// Tags.
		r.Get("/tags", tagsHandler.List)
		r.Post("/tags", tagsHandler.Create)
		r.Delete("/tags/{id}", tagsHandler.Delete)
		r.Post("/tasks/{id}/tags", tagsHandler.AddToTask)
		r.Delete("/tasks/{id}/tags/{tagId}", tagsHandler.RemoveFromTask)

		// Comments.
		r.Get("/tasks/{id}/comments", commentsHandler.List)
		r.Post("/tasks/{id}/comments", commentsHandler.Create)
		r.Patch("/tasks/{id}/comments/{cId}", commentsHandler.Update)
		r.Delete("/tasks/{id}/comments/{cId}", commentsHandler.Delete)

		// Dependencies.
		r.Get("/tasks/{id}/dependencies", depsHandler.Get)
		r.Post("/tasks/{id}/dependencies", depsHandler.Add)
		r.Delete("/tasks/{id}/dependencies/{depId}", depsHandler.Remove)

		// Notifications.
		r.Get("/notifications", notifHandler.List)
		r.Patch("/notifications/{id}/read", notifHandler.MarkRead)
		r.Post("/notifications/read-all", notifHandler.MarkAllRead)
		r.Get("/notifications/unread-count", notifHandler.UnreadCount)
	})

	// Admin routes.
	r.Group(func(r chi.Router) {
		r.Use(auth.Authenticate(cfg.JWTSecret))
		r.Use(auth.RequireAdmin)

		r.Get("/admin/tasks", adminHandler.ListTasks)
		r.Get("/admin/users", adminHandler.ListUsers)
		r.Get("/admin/analytics", adminHandler.Analytics)
	})

	return r
}

// ServerConfig holds the configuration needed to build the server.
type ServerConfig struct {
	JWTSecret  string
	CORSOrigin string
}
