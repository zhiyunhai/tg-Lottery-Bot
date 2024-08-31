package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

func (b *Bot) cmdPrize(msg *tgbotapi.Message) error {
	if !msg.Chat.IsPrivate() {
		return b.sendReply(msg, "请在私聊中发送")
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

	b.winInfoList, err = getWinInfoByUserID(db, msg.From.ID)
	if err != nil {
		return fmt.Errorf("getWinInfoByUserID failed: %w", err)
	}

	if len(b.winInfoList) == 0 {
		err = b.sendReply(msg, "没有中奖记录")
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
		if err == nil && parsedPage > 0 && parsedPage <= len(b.winInfoList) {
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
	b.sendPageCmdPrize(msg.Chat.ID, 0, page) // 传递 messageID 为 0，表示新消息
	return nil
}

func (b *Bot) sendPageCmdPrize(chatID int64, messageID int, page int) {
	// 初始化数据库
	db, err := initDB()
	if err != nil {
		log.Printf("initDB failed: %v", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close db err: %v", err)
		}
	}()

	totalPages := len(b.winInfoList)

	// 检查切片是否为空或页码是否超出范围
	if totalPages == 0 {
		log.Printf("No history events available")
		_, err := b.Bot.Send(tgbotapi.NewMessage(chatID, "没有中奖记录"))
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
	info := b.winInfoList[page-1]

	NumberOfParticipants, err := getParticipantCountByEventID(db, info.ID)
	if err != nil {
		log.Printf("getParticipantCountByEventID failed: %v", err)
	}

	outputMsg := fmt.Sprintf("🎉*你中奖的活动:*🎉\n\n*🎟️ 活动 ID:* `%s`\n*🏷️ 活动名称:* %s\n*🎁 奖品数量:* %d\n",
		info.ID, info.PrizeName, info.PrizeCount)
	if info.PrizeResultMethod == "1" { // 按时间开奖
		outputMsg += fmt.Sprintf("*⏰ 开奖时间:* %s %s\n*👥 参与人数:* %d\n", info.TimeOfWinners, config.TimeZone, NumberOfParticipants)
	} else if info.PrizeResultMethod == "2" { // 按人数开奖
		NumberOfParticipants, err = getParticipantCountByEventID(db, info.ID)
		if err != nil {
			log.Printf(err.Error())
		}
		outputMsg += fmt.Sprintf("*🏆 开奖人数:* %d\n*👥 参与人数:* %d\n", info.NumberOfWinners, NumberOfParticipants)

	}
	outputMsg += fmt.Sprintf("*🎁奖品：* %v", info.PrizeInfo)

	// 创建内联键盘
	keyboard := b.generateCmdPrizeKeyboard(page, totalPages)

	if messageID == 0 {
		// 发送初始消息
		msg := tgbotapi.NewMessage(chatID, outputMsg)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = keyboard
		_, err := b.Bot.Send(msg)
		if err != nil {
			log.Printf("sendMessage failed: %v", err)
		}
	} else {
		// 编辑已有消息
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, outputMsg)
		editMsg.ParseMode = tgbotapi.ModeMarkdown
		editMsg.ReplyMarkup = &keyboard
		_, err := b.Bot.Send(editMsg)
		if err != nil {
			log.Printf("sendMessage failed: %v", err)
		}
	}
}

func (b *Bot) generateCmdPrizeKeyboard(currentPage, totalPages int) tgbotapi.InlineKeyboardMarkup {
	prevPage := currentPage - 1
	nextPage := currentPage + 1

	var inlineKeyboard tgbotapi.InlineKeyboardMarkup

	if currentPage > 1 && currentPage < totalPages {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("上一页", "cmdPrizePage"+strconv.Itoa(prevPage)),
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
				tgbotapi.NewInlineKeyboardButtonData("下一页", "cmdPrizePage"+strconv.Itoa(nextPage)),
			),
		)
	} else if currentPage == 1 {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
				tgbotapi.NewInlineKeyboardButtonData("下一页", "cmdPrizePage"+strconv.Itoa(nextPage)),
			),
		)
	} else if currentPage == totalPages {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("上一页", "cmdPrizePage"+strconv.Itoa(prevPage)),
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
			),
		)
	}

	return inlineKeyboard
}
