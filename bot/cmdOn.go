package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

func (b *Bot) cmdOn(msg *tgbotapi.Message) error {
	if !b.checkAdmin(msg) {
		err := b.sendReply(msg, "You are not an admin.")
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

	//加载未开奖和未取消的所有活动信息
	b.onEvent, err = loadNoCancelAndNoOpenEvents(db)
	if err != nil {
		return fmt.Errorf("loadNoCancelAndNoOpenEvents failed: %w", err)
	}

	if len(b.onEvent) == 0 {
		err = b.sendReply(msg, "没有正在进行的活动")
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
		if err == nil && parsedPage > 0 && parsedPage <= len(b.onEvent) {
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
	b.sendPageCmdOn(msg.Chat.ID, 0, page) // 传递 messageID 为 0，表示新消息
	return nil
}

func (b *Bot) sendPageCmdOn(chatID int64, messageID int, page int) {
	totalPages := len(b.onEvent)

	// 检查切片是否为空或页码是否超出范围
	if totalPages == 0 {
		log.Printf("No history events available")
		_, err := b.Bot.Send(tgbotapi.NewMessage(chatID, "没有可显示的正在进行活动。"))
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
	info := b.onEvent[page-1]

	outputMsg, err := createAllEventInfoMsg(info)
	if err != nil {
		log.Printf("createAllEventInfoMsg err: %v", err)
		return
	}

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

	partnerList, err := getParticipantsByEventID(db, info.ID)
	if err != nil {
		log.Printf("getParticipantsByEventID err: %v", err)
		return
	}

	partnerString := "<b>当前参与者信息:</b>\n"
	for _, partner := range partnerList {
		partnerString += fmt.Sprintf("<b>用户ID: </b>%v | <b>用户名: </b>%v\n", partner.UserID, partner.UserName)
	}

	outputMsg += partnerString + "<b>正在进行的活动</b>\n"

	// 创建内联键盘
	keyboard := b.generateCmdOnKeyboard(page, totalPages)

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

func (b *Bot) generateCmdOnKeyboard(currentPage, totalPages int) tgbotapi.InlineKeyboardMarkup {
	prevPage := currentPage - 1
	nextPage := currentPage + 1

	var inlineKeyboard tgbotapi.InlineKeyboardMarkup

	if currentPage > 1 && currentPage < totalPages {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("上一页", "cmdOnPage"+strconv.Itoa(prevPage)),
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
				tgbotapi.NewInlineKeyboardButtonData("下一页", "cmdOnPage"+strconv.Itoa(nextPage)),
			),
		)
	} else if currentPage == 1 {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
				tgbotapi.NewInlineKeyboardButtonData("下一页", "cmdOnPage"+strconv.Itoa(nextPage)),
			),
		)
	} else if currentPage == totalPages {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("上一页", "cmdOnPage"+strconv.Itoa(prevPage)),
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
			),
		)
	}

	return inlineKeyboard
}
