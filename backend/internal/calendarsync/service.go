package calendarsync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/SachPlayZ/rivz-asn/backend/internal/tasks"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Service implements Google Calendar integration.
type Service struct {
	repo        *Repository
	googleConfig *oauth2.Config
	frontendURL string
	jwtSecret   string
}

// NewService creates a new calendarsync Service.
func NewService(
	repo *Repository,
	googleClientID, googleClientSecret,
	appURL, frontendURL, jwtSecret string,
) *Service {
	cfg := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  appURL + "/calendar/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.events",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

	return &Service{
		repo:         repo,
		googleConfig: cfg,
		frontendURL:  frontendURL,
		jwtSecret:    jwtSecret,
	}
}

// GetStatus checks if the user is connected to Google Calendar.
func (s *Service) GetStatus(ctx context.Context, userID string) (*ConnectionStatus, error) {
	conn, err := s.repo.GetConnection(ctx, userID)
	if errors.Is(err, ErrNotFound) {
		return &ConnectionStatus{Connected: false}, nil
	}
	if err != nil {
		return nil, err
	}
	return &ConnectionStatus{
		Connected: true,
		Email:     conn.Email,
	}, nil
}

// Disconnect removes the Google Calendar connection.
func (s *Service) Disconnect(ctx context.Context, userID string) error {
	return s.repo.DeleteConnection(ctx, userID)
}

// Connect exchanges code for a token and saves it.
func (s *Service) Connect(ctx context.Context, userID, authCode string) error {
	tok, err := s.googleConfig.Exchange(ctx, authCode)
	if err != nil {
		return fmt.Errorf("exchange code: %w", err)
	}

	email, err := s.fetchGoogleEmail(ctx, tok.AccessToken)
	if err != nil {
		return fmt.Errorf("fetch google email: %w", err)
	}

	// We want to request offline access (prompt=consent, access_type=offline) during redirect
	// to make sure we always get a refresh token.
	_, err = s.repo.SaveConnection(ctx, userID, email, tok.AccessToken, tok.RefreshToken, tok.Expiry)
	return err
}

func (s *Service) fetchGoogleEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}

	var res struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.Email, nil
}

// Structs for Google Calendar API
type eventDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone,omitempty"`
}

type googleEvent struct {
	ID          string        `json:"id,omitempty"`
	Summary     string        `json:"summary"`
	Description string        `json:"description"`
	Start       eventDateTime `json:"start"`
	End         eventDateTime `json:"end"`
}

// SyncTask pushes a task to Google Calendar.
func (s *Service) SyncTask(ctx context.Context, task *tasks.Task) error {
	conn, err := s.repo.GetConnection(ctx, task.UserID)
	if errors.Is(err, ErrNotFound) {
		return nil // Not connected, ignore silently
	}
	if err != nil {
		return err
	}

	ts := s.googleConfig.TokenSource(ctx, &oauth2.Token{
		AccessToken:  conn.AccessToken,
		RefreshToken: conn.RefreshToken,
		Expiry:       conn.Expiry,
	})

	refreshed, err := ts.Token()
	if err != nil {
		return fmt.Errorf("get token: %w", err)
	}

	if refreshed.AccessToken != conn.AccessToken {
		_ = s.repo.UpdateAccessToken(ctx, task.UserID, refreshed.AccessToken, refreshed.Expiry)
	}

	client := oauth2.NewClient(ctx, ts)

	// Case 1: No due date
	if task.DueDate == nil {
		if task.ParentTaskID != nil {
			// Subtask or clone: we do not sync these if they don't have a due date
			return nil
		}
		// If it was previously synced, delete it from calendar
		// Wait, let's search if task has an external event id.
		// Since s.repo can query the task directly or we check task.ExternalEventID
		// In tasks package, we modify repository to load t.external_event_id.
		// So task.ExternalEventID is populated.
		// But wait! We need to make sure we query or pass Task with ExternalEventID.
		// Wait, does tasks.Task have a field for ExternalEventID? Yes, we will add it.
		// Let's see: we should make sure that if task.ExternalEventID is not empty/nil, we delete it.
		// Wait, let's query the db directly to check if this task has an external_event_id stored in the DB, just to be completely safe from race conditions or unpopulated fields!
		var extEventID *string
		err = s.repo.pool.QueryRow(ctx, `SELECT external_event_id FROM tasks WHERE id=$1`, task.ID).Scan(&extEventID)
		if err != nil {
			return err
		}
		if extEventID != nil && *extEventID != "" {
			if err := s.deleteEvent(client, *extEventID); err != nil {
				log.Printf("calendarsync: delete event %s error: %v", *extEventID, err)
			}
			_ = s.repo.UpdateTaskExternalEventID(ctx, task.ID, nil)
		}
		return nil
	}

	// Case 2: Has due date -> Sync event
	var extEventID *string
	err = s.repo.pool.QueryRow(ctx, `SELECT external_event_id FROM tasks WHERE id=$1`, task.ID).Scan(&extEventID)
	if err != nil {
		return err
	}

	title := task.Title
	if task.Status == "done" || task.Status == "failed" {
		title = fmt.Sprintf("[%s] %s", strings.ToUpper(task.Status[:1])+task.Status[1:], task.Title)
	}

	desc := task.Description
	if s.frontendURL != "" {
		desc = fmt.Sprintf("%s\n\nLink to task: %s/tasks/%s", desc, s.frontendURL, task.ID)
	}

	startTime := task.DueDate.UTC().Format(time.RFC3339)
	endTime := task.DueDate.UTC().Add(30 * time.Minute).Format(time.RFC3339)

	eventData := googleEvent{
		Summary:     title,
		Description: desc,
		Start:       eventDateTime{DateTime: startTime},
		End:         eventDateTime{DateTime: endTime},
	}

	bodyBytes, _ := json.Marshal(eventData)

	if extEventID != nil && *extEventID != "" {
		// Update existing event
		url := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/primary/events/%s", *extEventID)
		req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			// If not found, it might have been deleted manually by the user on Google Calendar.
			// Let's create a new event.
			*extEventID = ""
		} else if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("google calendar returned status %d: %s", resp.StatusCode, string(body))
		}
	}

	if extEventID == nil || *extEventID == "" {
		// Create new event
		url := "https://www.googleapis.com/calendar/v3/calendars/primary/events"
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("google calendar returned status %d: %s", resp.StatusCode, string(body))
		}

		var created googleEvent
		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			return err
		}

		_ = s.repo.UpdateTaskExternalEventID(ctx, task.ID, &created.ID)
	}

	return nil
}

// DeleteEvent removes a Google Calendar event for a user.
func (s *Service) DeleteEvent(ctx context.Context, userID, eventID string) error {
	conn, err := s.repo.GetConnection(ctx, userID)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	ts := s.googleConfig.TokenSource(ctx, &oauth2.Token{
		AccessToken:  conn.AccessToken,
		RefreshToken: conn.RefreshToken,
		Expiry:       conn.Expiry,
	})

	refreshed, err := ts.Token()
	if err != nil {
		return err
	}

	if refreshed.AccessToken != conn.AccessToken {
		_ = s.repo.UpdateAccessToken(ctx, userID, refreshed.AccessToken, refreshed.Expiry)
	}

	client := oauth2.NewClient(ctx, ts)
	return s.deleteEvent(client, eventID)
}

func (s *Service) deleteEvent(client *http.Client, eventID string) error {
	url := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/primary/events/%s", eventID)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("google calendar delete status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
