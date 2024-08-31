package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

func (b *Bot) cmdList(msg *tgbotapi.Message) (err error) {
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

	b.prizeList, err = loadPrizes()
	if err != nil {
		log.Printf("loadPrizes err: %s", err)
		err = b.sendReply(msg, "加载奖品失败！")
		if err != nil {
			return err
		}
		return nil
	}

	if len(b.prizeList) == 0 {
		err = b.sendReply(msg, "没有奖品可显示")
		if err != nil {
			return err
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
		if err == nil && parsedPage > 0 && parsedPage <= len(b.prizeList) {
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
	b.sendPageCmdList(msg.Chat.ID, 0, page) // 传递 messageID 为 0，表示新消息
	return nil
}

func (b *Bot) sendPageCmdList(chatID int64, messageID int, page int) {
	totalPrizes := len(b.prizeList)
	totalPages := (totalPrizes + 9) / 10 // 计算总页数，每页10个奖品

	// 检查切片是否为空或页码是否超出范围
	if totalPages == 0 {
		log.Printf("No prizes available")
		_, err := b.Bot.Send(tgbotapi.NewMessage(chatID, "没有可显示的奖品。"))
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

	// 计算当前页的奖品范围
	startIndex := (page - 1) * 10
	endIndex := startIndex + 10
	if endIndex > totalPrizes {
		endIndex = totalPrizes
	}

	// 生成奖品列表字符串
	outputMsg := fmt.Sprintf("<b>😊 加载成功</b>  共 <b>%d</b> 个奖品\n", len(b.prizeList))
	for i := startIndex; i < endIndex; i++ {
		// 使用 <li> 标签生成列表项，并将奖品名称进行转义
		outputMsg += fmt.Sprintf("%d. %s\n", i+1, tgbotapi.EscapeText(tgbotapi.ModeHTML, b.prizeList[i]))
	}

	keyBoard := b.generateCmdListKeyboard(page, totalPages)
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

func (b *Bot) generateCmdListKeyboard(currentPage, totalPages int) tgbotapi.InlineKeyboardMarkup {
	prevPage := currentPage - 1
	nextPage := currentPage + 1

	var inlineKeyboard tgbotapi.InlineKeyboardMarkup

	if currentPage > 1 && currentPage < totalPages {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("上一页", "cmdListPage"+strconv.Itoa(prevPage)),
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
				tgbotapi.NewInlineKeyboardButtonData("下一页", "cmdListPage"+strconv.Itoa(nextPage)),
			),
		)
	} else if currentPage == 1 {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
				tgbotapi.NewInlineKeyboardButtonData("下一页", "cmdListPage"+strconv.Itoa(nextPage)),
			),
		)
	} else if currentPage == totalPages {
		inlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("上一页", "cmdListPage"+strconv.Itoa(prevPage)),
				tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(currentPage)+"/"+strconv.Itoa(totalPages), "noop"),
			),
		)
	}

	return inlineKeyboard
}
