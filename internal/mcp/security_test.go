package mcp

import (
	"context"
	"testing"
)

func TestMCPContextRoundTrip(t *testing.T) {
	want := MCPContext{OrgID: "org-1", UserID: "user-1", SessionID: "session-1", Role: "ai"}
	ctx := WithMCPContext(context.Background(), want)
	got, err := RequireContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestRateLimiterScopesQuotaByOrgAndTool(t *testing.T) {
	limiter := NewRateLimiter(1)
	if err := limiter.Allow("org-1", "tool-a"); err != nil {
		t.Fatal(err)
	}
	if err := limiter.Allow("org-1", "tool-a"); err == nil {
		t.Fatal("expected quota error")
	}
	if err := limiter.Allow("org-2", "tool-a"); err != nil {
		t.Fatalf("different org should have independent quota: %v", err)
	}
	if err := limiter.Allow("org-1", "tool-b"); err != nil {
		t.Fatalf("different tool should have independent quota: %v", err)
	}
	limiter.Reset()
	if err := limiter.Allow("org-1", "tool-a"); err != nil {
		t.Fatalf("quota should be available after reset: %v", err)
	}
}
