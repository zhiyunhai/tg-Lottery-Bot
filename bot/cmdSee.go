package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

func (b *Bot) cmdSee(msg *tgbotapi.Message) error {
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

	//加载用户参与过的所有活动信息
	b.userJoinEvents, err = GetUserEventsByUserID(db, msg.From.ID)
	if err != nil {
		return fmt.Errorf("GetUserEventsByUserID failed: %w", err)
	}

	if len(b.userJoinEvents) == 0 {
		err = b.sendReply(msg, "你尚未参与任何活动")
		if err != nil {
			return fmt.Errorf("sendReply failed: %w", err)
		}
		return nil
	}
	// 默认页码为1
	page := 1

	// 检查是否有页码参数
	args := msg.CommandArguments()
	if args != "" {
		// 尝试解析页码
		parsedPage, err := strconv.Atoi(args)
		if err == nil && parsedPage > 0 && parsedPage <= len(b.userJoinEvents) {
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
	b.sendPageCmdSee(msg.Chat.ID, 0, page) // 传递 messageID 为 0，表示新消息
	return nil
}

func (b *Bot) sendPageCmdSee(chatID int64, messageID int, page int) {
	totalPages := len(b.userJoinEvents)

	// 检查切片是否为空或页码是否超出范围
	if totalPages == 0 {
		log.Printf("No history events available")
		_, err := b.Bot.Send(tgbotapi.NewMessage(chatID, "你没有参与过活动"))
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

	// 获取当前页的活动信息
	info := b.userJoinEvents[page-1]

	outputMsg, err := createUserSeeEventInfoMsg(info)
	if err != nil {
		log.Printf("createUserSeeEventInfoMsg failed: %v", err)
	}
	outputMsg += "你参与过的活动\n"

	// 创建内联键盘
	keyboard := b.generateCmdSeeKeyboard(page, totalPages)

	if messageID == 0 {
		// 发送初始消息
		msg := tgbotapi.NewMessage(chatID, outputMsg)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = keyboard
		_, err := b.Bot.Send(msg)
		if err != nil {
			log.Printf("sendMessage failed: %v", err)
		}
	} else {
		// 编辑已有消息
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, outputMsg)
		editMsg.ParseMode = tgbotapi.ModeHTML
		editMsg.ReplyMarkup = &keyboard
		_, err := b.Bot.Send(editMsg)
		if err != nil {
			log.Printf("sendMessage failed: %v", err)
		}
	}
}

func (b *Bot) generateCmdSeeKeyboard(currentPage, totalPages int) tgbotapi.InlineKeyboardMarkup {
	prevPage := currentPage - 1
	nextPage := currentPage + 1

	var inlineKeyboard tgbotapi.InlineKeyboardMarkup

	if currentPage > 1 && currentPage < totalPages {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("上一页", "cmdSeePage"+strconv.Itoa(prevPage)),
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
				tgbotapi.NewInlineKeyboardButtonData("下一页", "cmdSeePage"+strconv.Itoa(nextPage)),
			),
		)
	} else if currentPage == 1 {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
				tgbotapi.NewInlineKeyboardButtonData("下一页", "cmdSeePage"+strconv.Itoa(nextPage)),
			),
		)
	} else if currentPage == totalPages {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("上一页", "cmdSeePage"+strconv.Itoa(prevPage)),
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
			),
		)
	}

	return inlineKeyboard
}
