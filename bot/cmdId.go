package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) cmdId(msg *tgbotapi.Message) error {
	userID := msg.From.ID
	message := fmt.Sprintf("*你的用户ID:* `%v`", userID)
	err := b.sendReplyMarkDown(msg, message)
	if err != nil {
		return fmt.Errorf("sendReplyMarkDown: %v", err)
	}
	return nil
}
