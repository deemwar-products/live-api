package mcp

import (
	"fmt"
	"sync"
)

type ToolQuota struct {
	DailyLimit int
	Used       int
}

type RateLimiter struct {
	mu           sync.Mutex
	defaultLimit int
	quotas       map[string]ToolQuota
}

func NewRateLimiter(defaultLimit int) *RateLimiter {
	if defaultLimit <= 0 {
		defaultLimit = 1000
	}
	return &RateLimiter{defaultLimit: defaultLimit, quotas: make(map[string]ToolQuota)}
}

func (r *RateLimiter) Allow(orgID, toolName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := orgID + ":" + toolName
	quota := r.quotas[key]
	if quota.DailyLimit == 0 {
		quota.DailyLimit = r.defaultLimit
	}
	if quota.Used >= quota.DailyLimit {
		return fmt.Errorf("tool quota exceeded for org %s and tool %s", orgID, toolName)
	}
	quota.Used++
	r.quotas[key] = quota
	return nil
}

func (r *RateLimiter) Snapshot() map[string]ToolQuota {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]ToolQuota, len(r.quotas))
	for key, quota := range r.quotas {
		out[key] = quota
	}
	return out
}

func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.quotas = make(map[string]ToolQuota)
}
