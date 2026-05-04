package bot

import (
	"sync"
	"time"
)

type cooldownKey struct {
	name    string
	channel string
	userID  string
}

type CooldownManager struct {
	mu      sync.Mutex
	entries map[cooldownKey]time.Time
}

var GlobalCooldowns = &CooldownManager{
	entries: make(map[cooldownKey]time.Time),
}

func (c *CooldownManager) IsOnCooldown(name, channel, userID string, duration time.Duration) bool {
	if duration <= 0 {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	key := cooldownKey{name, channel, userID}
	if last, ok := c.entries[key]; ok {
		if time.Since(last) < duration {
			return true
		}
	}
	return false
}

func (c *CooldownManager) Remaining(name, channel, userID string, duration time.Duration) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := cooldownKey{name, channel, userID}
	if last, ok := c.entries[key]; ok {
		remaining := duration - time.Since(last)
		if remaining > 0 {
			return remaining
		}
	}
	return 0
}

func (c *CooldownManager) Set(name, channel, userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[cooldownKey{name, channel, userID}] = time.Now()
}

func (c *CooldownManager) Reset(name, channel, userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, cooldownKey{name, channel, userID})
}
