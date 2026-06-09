package contracts

import (
	"encoding/json"
	"testing"
)

func TestKnowledgeThresholdOrdering(t *testing.T) {
	if !(LowConfidenceThreshold < MediumConfidenceThreshold &&
		MediumConfidenceThreshold < HighConfidenceThreshold) {
		t.Fatalf("invalid threshold ordering: low=%v medium=%v high=%v",
			LowConfidenceThreshold, MediumConfidenceThreshold, HighConfidenceThreshold)
	}
}

func TestContractTypesPreserveTenantFields(t *testing.T) {
	input := EscalationRequest{
		SessionID: "session-1", OrgID: "org-1", CustomerID: "customer-1",
		ScoringReason: EscalationReasonKnowledgeGap,
	}
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var output EscalationRequest
	if err := json.Unmarshal(raw, &output); err != nil {
		t.Fatal(err)
	}
	if output.OrgID != input.OrgID || output.SessionID != input.SessionID ||
		output.ScoringReason != EscalationReasonKnowledgeGap {
		t.Fatalf("contract round trip changed fields: %#v", output)
	}
}

func TestNotificationConstantsAreDistinct(t *testing.T) {
	values := []string{ChannelSMS, ChannelWhatsApp, ChannelEmail, ChannelSlack, ChannelWebhook, ChannelInApp}
	seen := make(map[string]bool)
	for _, value := range values {
		if value == "" || seen[value] {
			t.Fatalf("invalid or duplicate channel %q", value)
		}
		seen[value] = true
	}
}
