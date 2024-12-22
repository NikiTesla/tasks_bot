package main

import (
	"context"
	"os"
	"os/signal"
	"path"
	"syscall"
	"tasks_bot/internal/config"
	"tasks_bot/internal/reconciler"
	"tasks_bot/internal/repository"
	"tasks_bot/internal/service"
	"tasks_bot/internal/telegram"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := godotenv.Load(path.Join("./", ".env")); err != nil {
		log.WithError(err).Warn("failed to load .env")
	}
	cfg, err := config.Load()
	if err != nil {
		log.WithError(err).Fatal("failed to load config")
	}
	logger := createLogger(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	setSigintHandler(logger, cancel)

	storage, err := repository.NewStorage(ctx, cfg)
	if err != nil {
		logger.Fatalf("failed to create storage, err: %s", err)
	}

	service := service.New(
		logger,
		telegram.NewBot(logger, storage, cfg.TelegramConfig),
		reconciler.New(logger),
		storage,
	)

	if err := service.Start(ctx); err != nil {
		log.WithError(err).Fatal("service failed")
	}
}

func createLogger(cfg *config.Config) *log.Entry {
	logger := log.NewEntry(log.StandardLogger())
	if cfg.Debug {
		logger.Logger.SetLevel(log.DebugLevel)
	}
	return logger
}

func setSigintHandler(logger *log.Entry, cancelFunc context.CancelFunc) {
	done := make(chan os.Signal, 1)

	signal.Notify(done, syscall.SIGINT)
	signal.Notify(done, syscall.SIGTERM)

	go func() {
		s := <-done
		logger.Infof("received signal: %s", s)
		cancelFunc()
	}()
}

func getDBFile() string {
	dbFile, ok := os.LookupEnv("DB_FILENAME")
	if !ok {
		return "db.sql"
	}
	return dbFile
}
