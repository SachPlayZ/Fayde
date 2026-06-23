package webpush

import "time"

// Subscription is a browser Web Push subscription owned by a user.
type Subscription struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Endpoint  string    `json:"endpoint"`
	P256dh    string    `json:"p256dh"`
	Auth      string    `json:"auth"`
	CreatedAt time.Time `json:"created_at"`
}

// SubscribeRequest is the browser PushSubscription payload.
type SubscribeRequest struct {
	Endpoint string `json:"endpoint" validate:"required,url"`
	Keys     struct {
		P256dh string `json:"p256dh" validate:"required"`
		Auth   string `json:"auth"   validate:"required"`
	} `json:"keys"`
}

// Payload is the JSON body delivered to the service worker push handler.
type Payload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url,omitempty"`
	Tag   string `json:"tag,omitempty"`
}
