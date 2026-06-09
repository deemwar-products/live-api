package contracts

import "context"

// ConnectorRouter executes provider-specific operations without exposing every
// connector operation directly to Gemini.
type ConnectorRouter interface {
	ExecuteConnector(
		ctx context.Context,
		orgID string,
		connector string,
		operation string,
		input map[string]any,
	) (map[string]any, error)
}
