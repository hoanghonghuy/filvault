package chat

import (
	"sync"
	"time"
)

type PresenceTracker struct {
	mu          sync.RWMutex
	connections map[string]int
	lastSeen    map[string]time.Time
}

func NewPresenceTracker() *PresenceTracker {
	return &PresenceTracker{
		connections: make(map[string]int),
		lastSeen:    make(map[string]time.Time),
	}
}

func (p *PresenceTracker) Connect(userID string) (becameOnline bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connections[userID]++
	if p.connections[userID] == 1 {
		return true
	}
	return false
}

func (p *PresenceTracker) Disconnect(userID string, now time.Time) (becameOffline bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connections[userID]--
	if p.connections[userID] <= 0 {
		delete(p.connections, userID)
		p.lastSeen[userID] = now
		return true
	}
	return false
}

func (p *PresenceTracker) IsOnline(userID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connections[userID] > 0
}

func (p *PresenceTracker) GetPresence(userID string) (status string, lastSeen *time.Time) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.connections[userID] > 0 {
		return "online", nil
	}
	if t, ok := p.lastSeen[userID]; ok && !t.IsZero() {
		return "offline", &t
	}
	return "offline", nil
}

func (p *PresenceTracker) SetLastSeen(userID string, t time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !t.IsZero() {
		p.lastSeen[userID] = t
	}
}
