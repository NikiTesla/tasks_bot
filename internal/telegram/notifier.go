package telegram

import (
	"context"
	"fmt"
	"slices"
	"tasks_bot/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) NotifyObservers(ctx context.Context, task domain.Task, excludeChatIDs ...int64) error {
	observers, err := b.storage.GetObservers(ctx)
	if err != nil {
		return fmt.Errorf("b.storage.GetObservers: %w", err)
	}

	for chatID := range observers {
		if slices.Contains(excludeChatIDs, chatID) {
			continue
		}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("UPD: \n\n%s", task.String()))
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = b.bot.Send(msg); err != nil {
			return fmt.Errorf("b.bot.Send (%d): %w", msg.ChatID, err)
		}
	}

	return nil
}
