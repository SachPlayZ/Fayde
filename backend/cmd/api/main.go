package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SachPlayZ/rivz-asn/backend/internal/activitylog"
	"github.com/SachPlayZ/rivz-asn/backend/internal/admin"
	"github.com/SachPlayZ/rivz-asn/backend/internal/attachments"
	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/comments"
	"github.com/SachPlayZ/rivz-asn/backend/internal/config"
	"github.com/SachPlayZ/rivz-asn/backend/internal/db"
	"github.com/SachPlayZ/rivz-asn/backend/internal/dependencies"
	"github.com/SachPlayZ/rivz-asn/backend/internal/notifications"
	"github.com/SachPlayZ/rivz-asn/backend/internal/scheduler"
	"github.com/SachPlayZ/rivz-asn/backend/internal/server"
	"github.com/SachPlayZ/rivz-asn/backend/internal/sse"
	"github.com/SachPlayZ/rivz-asn/backend/internal/subtasks"
	"github.com/SachPlayZ/rivz-asn/backend/internal/tags"
	"github.com/SachPlayZ/rivz-asn/backend/internal/tasks"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()

	migrateURL := toPgx5URL(cfg.DatabaseURL)
	if err := db.RunMigrations(migrateURL); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	// Auth.
	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authSvc)

	// Activity log.
	activityRepo := activitylog.NewRepository(pool)
	activitySvc := activitylog.NewService(activityRepo)

	// SSE.
	sseBroker := sse.NewBroker()
	sseHandler := sse.NewHandler(sseBroker, cfg.JWTSecret)

	// Notifications (wired first; injected into tasks, comments, deps).
	notifRepo := notifications.NewRepository(pool)
	notifSvc := notifications.NewService(notifRepo, sseBroker)
	notifHandler := notifications.NewHandler(notifSvc)

	// Tasks.
	tasksRepo := tasks.NewRepository(pool)
	tasksSvc := tasks.NewService(tasksRepo, activitySvc, sseBroker)
	tasksSvc.SetNotificationsService(notifSvc)
	tasksHandler := tasks.NewHandler(tasksSvc, activitySvc)

	// Admin.
	adminHandler := admin.NewHandler(pool)

	// Attachments.
	attachmentsRepo := attachments.NewRepository(pool)
	var s3Client *attachments.S3Client
	if cfg.S3Bucket != "" {
		s3Client, err = attachments.NewS3Client(
			context.Background(),
			cfg.AWSRegion, cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, cfg.S3Bucket,
		)
		if err != nil {
			return fmt.Errorf("init s3 client: %w", err)
		}
	}
	attachmentsSvc := attachments.NewService(attachmentsRepo, s3Client)
	attachmentsHandler := attachments.NewHandler(attachmentsSvc, tasksSvc, cfg.S3Bucket)

	// Subtasks.
	subtasksRepo := subtasks.NewRepository(pool)
	subtasksSvc := subtasks.NewService(subtasksRepo)
	subtasksHandler := subtasks.NewHandler(subtasksSvc, tasksSvc)

	// Tags.
	tagsRepo := tags.NewRepository(pool)
	tagsSvc := tags.NewService(tagsRepo)
	tagsHandler := tags.NewHandler(tagsSvc)

	// Comments.
	commentsRepo := comments.NewRepository(pool)
	commentsSvc := comments.NewService(commentsRepo, notifSvc, pool)
	commentsHandler := comments.NewHandler(commentsSvc)

	// Dependencies.
	depsRepo := dependencies.NewRepository(pool)
	depsSvc := dependencies.NewService(depsRepo, notifSvc)
	depsHandler := dependencies.NewHandler(depsSvc)
	tasksSvc.SetDependenciesService(depsSvc)

	// Scheduler.
	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())
	defer schedulerCancel()
	go scheduler.Start(schedulerCtx, pool, notifSvc, cfg)

	handler := server.New(server.ServerConfig{
		JWTSecret:  cfg.JWTSecret,
		CORSOrigin: cfg.CORSOrigin,
	}, authHandler, tasksHandler, adminHandler, sseHandler, attachmentsHandler,
		subtasksHandler, tagsHandler, commentsHandler, depsHandler, notifHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")
	schedulerCancel()

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	log.Println("server stopped")
	return nil
}

func toPgx5URL(u string) string {
	for _, prefix := range []string{"postgresql://", "postgres://"} {
		if len(u) > len(prefix) && u[:len(prefix)] == prefix {
			return "pgx5://" + u[len(prefix):]
		}
	}
	return u
}
