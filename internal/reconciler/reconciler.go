package reconciler

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type LoopFunc = func(context.Context) error

// Reconciler performs some action (LoopFunc) with constant timeout
type Reconciler struct {
	log *log.Entry
}

func New(logger *log.Entry) *Reconciler {
	return &Reconciler{
		log: logger,
	}
}

func (r *Reconciler) Start(ctx context.Context, loop LoopFunc, tickTime time.Duration) error {
	ticker := time.NewTicker(tickTime)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("ctx.Done: %w", ctx.Err())
		case <-ticker.C:
			if err := loop(ctx); err != nil {
				r.log.WithError(err).Error("loop failed")
			}
		}
	}
}
