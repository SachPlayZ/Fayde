// Package sse implements a Server-Sent Events broker for real-time task updates.
package sse

import (
	"sync"
)

// Event is a JSON-serializable SSE payload.
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// subscriber is a single connected SSE client.
type subscriber struct {
	ch chan Event
}

// Broker fans out events to per-user and admin subscribers.
type Broker struct {
	mu          sync.RWMutex
	userSubs    map[string][]*subscriber
	adminSubs   []*subscriber
}

// NewBroker creates a new Broker.
func NewBroker() *Broker {
	return &Broker{
		userSubs: make(map[string][]*subscriber),
	}
}

// Subscribe registers a channel for the given userID.
// Returns the channel and an unsubscribe function that must be called on disconnect.
func (b *Broker) Subscribe(userID string) (chan Event, func()) {
	sub := &subscriber{ch: make(chan Event, 16)}

	b.mu.Lock()
	b.userSubs[userID] = append(b.userSubs[userID], sub)
	b.mu.Unlock()

	unsub := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subs := b.userSubs[userID]
		for i, s := range subs {
			if s == sub {
				b.userSubs[userID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		if len(b.userSubs[userID]) == 0 {
			delete(b.userSubs, userID)
		}
	}

	return sub.ch, unsub
}

// SubscribeAdmin registers a channel that receives ALL events regardless of user.
// Returns the channel and an unsubscribe function.
func (b *Broker) SubscribeAdmin() (chan Event, func()) {
	sub := &subscriber{ch: make(chan Event, 16)}

	b.mu.Lock()
	b.adminSubs = append(b.adminSubs, sub)
	b.mu.Unlock()

	unsub := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		for i, s := range b.adminSubs {
			if s == sub {
				b.adminSubs = append(b.adminSubs[:i], b.adminSubs[i+1:]...)
				break
			}
		}
	}

	return sub.ch, unsub
}

// Publish sends an event to all subscribers of the given userID and to all admin subscribers.
func (b *Broker) Publish(userID string, event Event) {
	b.mu.RLock()
	userTargets := make([]*subscriber, len(b.userSubs[userID]))
	copy(userTargets, b.userSubs[userID])
	adminTargets := make([]*subscriber, len(b.adminSubs))
	copy(adminTargets, b.adminSubs)
	b.mu.RUnlock()

	for _, sub := range userTargets {
		select {
		case sub.ch <- event:
		default:
			// drop if buffer full — prefer non-blocking
		}
	}
	for _, sub := range adminTargets {
		select {
		case sub.ch <- event:
		default:
		}
	}
}
