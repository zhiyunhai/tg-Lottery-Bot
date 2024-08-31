package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"time"
)

func (b *Bot) cmdCreate(msg *tgbotapi.Message) (err error) {
	var eventInfo EventInformation

	if !b.checkAdmin(msg) {
		err := b.sendReply(msg, "you are not an admin")
		if err != nil {
			log.Printf("Error sending reply: %v", err)
		}
		return nil
	}

	if !msg.Chat.IsPrivate() {
		return b.sendReply(msg, "请在私聊中使用管理员指令")
	}

	eventInfo.AllPrizes, err = loadPrizes()
	if err != nil {
		err = b.sendReply(msg, "Error loading prizes")
		if err != nil {
			log.Printf("Error sending reply: %v", err)
		}
		return fmt.Errorf("error loading prizes: %v", err)
	}

	args := strings.Split(msg.CommandArguments(), " ")
	if len(args) < 6 {
		err = b.sendReplyMarkDown(msg, "*开奖方法：*\n1.按时间开奖\n2.按人数开奖\n\n"+
			"*参与方法：*\n1.群组内发送关键词\n2.私聊机器人参与\n\n"+
			"*传递说明：*\n"+
			"`/create [活动名称] [奖品数量] [开奖方法1/2] [选1填时间，选2填人数] [参与方法1/2] [选1填关键词，选2填 私聊机器人参与]`\n\n"+
			"*示例：*\n"+
			"`/create 我要抽奖 10 1 20240823-23:07 1 抽奖`\n"+
			"`/create 我要抽奖 10 1 20240823-23:07 2 私聊机器人参与`\n"+
			"`/create 我要抽奖 10 2 30 1 抽奖`\n"+
			"`/create 我要抽奖 10 2 30 2 私聊机器人参与`")
		if err != nil {
			return fmt.Errorf("error sending reply MarkDown: %v", err)
		}
		return nil
	}

	timeZone, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		return fmt.Errorf("error loading timezone: %v", err)
	}

	eventInfo.ID = time.Now().In(timeZone).Format("20060102150405")

	groupInfo, err := b.getGroupInfo()
	if err != nil {
		return fmt.Errorf("error getting group info: %v", err)
	}
	eventInfo.GroupName = groupInfo.Title

	eventInfo.PrizeName = args[0]

	eventInfo.PrizeCount, err = strconv.Atoi(args[1])
	if err != nil {
		err = b.sendReply(msg, "传递了不受支持的参数--奖品数量")
		if err != nil {
			log.Printf("Error sending reply: %v", err)
		}
		return nil
	}

	if eventInfo.PrizeCount > len(eventInfo.AllPrizes) {
		err = b.sendReply(msg, "奖品数量超出了总奖品数量")
		if err != nil {
			log.Printf("Error sending reply: %v", err)
		}
		return nil
	}

	eventInfo.ChoosePrizes = eventInfo.AllPrizes[:eventInfo.PrizeCount]

	eventInfo.PrizeResultMethod = args[2]
	if eventInfo.PrizeResultMethod == "1" {
		eventInfo.PrizeResult = "按时间开奖"
		inputTime := args[3]
		err = CheckTime(inputTime)
		if err != nil {
			err = b.sendReply(msg, err.Error())
			if err != nil {
				log.Printf("Error sending reply: %v", err)
			}
			return nil
		}
		eventInfo.TimeOfWinners = inputTime
	} else if eventInfo.PrizeResultMethod == "2" {
		eventInfo.PrizeResult = "按人数开奖"
		eventInfo.NumberOfWinners, err = strconv.Atoi(args[3])
		if err != nil {
			err = b.sendReply(msg, "请传递一个整数--开奖人数")
			if err != nil {
				log.Printf("Error sending reply: %v", err)
			}
			return nil
		}
		if eventInfo.PrizeCount < 1 || eventInfo.PrizeCount > len(eventInfo.AllPrizes) || eventInfo.PrizeCount > eventInfo.NumberOfWinners {
			err = b.sendReply(msg, "无效的[奖品数量]，必须大于0,小于或等于开奖人数")
			if err != nil {
				log.Printf("Error sending reply: %v", err)
			}
			return nil
		}
	} else {
		err = b.sendReply(msg, "传递了不受支持的参数--[开奖方法1/2]")
		if err != nil {
			log.Printf("Error sending reply: %v", err)
		}
		return nil
	}

	eventInfo.HowToParticipate = args[4]
	if eventInfo.HowToParticipate == "1" {
		eventInfo.Participate = "群组内发送关键词"
		eventInfo.KeyWord = args[5]
	} else if eventInfo.HowToParticipate == "2" {
		eventInfo.Participate = args[5]
		if eventInfo.Participate != "私聊机器人参与" {
			err = b.sendReply(msg, "不支持的参数--[选2填 私聊机器人参与]")
			if err != nil {
				log.Printf("Error sending reply: %v", err)
			}
			return nil
		}
	} else {
		err = b.sendReply(msg, "传递了不受支持的参数--[参与方法1/2]")
		if err != nil {
			log.Printf("Error sending reply: %v", err)
		}
		return nil
	}

	eventInfo.PrizesList = fmt.Sprintf("%v", strings.Join(eventInfo.ChoosePrizes, "\n"))

	confirmation := fmt.Sprintf(
		"<b>抽奖群：</b> %s\n<b>奖品名称：</b> %s\n<b>奖品数量：</b> %d\n<b>开奖方式：</b> %s\n<b>参与方式：</b> %s\n<b>奖品列表：</b><pre>%v</pre>\n",
		eventInfo.GroupName,
		eventInfo.PrizeName,
		eventInfo.PrizeCount,
		eventInfo.PrizeResult,
		eventInfo.Participate,
		eventInfo.PrizesList,
	)

	if eventInfo.HowToParticipate == "1" {
		confirmation += fmt.Sprintf("<b>关键词：</b> <code>%s</code>\n<b>参与指令：</b> <code>/join %v</code>\n", eventInfo.KeyWord, eventInfo.KeyWord)
	} else if eventInfo.HowToParticipate == "2" {
		confirmation += "<b>参与指令：</b> <code>/join</code>\n"
	}

	if eventInfo.PrizeResultMethod == "1" {
		confirmation += fmt.Sprintf("<b>开奖时间：</b> <code>%s</code> %v\n", eventInfo.TimeOfWinners, timeZone)
	} else if eventInfo.PrizeResultMethod == "2" {
		confirmation += fmt.Sprintf("<b>开奖人数：</b> %d\n", eventInfo.NumberOfWinners)
	}

	//传递值
	b.EventInfoMap[msg.Chat.ID] = eventInfo

	// 添加“是”和“否”按钮用于确认发布抽奖活动
	yesButton := tgbotapi.NewInlineKeyboardButtonData("是", "confirm_create_event")
	noButton := tgbotapi.NewInlineKeyboardButtonData("否", "cancel_create_event")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(yesButton, noButton))

	// 发送带有按钮的消息
	editMsg := tgbotapi.NewMessage(msg.Chat.ID, confirmation)
	editMsg.ParseMode = tgbotapi.ModeHTML
	editMsg.ReplyMarkup = keyboard
	if _, err := b.Bot.Send(editMsg); err != nil {
		log.Printf("Error sending reply with buttons: %v", err)
		return nil
	}
	return nil
}
