package groq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
)

// TaskSummary is the minimal task representation used for AI features.
type TaskSummary struct {
	ID       string
	Title    string
	Status   string
	DueDate  *time.Time
	Priority string
}

// TasksFetcher is the interface for fetching tasks for AI analysis.
type TasksFetcher interface {
	ListForAI(ctx context.Context, userID string) ([]*TaskSummary, error)
}

// Handler handles HTTP requests for AI/Groq endpoints.
type Handler struct {
	client       *Client
	tasksFetcher TasksFetcher
}

// NewHandler creates a new Groq AI Handler.
func NewHandler(client *Client, tasksFetcher TasksFetcher) *Handler {
	return &Handler{client: client, tasksFetcher: tasksFetcher}
}

// ParseTask handles POST /ai/parse-task.
func (h *Handler) ParseTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Text == "" {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}

	content, err := h.client.Chat(r.Context(), []Message{
		{Role: "system", Content: `You are a task parser. Extract task details from natural language. Respond with ONLY valid JSON: {"title": string, "description": string, "priority": "low"|"medium"|"high"|null, "tags": [string], "due_date": "YYYY-MM-DD"|null}`},
		{Role: "user", Content: body.Text},
	})
	if err != nil {
		httputil.JSON(w, 502, map[string]string{"error": "AI request failed"})
		return
	}

	var result any
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		httputil.JSON(w, 200, map[string]string{"raw": content})
		return
	}
	httputil.JSON(w, 200, result)
}

// BreakdownTask handles POST /ai/breakdown.
func (h *Handler) BreakdownTask(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}

	userMsg := "Title: " + body.Title + "\nDescription: " + body.Description
	content, err := h.client.Chat(r.Context(), []Message{
		{Role: "system", Content: `You are a task planner. Break down the task into 3-7 concrete subtasks. Respond with ONLY valid JSON: {"subtasks": [{"title": string}]}`},
		{Role: "user", Content: userMsg},
	})
	if err != nil {
		httputil.JSON(w, 502, map[string]string{"error": "AI request failed"})
		return
	}

	var result any
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		httputil.JSON(w, 200, map[string]string{"raw": content})
		return
	}
	httputil.JSON(w, 200, result)
}

// SuggestTags handles POST /ai/suggest-tags.
func (h *Handler) SuggestTags(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}

	userMsg := "Title: " + body.Title + "\nDescription: " + body.Description
	content, err := h.client.Chat(r.Context(), []Message{
		{Role: "system", Content: `Suggest 2-5 relevant tags for this task. Respond with ONLY valid JSON: {"tags": [string]}`},
		{Role: "user", Content: userMsg},
	})
	if err != nil {
		httputil.JSON(w, 502, map[string]string{"error": "AI request failed"})
		return
	}

	var result any
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		httputil.JSON(w, 200, map[string]string{"raw": content})
		return
	}
	httputil.JSON(w, 200, result)
}

// SuggestPriority handles POST /ai/suggest-priority.
func (h *Handler) SuggestPriority(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}

	userMsg := "Title: " + body.Title + "\nDescription: " + body.Description
	content, err := h.client.Chat(r.Context(), []Message{
		{Role: "system", Content: `Determine priority for this task. Respond with ONLY valid JSON: {"priority": "low"|"medium"|"high", "reasoning": string}`},
		{Role: "user", Content: userMsg},
	})
	if err != nil {
		httputil.JSON(w, 502, map[string]string{"error": "AI request failed"})
		return
	}

	var result any
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		httputil.JSON(w, 200, map[string]string{"raw": content})
		return
	}
	httputil.JSON(w, 200, result)
}

