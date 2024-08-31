package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func (b *Bot) cmdAdd(msg *tgbotapi.Message) error {
	if !b.checkAdmin(msg) {
		err := b.sendReply(msg, "You are not an admin.")
		if err != nil {
			return err
		}
		return nil
	}

	if !msg.Chat.IsPrivate() {
		return b.sendReply(msg, "请在私聊中使用管理员指令")
	}

	args := strings.TrimSpace(msg.CommandArguments())
	if args == "" {
		err := b.sendReply(msg, "/add [需要添加的奖品，每个奖品用英文`分割]")
		if err != nil {
			return err
		}
		return nil
	}

	prizes := strings.Split(args, "`")
	for i := range prizes {
		prizes[i] = strings.TrimSpace(prizes[i])
	}

	var validPrizes []string
	for _, prize := range prizes {
		prize = strings.TrimSpace(prize)
		if prize != "" {
			validPrizes = append(validPrizes, prize)
		}
	}

	if len(validPrizes) == 0 {
		err := b.sendReply(msg, "没有有效的奖品需要添加")
		if err != nil {
			return err
		}
		return nil
	}

	// 添加奖品到文件
	err := addPrizesToPrizeTxtFile(validPrizes)
	if err != nil {
		log.Printf("addPrizesToPrizeTxtFile error: %v", err)
		err = b.sendReply(msg, "添加奖品失败")
		if err != nil {
			return err
		}
		return err
	}

	response := fmt.Sprintf("<b>添加奖品成功！共 %d 个</b>\n添加的奖品：\n", len(validPrizes))
	for _, prize := range validPrizes {
		response += fmt.Sprintf("%s,", tgbotapi.EscapeText(tgbotapi.ModeHTML, prize))
	}

	err = b.sendReplyHTML(msg, response)
	if err != nil {
		return err
	}
	return nil
}
