package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func (b *Bot) cmdJoin(msg *tgbotapi.Message) error {
	// 初始化数据库
	db, err := initDB()
	if err != nil {
		log.Printf("Unable to connect to database: %v", err)
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()

	// 加载历史抽奖活动信息
	eventInfo, err := loadAllEvents(db)
	if err != nil {
		log.Printf("Unable to load all events: %v", err)
		return err
	}
	if len(eventInfo) == 0 {
		err = b.sendReply(msg, "没有活动")
		if err != nil {
			return err
		}
		return nil
	}

	command := msg.Text
	args := strings.Fields(command)
	if len(args) == 0 {
		return b.sendReply(msg, "无效的命令")
	}

	var userKeyWord string
	if len(args) > 1 {
		userKeyWord = args[1]
	}

	userID := msg.From.ID
	userName := msg.From.UserName
	var haveEvents bool
	for _, value := range eventInfo {
		// 检查活动是否未取消且未开奖
		if !value.CancelStatus && !value.OpenStatus {
			haveEvents = true
			// 判断参与方式
			if value.HowToParticipate == "1" { // 通过关键词参与
				if userKeyWord != value.KeyWord {
					continue
				}
				if msg.Chat.IsPrivate() {
					return b.sendReply(msg, "请在群组中发送")
				}
			} else if value.HowToParticipate == "2" { // 直接通过 /join 参与
				if userKeyWord != "" {
					continue
				}
				if !msg.Chat.IsPrivate() {
					return b.sendReply(msg, "请在私聊中发送")
				}
			}

			// 检查用户是否已经参与过此活动
			alreadyParticipated, err := hasParticipated(db, value.ID, userID)
			if err != nil {
				log.Printf("hasParticipated: %v", err)
				err = b.sendReply(msg, fmt.Sprintf("你已经参与过活动: %v", value.ID))
				if err != nil {
					return err
				}
				return nil
			}
			if !alreadyParticipated {
				// 创建一个新的参与者项
				newPartner := Partner{
					UserID:   userID,
					UserName: userName,
				}

				// 保存到数据库
				err = saveParticipant(db, value.ID, newPartner)
				if err != nil {
					log.Printf("saveParticipant: %v", err)
					return b.sendReply(msg, err.Error())
				}

				// 构建回复消息
				var NumberOfParticipants int
				var replyMessage string

				replyMessage = fmt.Sprintf("🎉*你已成功参与活动:*🎉\n\n*🎟️ 活动 ID:* `%s`\n*🏷️ 活动名称:* %s\n*🎁 奖品数量:* %d\n",
					value.ID, value.PrizeName, value.PrizeCount)

				if value.PrizeResultMethod == "1" { // 按时间开奖
					NumberOfParticipantsTime, err := getParticipantCountByEventID(db, value.ID)
					if err != nil {
						log.Printf(err.Error())
					}
					replyMessage += fmt.Sprintf("*⏰ 开奖时间:* %s %s\n*👥 参与人数:* %d\n", value.TimeOfWinners, config.TimeZone, NumberOfParticipantsTime)
				} else if value.PrizeResultMethod == "2" { // 按人数开奖
					NumberOfParticipants, err = getParticipantCountByEventID(db, value.ID)
					if err != nil {
						log.Printf(err.Error())
					}
					replyMessage += fmt.Sprintf("*🏆 开奖人数:* %d\n*👥 参与人数:* %d\n", value.NumberOfWinners, NumberOfParticipants)

					// 人数到了自动开奖
					if value.NumberOfWinners == NumberOfParticipants {
						err = b.prizeDraw(value.ID)
						if err != nil {
							log.Printf("prizeDraw: %v", err)
							return b.sendReply(msg, "❌ 参与者数量不足或系统出现严重错误，开奖失败，请联系管理员！")
						}
					}
				}

				if value.HowToParticipate == "1" {
					replyMessage += fmt.Sprintf("*📲 参与方式:* %s\n*🔑 关键词:* `%s`\n", value.Participate, value.KeyWord)
				} else if value.HowToParticipate == "2" {
					replyMessage += fmt.Sprintf("*📲 参与方式:* %s\n", value.Participate)
				}

				// 发送 Markdown 格式的消息
				err = b.sendReplyMarkDown(msg, replyMessage)
				if err != nil {
					return err
				}
			} else {
				err = b.sendReply(msg, "你已经参与过活动: "+value.ID)
				if err != nil {
					log.Printf("send Reply err: %v", err)
					return err
				}
			}
		}
	}

	if !haveEvents {
		err = b.sendReply(msg, "没有活动")
		if err != nil {
			return err
		}
	}

	return nil
}