// ExpandDescription handles POST /ai/expand-description.
func (h *Handler) ExpandDescription(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	var body struct {
		Title   string `json:"title"`
		Bullets string `json:"bullets"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}

	userMsg := "Title: " + body.Title + "\nBullets:\n" + body.Bullets
	content, err := h.client.Chat(r.Context(), []Message{
		{Role: "system", Content: `Expand the task bullets into a clear, professional description. Respond with ONLY valid JSON: {"description": string}`},
		{Role: "user", Content: userMsg},
	})
	if err != nil {
		httputil.JSON(w, 502, map[string]string{"error": "AI request failed"})
		return
	}

	var result any
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		httputil.JSON(w, 200, map[string]string{"raw": content})
		return
	}
	httputil.JSON(w, 200, result)
}

// EstimateTime handles POST /ai/estimate-time.
func (h *Handler) EstimateTime(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.JSON(w, 400, map[string]string{"error": "invalid body"})
		return
	}

	userMsg := "Title: " + body.Title + "\nDescription: " + body.Description
	content, err := h.client.Chat(r.Context(), []Message{
		{Role: "system", Content: `Estimate how long this task will take. Respond with ONLY valid JSON: {"estimate_seconds": number, "reasoning": string}`},
		{Role: "user", Content: userMsg},
	})
	if err != nil {
		httputil.JSON(w, 502, map[string]string{"error": "AI request failed"})
		return
	}

	var result any
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		httputil.JSON(w, 200, map[string]string{"raw": content})
		return
	}
	httputil.JSON(w, 200, result)
}

// WeeklyDigest handles GET /ai/weekly-digest.
func (h *Handler) WeeklyDigest(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	tasks, err := h.tasksFetcher.ListForAI(r.Context(), userID)
	if err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to fetch tasks"})
		return
	}

	summary := buildTaskSummary(tasks)
	content, err := h.client.Chat(r.Context(), []Message{
		{Role: "system", Content: "You are a productivity assistant. Summarize the user's tasks for the week in a helpful, encouraging way. Format as markdown."},
		{Role: "user", Content: "My tasks:\n" + summary},
	})
	if err != nil {
		httputil.JSON(w, 502, map[string]string{"error": "AI request failed"})
		return
	}

	httputil.JSON(w, 200, map[string]string{"digest": content})
}

// PlanDay handles POST /ai/plan-day. It produces a time-boxed schedule for
// today's open tasks, respecting an optional working-hours window.
func (h *Handler) PlanDay(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	var body struct {
		StartHour int `json:"start_hour"`
		EndHour   int `json:"end_hour"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.StartHour == 0 && body.EndHour == 0 {
		body.StartHour, body.EndHour = 9, 17
	}

	tasks, err := h.tasksFetcher.ListForAI(r.Context(), userID)
	if err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to fetch tasks"})
		return
	}

	var open []*TaskSummary
	for _, t := range tasks {
		if t.Status != "done" && t.Status != "failed" {
			open = append(open, t)
		}
	}
	summary := buildTaskSummary(open)

	content, err := h.client.Chat(r.Context(), []Message{
		{Role: "system", Content: "You are a focus coach. Build a realistic time-boxed plan for today between the user's working hours. Group similar work, add short breaks, and order by priority and due date. Output concise markdown with time slots like '09:00–10:30 — <task>'. Keep it under 12 slots."},
		{Role: "user", Content: fmt.Sprintf("Working hours: %02d:00 to %02d:00.\nMy open tasks:\n%s", body.StartHour, body.EndHour, summary)},
	})
	if err != nil {
		httputil.JSON(w, 502, map[string]string{"error": "AI request failed"})
		return
	}
	httputil.JSON(w, 200, map[string]string{"plan": content})
}

// LoadAlert handles GET /ai/load-alert.
func (h *Handler) LoadAlert(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		httputil.JSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}

	tasks, err := h.tasksFetcher.ListForAI(r.Context(), userID)
	if err != nil {
		httputil.JSON(w, 500, map[string]string{"error": "failed to fetch tasks"})
		return
	}

	now := time.Now()
	weekEnd := now.AddDate(0, 0, 7)

	var overdue, dueThisWeek, inProgress int
	for _, t := range tasks {
		if t.Status == "in_progress" {
			inProgress++
		}
		if t.DueDate != nil {
			if t.DueDate.Before(now) && t.Status != "done" {
				overdue++
			} else if t.DueDate.Before(weekEnd) && t.Status != "done" {
				dueThisWeek++
			}
		}
	}

	prompt := "Task load analysis:\n" +
		"- Overdue tasks: " + itoa(overdue) + "\n" +
		"- Due this week: " + itoa(dueThisWeek) + "\n" +
		"- In progress: " + itoa(inProgress) + "\n" +
		"Analyze this task load and determine if the user is overloaded."

	aiContent, err := h.client.Chat(r.Context(), []Message{
		{Role: "system", Content: `Analyze this task load and determine if the user is overloaded. Respond with ONLY valid JSON: {"overloaded": bool, "message": string}`},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		httputil.JSON(w, 502, map[string]string{"error": "AI request failed"})
		return
	}

	var result any
	if err := json.Unmarshal([]byte(aiContent), &result); err != nil {
		httputil.JSON(w, 200, map[string]string{"raw": aiContent})
		return
	}
	httputil.JSON(w, 200, result)
}

func buildTaskSummary(tasks []*TaskSummary) string {
	var sb string
	for _, t := range tasks {
		due := "no due date"
		if t.DueDate != nil {
			due = t.DueDate.Format("2006-01-02")
		}
		sb += "- " + t.Title + " [" + t.Status + "] due: " + due + "\n"
	}
	return sb
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
