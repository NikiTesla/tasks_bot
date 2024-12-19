package telegram

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

const (
	workersAmount = 100
)

type Bot struct {
	bot *tgbotapi.BotAPI

	storage repository.Storage

	logger *log.Entry
}

func NewBot(logger *log.Entry, storage repository.Storage) *Bot {
	botToken, ok := os.LookupEnv("TELEBOT_API")
	if !ok {
		log.Fatal("bot token is not presented")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.WithError(err).Fatal("can't create Bot API")
	}
	if os.Getenv("DEBUG") == "true" {
		bot.Debug = true
	}

	if err := createAdminChat(storage); err != nil {
		log.WithError(err).Warn("Failed to create admin chat. Entering no admin mode")
	}

	return &Bot{
		bot:     bot,
		storage: storage,
		logger:  log.WithField("type", "telegram-bot"),
	}
}

func createAdminChat(db repository.Storage) error {
	adminIDRaw, ok := os.LookupEnv("ADMIN_ID")
	if !ok {
		return fmt.Errorf("ADMIN_ID env is not found. Enabling no admin mode")
	}
	adminID, err := strconv.ParseInt(adminIDRaw, 10, 64)
	if err != nil {
		return fmt.Errorf("ADMIN_ID env %s is invalid. Enabling no admin mode, err: %w", adminIDRaw, err)
	}
	if err := db.AddChat(context.Background(), adminID, "admin", domain.Admin); err != nil {
		return fmt.Errorf("db.AddChat: %w", err)
	}
	return nil
}

func (b *Bot) Start(ctx context.Context) error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updateChannel, err := b.bot.GetUpdatesChan(updateConfig)
	if err != nil {
		return fmt.Errorf("cannot create update channel, err: %w", err)
	}

	log.Info("Bot is handling updates")
	b.handleUpdates(ctx, updateChannel)

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
				tasksCh <- func() { b.handleCommand(update.Message) }
			} else {
				tasksCh <- func() { b.handleMessage(update.Message) }
			}
		}
	}
}
