package tags

type Tag struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
}
