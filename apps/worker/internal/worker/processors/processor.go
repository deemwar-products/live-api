package processors

import (
	"context"

	"github.com/deemwar/live-api/apps/worker/internal/queue"
)

// Processor handles a single job message.
// Implementations must be safe for concurrent use if the worker runs
// multiple goroutines pointing at the same processor.
type Processor interface {
	Process(ctx context.Context, msg queue.Message) error
}
