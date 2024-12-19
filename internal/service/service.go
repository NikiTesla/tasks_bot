package service

import (
	"context"
	"fmt"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/reconciler"
	"tasks_bot/internal/repository"
	"tasks_bot/internal/telegram"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	bot     *telegram.Bot
	rec     *reconciler.Reconciler
	storage repository.Storage

	logger *log.Entry
}

func New(logger *log.Entry, bot *telegram.Bot, rec *reconciler.Reconciler, storage repository.Storage) *Service {
	return &Service{
		bot:     bot,
		rec:     rec,
		storage: storage,
		logger:  logger.WithField("type", "service"),
	}
}

func (s *Service) Start(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return s.bot.Start(ctx)
	})

	eg.Go(func() error {
		return s.rec.Start(ctx, s.loop, time.Second)
	})

	eg.Go(func() error {
		<-ctx.Done()
		s.Close()
		return nil
	})

	return eg.Wait()
}

func (s *Service) Close() {
	s.logger.Info("service was closed")
	s.bot.Stop()
}

func (s *Service) loop(ctx context.Context) error {
	s.logger.Debug("loop function started")

	messages, err := s.storage.RetrieveMessages(ctx)
	if err != nil {
		return fmt.Errorf("n.db.RetrieveMessages: %w", err)
	}

	for _, message := range messages {
		if err = s.processMessage(ctx, message); err != nil {
			return fmt.Errorf("s.processMessage: %w", err)
		}
		if err := s.storage.SetHandledMessage(ctx, message.ID); err != nil {
			return fmt.Errorf("s.storage.SetHandledMessage: %w", err)
		}
	}

	return nil
}

func (s *Service) processMessage(_ context.Context, message domain.Message) error {
	s.logger.Infof("message %+v", message)

	return nil
}
