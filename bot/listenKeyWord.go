package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

// 监听包含抽奖的关键字
func (b *Bot) listenKeyWordMsg(msg *tgbotapi.Message) error {
	// 检查消息是否来自私聊
	if msg.Chat.IsPrivate() {
		return nil // 忽略来自私聊的消息
	}

	// 忽略命令消息
	if msg.IsCommand() {
		return nil // 忽略命令消息
	}

	// 处理 "抽奖" 关键字
	if strings.Contains(msg.Text, "抽奖") {
		// 初始化数据库
		db, err := initDB()
		if err != nil {
			log.Printf("无法连接到数据库: %v", err)
			return err
		}
		defer func() {
			if err := db.Close(); err != nil {
				log.Printf("关闭数据库连接失败: %v", err)
			}
		}()

		allEventInfo, err := loadAllEvents(db)
		if err != nil {
			log.Println(err)
			return err
		}

		var outputMsg string
		var count int
		for _, val := range allEventInfo {
			if !val.CancelStatus && !val.OpenStatus {
				// 获取参与人数
				NumberOfParticipants, err := getParticipantCountByEventID(db, val.ID)
				if err != nil {
					log.Printf(err.Error())
				}

				eventMsg := fmt.Sprintf("🎟️ *活动 ID:* `%v`\n*🏷️ 群组名称：* %v\n*🏆 活动名称：* %v\n*🎯 开奖方式：* %v\n*👥 参与方式：* %v\n*🎁 奖品数量：* %v\n",
					val.ID, val.GroupName, val.PrizeName, val.PrizeResult, val.Participate, val.PrizeCount)

				if val.PrizeResultMethod == "1" {
					eventMsg += fmt.Sprintf("*⏰ 开奖时间：* `%v` %v\n*👤 参与人数：* %v\n", val.TimeOfWinners, config.TimeZone, NumberOfParticipants)
				} else if val.PrizeResultMethod == "2" {
					eventMsg += fmt.Sprintf("*🏅 开奖人数：* %v\n*👤 参与人数：* %v\n", val.NumberOfWinners, NumberOfParticipants)
				}
				if val.HowToParticipate == "1" {
					eventMsg += fmt.Sprintf("*🔑 抽奖关键词：* %v\n*📩 参与抽奖指令:* `/join %v`\n", val.KeyWord, val.KeyWord)
				} else if val.HowToParticipate == "2" {
					eventMsg += "*📩 参与抽奖指令:* `/join`\n"
				}
				outputMsg += eventMsg + "\n"
				count++
			}
		}

		if outputMsg != "" {
			msgText := fmt.Sprintf("🎉 *共 %d 个正在进行的活动:* 🎉\n\n%v", count, outputMsg)
			return b.sendReplyMarkDown(msg, msgText)
		} else {
			return b.sendReplyMarkDown(msg, "🚫 *当前没有正在进行的抽奖活动。*")
		}
	}
	return nil
}
