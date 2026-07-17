package leetcode

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	List(ctx context.Context, userID string, f ListFilter) ([]*Problem, error)
	Get(ctx context.Context, id, userID string) (*ProblemDetail, error)
	Create(ctx context.Context, userID string, req CreateRequest) (*Problem, error)
	Update(ctx context.Context, id, userID string, req UpdateRequest) (*Problem, error)
	Delete(ctx context.Context, id, userID string) error

	GetCard(ctx context.Context, problemID, userID string) (*CardState, bool, error)
	SubmitReview(ctx context.Context, card *CardState, log *ReviewLog, markSolved bool) error

	Queue(ctx context.Context, userID string) ([]*Problem, error)
	RawStats(ctx context.Context, userID string) (*Stats, error)
	ReviewDates(ctx context.Context, userID string) ([]time.Time, error)
	DueTodayCount(ctx context.Context, userID string) (int, error)

	owns(ctx context.Context, id, userID string) (bool, error)
}

type pgRepository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &pgRepository{pool: pool} }

const problemJoinCols = `p.id, p.user_id, p.lc_number, p.slug, p.title, p.url, p.difficulty,
	p.topics, p.notes, p.solved_at, p.created_at, p.updated_at,
	c.id, c.problem_id, c.user_id, c.stability, c.difficulty, c.elapsed_days, c.scheduled_days,
	c.reps, c.lapses, c.card_state, c.due_date, c.last_review`

func scanProblemWithCard(row pgx.Row) (*Problem, error) {
	p := &Problem{}
	var (
		cardID, cardProblemID, cardUserID *string
		cardStability, cardDifficulty     *float64
		cardElapsed, cardScheduled        *int
		cardReps, cardLapses              *int
		cardState                         *string
		cardDue                           *time.Time
		cardLastReview                    *time.Time
	)
	err := row.Scan(
		&p.ID, &p.UserID, &p.LCNumber, &p.Slug, &p.Title, &p.URL, &p.Difficulty,
		&p.Topics, &p.Notes, &p.SolvedAt, &p.CreatedAt, &p.UpdatedAt,
		&cardID, &cardProblemID, &cardUserID, &cardStability, &cardDifficulty,
		&cardElapsed, &cardScheduled, &cardReps, &cardLapses, &cardState, &cardDue, &cardLastReview,
	)
	if err != nil {
		return nil, err
	}
	if cardID != nil {
		p.Card = &CardState{
			ID:            *cardID,
			ProblemID:     *cardProblemID,
			UserID:        *cardUserID,
			Stability:     *cardStability,
			Difficulty:    *cardDifficulty,
			ElapsedDays:   *cardElapsed,
			ScheduledDays: *cardScheduled,
			Reps:          *cardReps,
			Lapses:        *cardLapses,
			State:         *cardState,
			DueDate:       *cardDue,
			LastReview:    cardLastReview,
		}
	}
	return p, nil
}

