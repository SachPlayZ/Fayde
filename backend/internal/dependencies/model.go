package dependencies

type Dependency struct {
	TaskID      string `json:"task_id"`
	DependsOnID string `json:"depends_on_id"`
	Title       string `json:"title,omitempty"`
}

type DependencyList struct {
	BlockedBy []Dependency `json:"blocked_by"`
	Blocking  []Dependency `json:"blocking"`
}
