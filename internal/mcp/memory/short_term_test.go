package memory

import (
	"context"
	"testing"
)

func TestInMemoryStoreAppendsAndReturnsRecentMessages(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	for _, content := range []string{"one", "two", "three"} {
		if err := store.Append(ctx, "session-1", Message{Role: "user", Content: content}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := store.List(ctx, "session-1", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Content != "two" || got[1].Content != "three" {
		t.Fatalf("unexpected messages: %#v", got)
	}
	got[0].Content = "mutated"
	again, _ := store.List(ctx, "session-1", 2)
	if again[0].Content != "two" {
		t.Fatal("List returned internal storage")
	}
}

func TestInMemoryStoreValidatesInputAndContext(t *testing.T) {
	store := NewInMemoryStore()
	if err := store.Append(context.Background(), "", Message{}); err == nil {
		t.Fatal("expected session validation error")
	}
	if err := store.Append(context.Background(), "session", Message{}); err == nil {
		t.Fatal("expected message validation error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.List(ctx, "session", 0); err == nil {
		t.Fatal("expected cancellation error")
	}
}
