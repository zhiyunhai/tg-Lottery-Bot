package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

func (b *Bot) cmdCancel(msg *tgbotapi.Message) error {
	if !b.checkAdmin(msg) {
		err := b.sendReply(msg, "You are not an admin")
		if err != nil {
			return fmt.Errorf("sendReply failed: %w", err)
		}
		return nil
	}

	if !msg.Chat.IsPrivate() {
		return b.sendReply(msg, "请在私聊中使用管理员指令")
	}

	// 初始化数据库
	db, err := initDB()
	if err != nil {
		return fmt.Errorf("initDB failed: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close db err: %v", err)
		}
	}()

	//加载取消的活动
	b.cancelEvents, err = loadCancelEvents(db)
	if err != nil {
		return fmt.Errorf("loadCancelEvents failed: %w", err)
	}

	if len(b.cancelEvents) == 0 {
		err = b.sendReply(msg, "没有取消的活动")
		if err != nil {
			return fmt.Errorf("sendReply failed: %w", err)
		}
		return nil
	}

	page := 1
	// 检查是否有页码参数
	args := msg.CommandArguments()
	if args != "" {
		// 尝试解析页码
		parsedPage, err := strconv.Atoi(args)
		if err == nil && parsedPage > 0 && parsedPage <= len(b.cancelEvents) {
			page = parsedPage
		} else {
			// 如果解析失败，返回一个错误提示
			err = b.sendReply(msg, "无效的页码，非正整数或超出范围")
			if err != nil {
				return fmt.Errorf("sendReply failed: %w", err)
			}
			return nil
		}
	}
	// 发送指定页码的消息
	b.sendPageCmdCancel(msg.Chat.ID, 0, page) // 传递 messageID 为 0，表示新消息
	return nil
}

func (b *Bot) sendPageCmdCancel(chatID int64, messageID int, page int) {
	totalPages := len(b.cancelEvents)

	// 检查切片是否为空或页码是否超出范围
	if totalPages == 0 {
		log.Printf("No history events available")
		_, err := b.Bot.Send(tgbotapi.NewMessage(chatID, "没有可显示的活动。"))
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}

	if page > totalPages || page < 1 {
		log.Printf("Invalid page number: %d, total pages: %d", page, totalPages)
		_, err := b.Bot.Send(tgbotapi.NewMessage(chatID, "页码无效。"))
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return
	}

	info := b.cancelEvents[page-1]

	outputMsg, err := createAllEventInfoMsg(info)
	if err != nil {
		log.Printf("createAllEventInfoMsg failed: %v", err)
		return
	}
	outputMsg += "已取消的活动\n"

	keyBoard := b.generateCmdCancelKeyboard(page, totalPages)
	if messageID == 0 {
		msg := tgbotapi.NewMessage(chatID, outputMsg)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = keyBoard
		_, err := b.Bot.Send(msg)
		if err != nil {
			log.Printf("sendMessage failed: %v", err)
		}
	} else {
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, outputMsg)
		editMsg.ParseMode = tgbotapi.ModeHTML
		editMsg.ReplyMarkup = &keyBoard
		_, err := b.Bot.Send(editMsg)
		if err != nil {
			log.Printf("sendMessage failed: %v", err)
		}
	}
}

func (b *Bot) generateCmdCancelKeyboard(currentPage, totalPages int) tgbotapi.InlineKeyboardMarkup {
	prevPage := currentPage - 1
	nextPage := currentPage + 1

	var inlineKeyboard tgbotapi.InlineKeyboardMarkup

	if currentPage > 1 && currentPage < totalPages {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("上一页", "cmdCancelPage"+strconv.Itoa(prevPage)),
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
				tgbotapi.NewInlineKeyboardButtonData("下一页", "cmdCancelPage"+strconv.Itoa(nextPage)),
			),
		)
	} else if currentPage == 1 {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
				tgbotapi.NewInlineKeyboardButtonData("下一页", "cmdCancelPage"+strconv.Itoa(nextPage)),
			),
		)
	} else if currentPage == totalPages {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("上一页", "cmdCancelPage"+strconv.Itoa(prevPage)),
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
			),
		)
	}

	return inlineKeyboard
}
