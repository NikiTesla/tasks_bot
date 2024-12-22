package service

import (
	"context"
	"fmt"
	"tasks_bot/internal/reconciler"
	"tasks_bot/internal/repository"
	"tasks_bot/internal/telegram"
	"time"

	_ "net/http/pprof"

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
	// Start pprof server
	go func() {

	}()

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

	// eg.Go(func() error {
	// 	s.logger.Info("Starting pprof server on :6060")
	// 	if err := http.ListenAndServe(":6060", nil); err != nil {
	// 		return fmt.Errorf("pprof server failed: %v", err)
	// 	}
	// 	return nil
	// })

	return eg.Wait()
}

func (s *Service) Close() {
	s.logger.Info("service was closed")
	s.bot.Stop()
}

func (s *Service) loop(ctx context.Context) error {
	if err := s.processMessages(ctx); err != nil {
		return fmt.Errorf("s.processMessages: %s", err)
	}
	if err := s.processExpiredTasks(ctx); err != nil {
		return fmt.Errorf("s.processExpiredTasks: %w", err)
	}
	return nil
}

func (s *Service) processMessages(ctx context.Context) error {
	messages, err := s.storage.RetrieveMessages(ctx)
	if err != nil {
		return fmt.Errorf("n.db.RetrieveMessages: %w", err)
	}

	for _, message := range messages {
		s.logger.Infof("message processed: %+v", message)

		if err := s.storage.SetHandledMessage(ctx, message.ID); err != nil {
			return fmt.Errorf("s.storage.SetHandledMessage: %w", err)
		}
	}
	return nil
}

func (s *Service) processExpiredTasks(ctx context.Context) error {
	tasks, err := s.storage.GetExpiredTasksToMark(ctx)
	if err != nil {
		return fmt.Errorf("s.storage.GetExpiredTasks: %w", err)
	}
	for _, task := range tasks {
		if err := s.bot.NotifyObservers(ctx, task); err != nil {
			return fmt.Errorf("s.bot.NotifyObservers: %w", err)
		}
	}
	return nil
}
