package analytics

type ConversationScore struct {
	Score  float64
	Reason string
}

func ScoreConversation(signals map[string]float64) ConversationScore {
	score := 1.0
	score -= signals["frustration"] * 0.35
	score -= signals["knowledge_gap"] * 0.30
	score -= signals["tool_failure"] * 0.25
	if score < 0 {
		score = 0
	}
	return ConversationScore{Score: score, Reason: reasonFor(score)}
}

func reasonFor(score float64) string {
	if score < 0.3 {
		return "escalation_recommended"
	}
	if score < 0.6 {
		return "monitor"
	}
	return "healthy"
}
