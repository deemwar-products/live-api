package gemini

const DefaultMaxToolCalls = 3

type SafetyConfig struct {
	MaxToolCalls int
}

func (c SafetyConfig) EffectiveMaxToolCalls() int {
	if c.MaxToolCalls <= 0 {
		return DefaultMaxToolCalls
	}
	if c.MaxToolCalls > 5 {
		return 5
	}
	return c.MaxToolCalls
}
