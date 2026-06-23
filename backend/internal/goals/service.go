package goals

import "context"

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// List returns goals with key results and computed progress.
func (s *Service) List(ctx context.Context, userID string) ([]*Goal, error) {
	gs, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, g := range gs {
		s.hydrate(ctx, g)
	}
	return gs, nil
}

func (s *Service) Get(ctx context.Context, id, userID string) (*Goal, error) {
	g, err := s.repo.Get(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	s.hydrate(ctx, g)
	return g, nil
}

func (s *Service) Create(ctx context.Context, userID string, req CreateGoalRequest) (*Goal, error) {
	g, err := s.repo.Create(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	g.KeyResults = []*KeyResult{}
	return g, nil
}

func (s *Service) Update(ctx context.Context, id, userID string, req UpdateGoalRequest) (*Goal, error) {
	g, err := s.repo.Update(ctx, id, userID, req)
	if err != nil {
		return nil, err
	}
	s.hydrate(ctx, g)
	return g, nil
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *Service) AddKR(ctx context.Context, goalID, userID string, req KRRequest) (*KeyResult, error) {
	return s.repo.AddKR(ctx, goalID, userID, req)
}

func (s *Service) UpdateKR(ctx context.Context, krID, userID string, req KRRequest) (*KeyResult, error) {
	return s.repo.UpdateKR(ctx, krID, userID, req)
}

func (s *Service) DeleteKR(ctx context.Context, krID, userID string) error {
	return s.repo.DeleteKR(ctx, krID, userID)
}

func (s *Service) LinkTask(ctx context.Context, goalID, taskID, userID string) error {
	return s.repo.LinkTask(ctx, goalID, taskID, userID)
}

func (s *Service) UnlinkTask(ctx context.Context, taskID, userID string) error {
	return s.repo.UnlinkTask(ctx, taskID, userID)
}

func (s *Service) ListTasks(ctx context.Context, goalID, userID string) ([]*LinkedTask, error) {
	return s.repo.ListTasks(ctx, goalID, userID)
}

// hydrate loads key results and computes overall progress (0..100).
// task_completion KRs derive their value live from linked task status.
func (s *Service) hydrate(ctx context.Context, g *Goal) {
	krs, err := s.repo.keyResults(ctx, g.ID)
	if err != nil {
		g.KeyResults = []*KeyResult{}
		return
	}
	g.KeyResults = krs
	if len(krs) == 0 {
		// Fall back to linked-task completion if no KRs defined.
		done, total, err := s.repo.taskCompletion(ctx, g.ID)
		if err == nil && total > 0 {
			g.Progress = float64(done) / float64(total) * 100
		}
		return
	}
	var sum float64
	for _, kr := range krs {
		if kr.MetricType == "task_completion" {
			done, total, err := s.repo.taskCompletion(ctx, g.ID)
			if err == nil && total > 0 {
				kr.CurrentVal = float64(done)
				kr.TargetVal = float64(total)
			}
		}
		if kr.TargetVal > 0 {
			p := kr.CurrentVal / kr.TargetVal * 100
			if p > 100 {
				p = 100
			}
			sum += p
		}
	}
	g.Progress = sum / float64(len(krs))
}
