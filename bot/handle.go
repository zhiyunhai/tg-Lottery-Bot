package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

func (b *Bot) handleUpdate(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		err := b.cmdStart(msg)
		if err != nil {
			log.Printf("cmdStart failed: %v", err)
		}
	case "id":
		err := b.cmdId(msg)
		if err != nil {
			log.Printf("cmdId failed: %v", err)
		}
	case "add":
		err := b.cmdAdd(msg)
		if err != nil {
			log.Printf("cmdAdd failed: %v", err)
		}
	case "delete":
		err := b.cmdDelete(msg)
		if err != nil {
			log.Printf("cmdDelete failed: %v", err)
		}
	case "list":
		err := b.cmdList(msg)
		if err != nil {
			log.Printf("cmdPrizeList failed: %v", err)
		}
	case "on":
		err := b.cmdOn(msg)
		if err != nil {
			log.Printf("cmdOn failed: %v", err)
		}
	case "cancel":
		err := b.cmdCancel(msg)
		if err != nil {
			log.Printf("cmdCancel failed: %v", err)
		}
	case "history":
		err := b.cmdHistory(msg)
		if err != nil {
			log.Printf("cmdHistory failed: %v", err)
		}
	case "create":
		err := b.cmdCreate(msg)
		if err != nil {
			log.Printf("cmdCreate failed: %v", err)
		}
	case "open":
		err := b.cmdOpen(msg)
		if err != nil {
			log.Printf("cmdOpen failed: %v", err)
		}
	case "close":
		err := b.cmdClose(msg)
		if err != nil {
			log.Printf("cmdClose failed: %v", err)
		}
	case "see":
		err := b.cmdSee(msg)
		if err != nil {
			log.Printf("cmdSee failed: %v", err)
		}
	case "join":
		err := b.cmdJoin(msg)
		if err != nil {
			log.Printf("cmdJoin failed: %v", err)
		}
	case "prize":
		err := b.cmdPrize(msg)
		if err != nil {
			log.Printf("cmdPrize failed: %v", err)
		}

	}
}