func (r *pgRepository) List(ctx context.Context, userID string, f ListFilter) ([]*Problem, error) {
	where := []string{"p.user_id=$1"}
	args := []any{userID}
	idx := 2
	if f.Difficulty != "" {
		where = append(where, fmt.Sprintf("p.difficulty=$%d", idx))
		args = append(args, f.Difficulty)
		idx++
	}
	if f.Topic != "" {
		where = append(where, fmt.Sprintf("$%d = ANY(p.topics)", idx))
		args = append(args, f.Topic)
		idx++
	}
	switch f.Status {
	case "due":
		where = append(where, "c.due_date <= now()")
	case "overdue":
		where = append(where, "c.due_date < date_trunc('day', now())")
	}
	q := `SELECT ` + problemJoinCols + `
		FROM leetcode_problems p
		LEFT JOIN leetcode_cards c ON c.problem_id = p.id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY CASE WHEN c.due_date <= now() THEN 0 ELSE 1 END, c.due_date ASC NULLS LAST, p.created_at DESC
		LIMIT 500`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("leetcode: list: %w", err)
	}
	defer rows.Close()
	out := []*Problem{}
	for rows.Next() {
		p, err := scanProblemWithCard(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *pgRepository) Get(ctx context.Context, id, userID string) (*ProblemDetail, error) {
	q := `SELECT ` + problemJoinCols + `
		FROM leetcode_problems p
		LEFT JOIN leetcode_cards c ON c.problem_id = p.id
		WHERE p.id=$1 AND p.user_id=$2`
	p, err := scanProblemWithCard(r.pool.QueryRow(ctx, q, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("leetcode: get: %w", err)
	}

	reviews := []*ReviewLog{}
	if p.Card != nil {
		rows, err := r.pool.Query(ctx,
			`SELECT id, card_id, user_id, reviewed_at, rating, scheduled_days, elapsed_days, stability, difficulty, card_state
			 FROM leetcode_review_logs WHERE card_id=$1 ORDER BY reviewed_at DESC`, p.Card.ID)
		if err != nil {
			return nil, fmt.Errorf("leetcode: get reviews: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			l := &ReviewLog{}
			if err := rows.Scan(&l.ID, &l.CardID, &l.UserID, &l.ReviewedAt, &l.Rating,
				&l.ScheduledDays, &l.ElapsedDays, &l.Stability, &l.Difficulty, &l.State); err != nil {
				return nil, err
			}
			reviews = append(reviews, l)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return &ProblemDetail{Problem: p, Reviews: reviews}, nil
}

func (r *pgRepository) Create(ctx context.Context, userID string, req CreateRequest) (*Problem, error) {
	difficulty := req.Difficulty
	if difficulty == "" {
		difficulty = DifficultyMedium
	}
	topics := req.Topics
	if topics == nil {
		topics = []string{}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("leetcode: create begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	p := &Problem{}
	err = tx.QueryRow(ctx,
		`INSERT INTO leetcode_problems (user_id, lc_number, slug, title, url, difficulty, topics, notes)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 RETURNING id, user_id, lc_number, slug, title, url, difficulty, topics, notes, solved_at, created_at, updated_at`,
		userID, req.LCNumber, req.Slug, req.Title, req.URL, difficulty, topics, req.Notes,
	).Scan(&p.ID, &p.UserID, &p.LCNumber, &p.Slug, &p.Title, &p.URL, &p.Difficulty,
		&p.Topics, &p.Notes, &p.SolvedAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("leetcode: create problem: %w", err)
	}

	var cardID string
	err = tx.QueryRow(ctx,
		`INSERT INTO leetcode_cards (problem_id, user_id) VALUES ($1,$2) RETURNING id`,
		p.ID, userID,
	).Scan(&cardID)
	if err != nil {
		return nil, fmt.Errorf("leetcode: create card: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("leetcode: create commit: %w", err)
	}

	p.Card = &CardState{ID: cardID, ProblemID: p.ID, UserID: userID, State: StateNew, DueDate: time.Now().UTC()}
	return p, nil
}

func (r *pgRepository) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Problem, error) {
	sets := []string{"updated_at=now()"}
	args := []any{}
	idx := 1
	add := func(frag string, v any) {
		sets = append(sets, fmt.Sprintf(frag, idx))
		args = append(args, v)
		idx++
	}
	if req.LCNumber != nil {
		add("lc_number=$%d", *req.LCNumber)
	}
	if req.Slug != nil {
		add("slug=$%d", *req.Slug)
	}
	if req.Title != nil {
		add("title=$%d", *req.Title)
	}
	if req.URL != nil {
		add("url=$%d", *req.URL)
	}
	if req.Difficulty != nil {
		add("difficulty=$%d", *req.Difficulty)
	}
	if req.Topics != nil {
		add("topics=$%d", *req.Topics)
	}
	if req.Notes != nil {
		add("notes=$%d", *req.Notes)
	}
	q := fmt.Sprintf(`UPDATE leetcode_problems SET %s WHERE id=$%d AND user_id=$%d
		RETURNING id, user_id, lc_number, slug, title, url, difficulty, topics, notes, solved_at, created_at, updated_at`,
		strings.Join(sets, ", "), idx, idx+1)
	args = append(args, id, userID)

	p := &Problem{}
	err := r.pool.QueryRow(ctx, q, args...).Scan(&p.ID, &p.UserID, &p.LCNumber, &p.Slug, &p.Title, &p.URL,
		&p.Difficulty, &p.Topics, &p.Notes, &p.SolvedAt, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("leetcode: update: %w", err)
	}
	return p, nil
}

func (r *pgRepository) Delete(ctx context.Context, id, userID string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM leetcode_problems WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return fmt.Errorf("leetcode: delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgRepository) GetCard(ctx context.Context, problemID, userID string) (*CardState, bool, error) {
	c := &CardState{}
	var isSolved bool
	err := r.pool.QueryRow(ctx,
		`SELECT c.id, c.problem_id, c.user_id, c.stability, c.difficulty, c.elapsed_days, c.scheduled_days,
		        c.reps, c.lapses, c.card_state, c.due_date, c.last_review, p.solved_at IS NOT NULL
		 FROM leetcode_cards c
		 JOIN leetcode_problems p ON p.id = c.problem_id
		 WHERE c.problem_id=$1 AND p.user_id=$2`, problemID, userID,
	).Scan(&c.ID, &c.ProblemID, &c.UserID, &c.Stability, &c.Difficulty, &c.ElapsedDays, &c.ScheduledDays,
		&c.Reps, &c.Lapses, &c.State, &c.DueDate, &c.LastReview, &isSolved)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, ErrNotFound
	}
	if err != nil {
		return nil, false, fmt.Errorf("leetcode: get card: %w", err)
	}
	return c, isSolved, nil
}

func (r *pgRepository) SubmitReview(ctx context.Context, card *CardState, log *ReviewLog, markSolved bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("leetcode: submit review begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx,
		`UPDATE leetcode_cards SET stability=$1, difficulty=$2, elapsed_days=$3, scheduled_days=$4,
		 reps=$5, lapses=$6, card_state=$7, due_date=$8, last_review=$9 WHERE id=$10`,
		card.Stability, card.Difficulty, card.ElapsedDays, card.ScheduledDays,
		card.Reps, card.Lapses, card.State, card.DueDate, card.LastReview, card.ID)
	if err != nil {
		return fmt.Errorf("leetcode: submit review update card: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO leetcode_review_logs
		 (card_id, user_id, reviewed_at, rating, scheduled_days, elapsed_days, stability, difficulty, card_state)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		card.ID, card.UserID, log.ReviewedAt, log.Rating, log.ScheduledDays, log.ElapsedDays,
		log.Stability, log.Difficulty, log.State)
	if err != nil {
		return fmt.Errorf("leetcode: submit review insert log: %w", err)
	}

	if markSolved {
		_, err = tx.Exec(ctx,
			`UPDATE leetcode_problems SET solved_at=now(), updated_at=now() WHERE id=$1 AND solved_at IS NULL`,
			card.ProblemID)
		if err != nil {
			return fmt.Errorf("leetcode: submit review mark solved: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *pgRepository) Queue(ctx context.Context, userID string) ([]*Problem, error) {
	q := `SELECT ` + problemJoinCols + `
		FROM leetcode_problems p
		JOIN leetcode_cards c ON c.problem_id = p.id
		WHERE p.user_id=$1 AND c.due_date <= now()
		ORDER BY c.due_date ASC
		LIMIT 200`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("leetcode: queue: %w", err)
	}
	defer rows.Close()
	out := []*Problem{}
	for rows.Next() {
		p, err := scanProblemWithCard(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *pgRepository) DueTodayCount(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM leetcode_cards WHERE user_id=$1 AND due_date <= now()`, userID).Scan(&n)
	return n, err
}

func (r *pgRepository) RawStats(ctx context.Context, userID string) (*Stats, error) {
	s := &Stats{}
	err := r.pool.QueryRow(ctx,
		`SELECT
			COUNT(*)                                                          AS total_problems,
			COUNT(*) FILTER (WHERE p.solved_at IS NOT NULL)                  AS total_solved,
			COUNT(*) FILTER (WHERE c.due_date <= now())                      AS due_today_count,
			COUNT(*) FILTER (WHERE c.due_date < date_trunc('day', now()))    AS overdue_count,
			(SELECT COUNT(*) FROM leetcode_review_logs WHERE user_id=$1)     AS total_reviews,
			(SELECT COALESCE(
				COUNT(*) FILTER (WHERE rating >= 3)::float / NULLIF(COUNT(*), 0), 0
			) FROM leetcode_review_logs WHERE user_id=$1)                    AS retention_rate
		 FROM leetcode_problems p
		 LEFT JOIN leetcode_cards c ON c.problem_id = p.id
		 WHERE p.user_id=$1`, userID,
	).Scan(&s.TotalProblems, &s.TotalSolved, &s.DueTodayCount, &s.OverdueCount, &s.TotalReviews, &s.RetentionRate)
	if err != nil {
		return nil, fmt.Errorf("leetcode: raw stats: %w", err)
	}
	return s, nil
}

func (r *pgRepository) ReviewDates(ctx context.Context, userID string) ([]time.Time, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT reviewed_at::date FROM leetcode_review_logs WHERE user_id=$1 ORDER BY 1 DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("leetcode: review dates: %w", err)
	}
	defer rows.Close()
	var out []time.Time
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *pgRepository) owns(ctx context.Context, id, userID string) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM leetcode_problems WHERE id=$1 AND user_id=$2)`, id, userID).Scan(&ok)
	return ok, err
}
