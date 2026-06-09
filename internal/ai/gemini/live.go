package gemini

import "context"

type LiveSessionConfig struct {
	Model string
}

func (s *Service) LiveReady(_ context.Context) bool {
	return s != nil && s.Client != nil
}
