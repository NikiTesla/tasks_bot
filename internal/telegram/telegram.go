package telegram

import (
	"context"
	"fmt"
	"sync"
	"tasks_bot/internal/config"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	workersAmount = 100
)

var (
	observerPasswordHash []byte
	executorPasswordHash []byte
	chiefPasswordHash    []byte
	adminPasswordHash    []byte
)

type Bot struct {
	bot *tgbotapi.BotAPI

	storage repository.Storage
	cfg     *config.TelegramConfig

	logger *log.Entry
}

func NewBot(logger *log.Entry, storage repository.Storage, cfg *config.TelegramConfig) *Bot {
	bot, err := tgbotapi.NewBotAPI(cfg.APIToken)
	if err != nil {
		log.WithError(err).Fatal("can't create Bot API")
	}
	bot.Debug = cfg.Debug

	if err := createAdminChat(storage, cfg); err != nil {
		log.WithError(err).Warn("Failed to create admin chat. Entering no admin mode")
	}
	if err := setPasswords(cfg); err != nil {
		log.WithError(err).Error("failed to set passwords")
	}

	return &Bot{
		bot:     bot,
		storage: storage,
		logger:  log.WithField("type", "telegram-bot"),
	}
}

func createAdminChat(db repository.Storage, cfg *config.TelegramConfig) error {
	if err := db.AddChat(context.Background(), cfg.AdminID, cfg.AdminUsername, "-", domain.Admin); err != nil {
		return fmt.Errorf("db.AddChat: %w", err)
	}
	return nil
}

func setPasswords(cfg *config.TelegramConfig) (err error) {
	observerPasswordHash, err = bcrypt.GenerateFromPassword([]byte(cfg.ObserverPasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("generate observer password hash, err: %w", err)
	}
	chiefPasswordHash, err = bcrypt.GenerateFromPassword([]byte(cfg.ChiefPasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("generate chief password hash, err: %w", err)
	}
	executorPasswordHash, err = bcrypt.GenerateFromPassword([]byte(cfg.ExecutorPasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("generate executor password hash, err: %w", err)
	}
	adminPasswordHash, err = bcrypt.GenerateFromPassword([]byte(cfg.AdminPasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("generate admin password hash, err: %w", err)
	}
	return nil
}

func (b *Bot) Start(ctx context.Context) error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	log.Info("Bot is handling updates")
	b.handleUpdates(ctx, b.bot.GetUpdatesChan(updateConfig))

	return nil
}

func (b *Bot) Stop() {
	log.Info("stopping bot")
	b.bot.StopReceivingUpdates()
}

func (b *Bot) handleUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	tasksCh := make(chan func(), workersAmount)
	wg := &sync.WaitGroup{}
	for range workersAmount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasksCh {
				task()
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("closing tasks channel")

			close(tasksCh)
			wg.Wait()
			return

		case update := <-updates:
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() {
				tasksCh <- func() { b.handleCommand(ctx, update.Message) }
			} else {
				tasksCh <- func() { b.handleMessage(ctx, update.Message) }
			}
		}
	}
}
