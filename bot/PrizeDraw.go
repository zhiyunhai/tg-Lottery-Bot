package bot

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

// 执行开奖操作
func (b *Bot) prizeDraw(eventID string) error {
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
	// 读取活动基本信息
	eventInfo, err := checkEventInformationFromId(db, eventID)
	if err != nil {
		log.Printf("checkCreateInformation ERROR %v\n", err)
		return err
	}

	if eventInfo.CancelStatus == true {
		return fmt.Errorf("取消的活动ID: %v 取消开奖", eventInfo.ID)
	}

	// 读取活动参与者数组
	partnerList, err := getParticipantsByEventID(db, eventID)
	if err != nil {
		log.Println(err)
		return err
	}

	if len(partnerList) == 0 {
		return fmt.Errorf("not found partner")
	}
	// 确保奖品数量不超过参与者数量
	if eventInfo.PrizeCount > len(partnerList) {
		// 修改取消状态为True
		eventInfo.CancelStatus = true
		err = saveEventsInformation(db, eventInfo)
		if err != nil {
			return fmt.Errorf("save create information ERROR %v\n", err)
		}
		return fmt.Errorf("参与者数量不足，无法开奖,活动已取消")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(len(eventID))))

	// 随机打乱参与者列表
	rng.Shuffle(len(partnerList), func(i, j int) {
		partnerList[i], partnerList[j] = partnerList[j], partnerList[i]
	})

	// 抽取中奖者并分配奖品
	for i, winner := range partnerList[:eventInfo.PrizeCount] {
		prize := eventInfo.ChoosePrizes[i]

		// 创建 LuckyUser 结构体
		luckyUser := LuckyUser{
			UserID:    winner.UserID,
			UserName:  winner.UserName,
			PrizeInfo: prize,
		}

		// 将中奖者信息保存到数据库
		err = saveLuckyUser(db, luckyUser, eventID)
		if err != nil {
			log.Printf("saveLuckyUser: %v", err)
			return err
		}
	}
	// 修改开奖状态为True
	eventInfo.OpenStatus = true
	err = saveEventsInformation(db, eventInfo)
	if err != nil {
		log.Printf("saveEventsInformation ERROR %v\n", err)
		return err
	}
	luckyUsersList, err := getLuckyUsersListByEventID(db, eventID)
	if err != nil {
		log.Printf("getLuckyUsersListByEventID: %v", err)
		return err
	}
	err = b.sendPrizeDrawMsgToGroup(eventInfo)
	if err != nil {
		log.Printf("sendPrizeDrawMsgToGroup err %v\n", err)
		return err
	}
	err = b.sendPrizeToUser(eventID, luckyUsersList)
	if err != nil {
		log.Printf("sendPrizeToUser err %v\n", err)
		return err
	}
	log.Printf("开奖完成，活动ID: %s", eventID)
	return nil
}

// 根据时间开奖
func (b *Bot) regularPrizeDraw() error {
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
	// 解析时区
	timeLoc, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		log.Printf("load timezone error: %v", err)
		return err
	}
	// 加载所有活动
	eventInfo, err := loadAllEvents(db)
	if err != nil {
		return fmt.Errorf("loadCreateInformation: %v", err)
	}
	if len(eventInfo) == 0 {
		log.Printf("没有活动：%v", eventInfo)
		return nil
	}
	for _, value := range eventInfo {
		if !value.OpenStatus && !value.CancelStatus {
			if value.PrizeResultMethod == "1" {
				openTime, err := time.ParseInLocation("20060102-15:04", value.TimeOfWinners, timeLoc)
				if err != nil {
					log.Printf("解析开奖时间失败: %v", err)
					return err
				}

				b.timersMu.Lock()
				// 检查是否已经存在定时任务
				if _, exists := b.drawTimers[value.ID]; exists {
					log.Printf("定时任务已存在，活动ID: %s", value.ID)
					b.timersMu.Unlock()
					continue
				}

				if time.Now().In(timeLoc).After(openTime) {
					// 如果开奖时间小于当前时间，直接开奖
					err = b.prizeDraw(value.ID)
					if err != nil {
						log.Printf("prizeDraw: %v", err)
						b.timersMu.Unlock()
						return err
					}
				} else if value.PrizeResultMethod != "2" {
					// 否则设定定时任务
					duration := time.Until(openTime)
					timer := time.AfterFunc(duration, func() {
						err := b.prizeDraw(value.ID)
						if err != nil {
							log.Printf("定时开奖失败: %v", err)
						} else {
							log.Printf("定时开奖成功，活动ID: %s", value.ID)
						}

						// 任务执行后删除记录
						b.timersMu.Lock()
						delete(b.drawTimers, value.ID)
						b.timersMu.Unlock()
					})
					// 记录定时任务
					b.drawTimers[value.ID] = timer
					log.Printf("定时任务已设定，活动ID: %s，时间: %v", value.ID, openTime)
				}
				b.timersMu.Unlock()
			}
		}
	}
	return nil
}
