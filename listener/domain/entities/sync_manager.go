package entities

import (
	"sync"
	"time"
)

type SyncManager struct {
	mu      sync.Mutex
	ignored map[string]time.Time
}

func NewSyncManager() *SyncManager {
	return &SyncManager{
		ignored: make(map[string]time.Time),
	}
}

func (this *SyncManager) Ignore(path string) {
	this.mu.Lock()
	defer this.mu.Unlock()

	this.ignored[path] = time.Now().UTC().Add(2 * time.Second)
}

func (this *SyncManager) ShouldIgnore(path string) bool {
	this.mu.Lock()
	defer this.mu.Unlock()

	expiresAt, exists := this.ignored[path]
	if !exists {
		return false
	}

	if time.Now().UTC().Before(expiresAt) {
		return true
	}

	delete(this.ignored, path)
	return false
}
