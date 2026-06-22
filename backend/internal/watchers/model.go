package watchers

import "errors"

var ErrNotFound = errors.New("not found")
var ErrAlreadyWatching = errors.New("already watching")

type Watcher struct {
	TaskID    string `json:"task_id"`
	UserID    string `json:"user_id"`
	UserEmail string `json:"user_email"`
}
