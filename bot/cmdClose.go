package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func (b *Bot) cmdClose(msg *tgbotapi.Message) error {
	if !b.checkAdmin(msg) {
		err := b.sendReply(msg, "You are not an admin")
		if err != nil {
			return err
		}
		return nil
	}

	if !msg.Chat.IsPrivate() {
		return b.sendReply(msg, "请在私聊中使用管理员指令")
	}

	// 初始化数据库
	db, err := initDB()
	if err != nil {
		log.Printf("无法连接到数据库: %v", err)
		reply := tgbotapi.NewMessage(msg.Chat.ID, "无法连接到数据库")
		_, err := b.Bot.Send(reply)
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		}
	}()

	args := strings.Split(msg.CommandArguments(), " ")
	if len(args) == 0 || args[0] == "" {
		err = b.sendReply(msg, "/close [活动ID]")
		if err != nil {
			log.Printf("sendReply: %v", err)
			return err
		}
		return nil
	}

	inputID := args[0]
	info, err := checkEventInformationFromId(db, inputID)
	if err != nil {
		log.Printf("checkEventInformationFromId: %v", err)
		return err
	}

	if info.CancelStatus {
		err = b.sendReply(msg, "此活动已取消，请勿重复取消")
		if err != nil {
			log.Printf("sendReply: %v", err)
			return err
		}
		return nil
	}

	if info.OpenStatus {
		err = b.sendReply(msg, "此活动已经开奖，无需取消")
		if err != nil {
			log.Printf("sendReply: %v", err)
			return err
		}
		return nil
	}

	info.CancelStatus = true
	err = saveEventsInformation(db, info)
	if err != nil {
		log.Printf("saveEventsInformation: %v", err)
	}

	outputMsg, err := createAllEventInfoMsg(info)
	if err != nil {
		log.Printf("createAllEventInfoMsg: %v", err)
	}
	outputMsg += "取消成功"

	err = b.sendReplyHTML(msg, outputMsg)
	if err != nil {
		return err
	}
	return nil
}
