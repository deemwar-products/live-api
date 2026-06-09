package analytics

import "testing"

func TestScoreConversationBands(t *testing.T) {
	tests := []struct {
		name    string
		signals map[string]float64
		reason  string
	}{
		{name: "healthy", signals: map[string]float64{}, reason: "healthy"},
		{name: "monitor", signals: map[string]float64{"frustration": 1, "knowledge_gap": 0.5}, reason: "monitor"},
		{name: "escalate", signals: map[string]float64{"frustration": 1, "knowledge_gap": 1, "tool_failure": 1}, reason: "escalation_recommended"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ScoreConversation(tt.signals); got.Reason != tt.reason {
				t.Fatalf("expected %q, got %#v", tt.reason, got)
			}
		})
	}
}

func TestScoreConversationClampsAtZero(t *testing.T) {
	got := ScoreConversation(map[string]float64{
		"frustration": 10, "knowledge_gap": 10, "tool_failure": 10,
	})
	if got.Score != 0 {
		t.Fatalf("expected zero score, got %v", got.Score)
	}
}
