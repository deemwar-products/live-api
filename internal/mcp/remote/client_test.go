package remote

import (
	"context"
	"testing"
)

func TestUnconnectedClientFailsSafely(t *testing.T) {
	var nilClient *Client
	if err := nilClient.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := nilClient.DiscoverTools(context.Background()); err == nil {
		t.Fatal("expected unconnected discovery error")
	}
	if _, err := (&Client{}).ExecuteTool(context.Background(), "echo", nil); err == nil {
		t.Fatal("expected unconnected execution error")
	}
}

func TestConnectCommandReportsStartupFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client, err := ConnectCommand(ctx, "command-that-does-not-exist-live-api")
	if err == nil {
		_ = client.Close()
		t.Fatal("expected command startup failure")
	}
}