func (b *Bot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	data := callbackQuery.Data
	userID := callbackQuery.From.ID

	switch {
	case len(data) == 0:
		log.Printf("callback query is empty")

	case data[:4] == "noop":
		log.Println("Ignore operation noop")

	case len(data) >= 9 && data[:9] == "cmdOnPage":
		page, err := strconv.Atoi(data[9:])
		if err != nil {
			log.Printf("Invalid page number: %v", err)
			return
		}
		b.sendPageCmdOn(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, page)

	case len(data) >= 13 && data[:13] == "cmdCancelPage":
		page, err := strconv.Atoi(data[13:])
		if err != nil {
			log.Printf("Invalid page number: %v", err)
			return
		}
		b.sendPageCmdCancel(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, page)

	case len(data) >= 14 && data[:14] == "cmdHistoryPage":
		page, err := strconv.Atoi(data[14:])
		if err != nil {
			log.Printf("Invalid page number: %v", err)
			return
		}
		b.sendPageCmdHistory(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, page)

	case len(data) >= 10 && data[:10] == "cmdSeePage":
		page, err := strconv.Atoi(data[10:])
		if err != nil {
			log.Printf("Invalid page number: %v", err)
			return
		}
		b.sendPageCmdSee(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, page)

	case len(data) >= 12 && data[:12] == "cmdPrizePage":
		page, err := strconv.Atoi(data[12:])
		if err != nil {
			log.Printf("Invalid page number: %v", err)
		}
		b.sendPageCmdPrize(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, page)

	case len(data) >= 11 && data[:11] == "cmdListPage":
		page, err := strconv.Atoi(data[11:])
		if err != nil {
			log.Printf("Invalid page number: %v", err)
		}
		b.sendPageCmdList(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, page)

	case data == "confirm_create_event":
		if len(b.EventInfoMap) == 0 {
			err := b.sendReply(callbackQuery.Message, "无效，请重新创建")
			if err != nil {
				log.Printf("sendReply failed: %v", err)
			}
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

		eventInfo := b.EventInfoMap[callbackQuery.Message.Chat.ID]
		if eventInfo.PrizeResultMethod == "1" {
			err := CheckTime(eventInfo.TimeOfWinners)
			if err != nil {
				err = b.sendReply(callbackQuery.Message, err.Error())
				if err != nil {
					log.Printf("Error sending reply: %v", err)
				}
				return
			}
		}
		// 保存活动到数据库
		err = saveEventsInformation(db, eventInfo)
		if err != nil {
			log.Printf("save CreateInformation to Database ERROR: %v", err)
			err = b.sendReply(callbackQuery.Message, err.Error())
			if err != nil {
				log.Printf("Error sending reply: %v", err)
			}
			return
		}

		// 发布抽奖活动到群组
		sentGroupMsg := fmt.Sprintf(
			"🎉 <b>新的抽奖活动发布啦</b> 🎁\n"+
				"<b>抽奖群：</b> %s\n"+
				"<b>奖品名称：</b> %s\n"+
				"<b>奖品数量：</b> %d\n"+
				"<b>开奖方式：</b> %s\n"+
				"<b>参与方式：</b> %s\n",
			tgbotapi.EscapeText(tgbotapi.ModeHTML, eventInfo.GroupName),
			tgbotapi.EscapeText(tgbotapi.ModeHTML, eventInfo.PrizeName),
			eventInfo.PrizeCount,
			tgbotapi.EscapeText(tgbotapi.ModeHTML, eventInfo.PrizeResult),
			tgbotapi.EscapeText(tgbotapi.ModeHTML, eventInfo.Participate),
		)

		if eventInfo.HowToParticipate == "1" {
			sentGroupMsg += fmt.Sprintf("<b>关键词：</b> <code>%s</code>\n<b>参与抽奖指令：</b> <code>/join %v</code>\n",
				tgbotapi.EscapeText(tgbotapi.ModeHTML, eventInfo.KeyWord),
				tgbotapi.EscapeText(tgbotapi.ModeHTML, eventInfo.KeyWord),
			)
		}

		if eventInfo.PrizeResultMethod == "1" {
			sentGroupMsg += fmt.Sprintf("<b>开奖时间：</b> <code>%s</code> %v\n",
				tgbotapi.EscapeText(tgbotapi.ModeHTML, eventInfo.TimeOfWinners),
				config.TimeZone,
			)
		} else if eventInfo.PrizeResultMethod == "2" {
			sentGroupMsg += fmt.Sprintf("<b>开奖人数：</b> %d\n", eventInfo.NumberOfWinners)
		}

		if eventInfo.HowToParticipate == "2" {
			sentGroupMsg += "<b>参与抽奖指令：</b> <code>/join</code>\n"
		}

		err = b.sendMsgToGroup(sentGroupMsg)
		if err != nil {
			log.Printf("Error sending msg to group: %v", err)
			return
		}

		err = b.sendReply(callbackQuery.Message, "抽奖活动已发布！")
		if err != nil {
			log.Printf("Error sending reply: %v", err)
			return
		}

		// 去除奖品文件中已经选择的奖品
		err = ReplaceUnChoosePrizes(eventInfo.AllPrizes, eventInfo.ChoosePrizes)
		if err != nil {
			log.Printf("ReplaceUnChoosePrizes err %v\n", err)
			return
		}
		// 刷新新的活动开奖的时间定时
		err = b.regularPrizeDraw()
		if err != nil {
			log.Printf("regularPrizeDraw err %v\n", err)
			return
		}
		// 清理用户状态
		b.userStatesMu.Lock()
		delete(b.UserStates, userID)
		b.userStatesMu.Unlock()

		b.eventInfoMapMu.Lock()
		delete(b.EventInfoMap, userID)
		b.eventInfoMapMu.Unlock()
	case data == "cancel_create_event":
		err := b.sendReply(callbackQuery.Message, "抽奖活动创建已取消。")
		if err != nil {
			log.Printf("Error sending reply: %v", err)
		}
		// 清理用户状态
		b.userStatesMu.Lock()
		delete(b.UserStates, userID)
		b.userStatesMu.Unlock()

		b.eventInfoMapMu.Lock()
		delete(b.EventInfoMap, userID)
		b.eventInfoMapMu.Unlock()
	default:
		log.Printf("Invalid callback query: %v", data)
	}
	// 发送回调响应以告知 Telegram 操作已处理
	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := b.Bot.Request(callback); err != nil {
		log.Printf("Error sending callback: %v", err)
	}
}
