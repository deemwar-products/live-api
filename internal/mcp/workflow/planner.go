package workflow

func SupportEscalationPlan(summary, reason string) []Step {
	return []Step{
		{ToolName: "create_ticket", Args: map[string]any{
			"title":       "Human follow-up requested",
			"description": summary,
			"priority":    "high",
		}},
		{ToolName: "create_escalation", Args: map[string]any{
			"reason":   reason,
			"summary":  summary,
			"priority": "high",
		}},
	}
}
