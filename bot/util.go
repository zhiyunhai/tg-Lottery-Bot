package bot

import (
	"bufio"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

func (b *Bot) checkAdmin(msg *tgbotapi.Message) bool {
	if msg.From.ID != config.AdminUserID {
		return false
	}
	return true
}

func (b *Bot) send(msg *tgbotapi.Message, text string) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, text)
	_, err := b.Bot.Send(message)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) sendReply(msg *tgbotapi.Message, text string) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, text)
	message.ReplyToMessageID = msg.MessageID
	_, err := b.Bot.Send(message)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) sendMarkDown(msg *tgbotapi.Message, text string) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, text)
	message.ParseMode = tgbotapi.ModeMarkdown
	_, err := b.Bot.Send(message)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) sendReplyMarkDown(msg *tgbotapi.Message, text string) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, text)
	message.ParseMode = tgbotapi.ModeMarkdown
	message.ReplyToMessageID = msg.MessageID
	_, err := b.Bot.Send(message)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) sendReplyHTML(msg *tgbotapi.Message, text string) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, text)
	message.ParseMode = tgbotapi.ModeHTML
	message.ReplyToMessageID = msg.MessageID
	_, err := b.Bot.Send(message)
	if err != nil {
		return err
	}
	return nil
}

func CheckTime(inputTime string) error {
	location, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		return fmt.Errorf("load timezone error: %v", err)
	}
	parsedTime, err := time.ParseInLocation("20060102-15:04", inputTime, location)
	if err != nil {
		log.Printf("parsed time error: %v", err)
		return fmt.Errorf("输入的时间格式不符合要求")
	}
	currentTime := time.Now()
	// 如果输入时间小于当前时间则返回错误
	if parsedTime.Before(currentTime) {
		return fmt.Errorf("时间不能是过去的时间")
	}
	// 检查输入时间和当前时间的差异，确保至少比当前时间晚一分钟
	timeDifference := parsedTime.Sub(currentTime)
	if timeDifference < time.Minute {
		return fmt.Errorf("时间太近了，必须晚一分钟")
	}
	return nil
}

// 加载奖品信息，并去除空白行
func loadPrizes() ([]string, error) {

	// 打开奖品文件
	file, err := os.Open(config.PrizeTxtFilePath)
	if err != nil {
		return nil, fmt.Errorf("open file error: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("prizes file close error: %v", err)
		}
	}()

	var prizes []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			prizes = append(prizes, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file error: %v", err)
	}

	return prizes, nil
}

// ReplaceUnChoosePrizes 删除已经选择的奖品
func ReplaceUnChoosePrizes(allPrizes []string, choosePrizes []string) error {
	// 创建一个 map 来存储已选择的奖品
	selectedMap := make(map[string]bool)
	for _, prize := range choosePrizes {
		selectedMap[prize] = true
	}

	// 创建一个切片存储未选择的奖品
	var unselectedPrizes []string
	for _, prize := range allPrizes {
		if !selectedMap[prize] {
			unselectedPrizes = append(unselectedPrizes, prize)
		}
	}
	// 打开文件以写入（覆盖模式）
	file, err := os.Create(config.PrizeTxtFilePath)
	if err != nil {
		return fmt.Errorf("无法打开文件: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("failed to close prize TXT file: %v", err)
		}
	}()

	// 使用缓冲写入以提高写入性能
	writer := bufio.NewWriter(file)
	for _, prize := range unselectedPrizes {
		_, err := writer.WriteString(prize + "\n")
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}

	// 确保所有缓冲区内容写入文件
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("刷新缓冲区失败: %v", err)
	}

	return nil
}

// 添加奖品到txt
func addPrizesToPrizeTxtFile(prizes []string) error {
	// 打开文件进行读写操作
	file, err := os.OpenFile(config.PrizeTxtFilePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open file error: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close prize TXT file: %v", err)
		}
	}()

	// 创建一个切片来存储所有非空白行
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading file error: %v", err)
	}

	// 将奖品列表添加到文件的最后
	lines = append(lines, prizes...)

	// 清空文件内容并将所有行写回文件
	err = file.Truncate(0) // 清空文件
	if err != nil {
		return fmt.Errorf("truncate file error: %v", err)
	}
	_, err = file.Seek(0, 0) // 重置文件指针
	if err != nil {
		return fmt.Errorf("seek file error: %v", err)
	}

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("writing to file error: %v", err)
		}
	}
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("flush writer error: %v", err)
	}

	return nil
}

// 删除奖品txt中的奖品
func removePrizesFromPrizeTxtFile(prizesToRemove []string) error {
	// 打开文件进行读写操作
	file, err := os.OpenFile(config.PrizeTxtFilePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open file error: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close prize TXT file: %v", err)
		}
	}()

	// 读取文件内容
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading file error: %v", err)
	}

	// 创建一个map来快速查找需要删除的奖品
	toRemove := make(map[string]struct{})
	for _, prize := range prizesToRemove {
		toRemove[strings.TrimSpace(prize)] = struct{}{}
	}

	// 保留不在删除列表中的奖品
	var updatedLines []string
	for _, line := range lines {
		if _, found := toRemove[line]; !found {
			updatedLines = append(updatedLines, line)
		}
	}

	// 清空文件内容并将更新后的内容写回文件
	err = file.Truncate(0) // 清空文件
	if err != nil {
		return fmt.Errorf("truncate file error: %v", err)
	}
	_, err = file.Seek(0, 0) // 重置文件指针
	if err != nil {
		return fmt.Errorf("seek file error: %v", err)
	}

	writer := bufio.NewWriter(file)
	for _, line := range updatedLines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("writing to file error: %v", err)
		}
	}
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("flush writer error: %v", err)
	}

	return nil
}

func createAllEventInfoMsg(info EventInformation) (outputMsg string, err error) {
	// 初始化数据库
	db, err := initDB()
	if err != nil {
		return "", fmt.Errorf("init DB error: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close DB: %v", err)
		}
	}()
	// 获取中奖者用户名
	LuckyUserStr, err := getAllLuckyUserName(db, info.ID)
	if err != nil {
		return "", fmt.Errorf("get all lucky user name error: %v", err)
	}

	// 检查参与人数
	NumberOfParticipants, err := getParticipantCountByEventID(db, info.ID)
	if err != nil {
		return "", fmt.Errorf("get number of participants error: %v", err)
	}

	// 创建要发送的消息内容
	outputMsg = fmt.Sprintf(
		"<b>ID:</b> <code>%v</code>\n<b>群组名称:</b> %v\n<b>开奖方式:</b> %v\n<b>参与方式:</b> %v\n<b>奖品数量:</b> %v\n<b>奖品列表:</b> <code>%v</code>\n",
		info.ID, info.GroupName, info.PrizeResult, info.Participate, info.PrizeCount, info.PrizesList)

	if info.PrizeResultMethod == "1" {
		outputMsg += fmt.Sprintf("<b>开奖时间:</b> <code>%v</code> %v\n<b>参与人数:</b> %v\n", info.TimeOfWinners, config.TimeZone, NumberOfParticipants)
	} else if info.PrizeResultMethod == "2" {
		outputMsg += fmt.Sprintf("<b>开奖人数:</b> %v\n<b>参与人数:</b> %v\n", info.NumberOfWinners, NumberOfParticipants)
	}

	if info.HowToParticipate == "1" {
		outputMsg += fmt.Sprintf("<b>抽奖关键词:</b> %v\n", info.KeyWord)
	}

	if info.OpenStatus {
		outputMsg += "<b>开奖状态:</b> 已开奖\n" + fmt.Sprintf("<b>中奖者列表:</b> %v\n", LuckyUserStr)
	} else {
		outputMsg += "<b>开奖状态:</b> 未开奖\n"
	}

	if info.CancelStatus {
		outputMsg += "<b>此活动已取消</b>\n"
	}
	return outputMsg, nil
}

func createUserSeeEventInfoMsg(info EventInformation) (outputMsg string, err error) {
	// 初始化数据库
	db, err := initDB()
	if err != nil {
		return "", fmt.Errorf("init DB error: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close DB: %v", err)
		}
	}()
	// 获取中奖者用户名
	LuckyUserStr, err := getAllLuckyUserName(db, info.ID)
	if err != nil {
		return "", fmt.Errorf("get all lucky user name error: %v", err)
	}

	// 检查参与人数
	NumberOfParticipants, err := getParticipantCountByEventID(db, info.ID)
	if err != nil {
		return "", fmt.Errorf("get number of participants error: %v", err)
	}

	// 创建要发送的消息内容
	outputMsg = fmt.Sprintf(
		"<b>ID:</b> <code>%v</code>\n<b>群组名称:</b> %v\n<b>开奖方式:</b> %v\n<b>参与方式:</b> %v\n<b>奖品数量:</b> %v\n",
		info.ID, info.GroupName, info.PrizeResult, info.Participate, info.PrizeCount)

	if info.PrizeResultMethod == "1" {
		outputMsg += fmt.Sprintf("<b>开奖时间:</b> <code>%v</code> %v\n<b>参与人数:</b> %v\n", info.TimeOfWinners, config.TimeZone, NumberOfParticipants)
	} else if info.PrizeResultMethod == "2" {
		outputMsg += fmt.Sprintf("<b>开奖人数:</b> %v\n<b>参与人数:</b> %v\n", info.NumberOfWinners, NumberOfParticipants)
	}

	if info.HowToParticipate == "1" {
		outputMsg += fmt.Sprintf("<b>抽奖关键词:</b> %v\n", info.KeyWord)
	}

	if info.OpenStatus {
		outputMsg += "<b>开奖状态:</b> 已开奖\n" + fmt.Sprintf("<b>中奖者列表:</b> %v\n", LuckyUserStr)
	} else {
		outputMsg += "<b>开奖状态:</b> 未开奖\n"
	}

	if info.CancelStatus {
		outputMsg += "<b>此活动已取消</b>\n"
	}
	return outputMsg, nil
}

func (b *Bot) sendPrizeToUser(eventID string, luckyUserList []LuckyUser) error {
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

	// 获取活动信息
	eventInfo, err := checkEventInformationFromId(db, eventID)
	if err != nil {
		log.Printf("checkCreateInformation ERROR %v\n", err)
		return err
	}
	var wg sync.WaitGroup

	for _, luckyUser := range luckyUserList {
		wg.Add(1)
		go func(user LuckyUser) {
			defer wg.Done()

			// 构建消息内容，使用Markdown格式
			msgText := fmt.Sprintf(
				"🎉 恭喜你中奖了！\n"+
					"活动ID: %s\n"+
					"群名称: %s\n"+
					"活动名称: %s\n"+
					"奖品: %v",
				eventID, eventInfo.GroupName, eventInfo.PrizeName, user.PrizeInfo,
			)

			// 创建消息对象并指定接收者的ChatID
			message := tgbotapi.NewMessage(user.UserID, msgText)
			// 发送消息
			_, err := b.Bot.Send(message)
			if err != nil {
				log.Printf("无法发送消息给用户 %d: %v", user.UserID, err)
			} else {
				log.Printf("成功发送中奖消息给用户 %s (ID: %d)", user.UserName, user.UserID)
			}
		}(luckyUser)
	}

	wg.Wait() // 等待所有Goroutines完成
	return nil
}

func (b *Bot) sendPrizeDrawMsgToGroup(eventInfo EventInformation) error {
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
	// 获取中奖者用户名
	userNameStr, err := getAllLuckyUserName(db, eventInfo.ID)
	if err != nil {
		log.Printf("getAllLuckyUserName error: %v", err)
		return err
	}
	// 获取参与人数
	NumberOfParticipants, err := getParticipantCountByEventID(db, eventInfo.ID)
	if err != nil {
		log.Printf(err.Error())
	}
	prizeDrawMsg := fmt.Sprintf("🎉抽奖活动开奖啦🎁\n活动ID：%v\n抽奖群：%s\n奖品名称：%s\n奖品数量：%d\n开奖方式：%s\n参与方式：%s\n中奖者名单：%v\n", eventInfo.ID, eventInfo.GroupName, eventInfo.PrizeName, eventInfo.PrizeCount, eventInfo.PrizeResult, eventInfo.Participate, userNameStr)
	if eventInfo.Participate == "群组内发送关键词" {
		prizeDrawMsg += fmt.Sprintf("关键词：%s\n", eventInfo.KeyWord)
	}
	if eventInfo.PrizeResult == "按人数开奖" {
		prizeDrawMsg += fmt.Sprintf("开奖人数：%d\n参与人数：%d\n", eventInfo.NumberOfWinners, NumberOfParticipants)
	} else if eventInfo.PrizeResult == "按时间开奖" {
		prizeDrawMsg += fmt.Sprintf("开奖时间：%s %v\n参与人数：%d\n", eventInfo.TimeOfWinners, config.TimeZone, NumberOfParticipants)
	}
	err = b.sendMsgToGroup(prizeDrawMsg)
	if err != nil {
		log.Printf("sendMsgToGroup err %v\n", err)
		return err
	}
	return nil
}

func (b *Bot) sendMsgToGroup(text string) error {
	// 发送消息到指定群组
	chatID := tgbotapi.ChatConfigWithUser{
		ChatID:             0,
		SuperGroupUsername: config.GroupUserName,
	}

	message := tgbotapi.NewMessageToChannel(chatID.SuperGroupUsername, text)
	message.ParseMode = tgbotapi.ModeHTML

	_, err := b.Bot.Send(message)
	if err != nil {
		log.Printf("sendMsgToGroup err %v\n", err)
		return err
	}
	return nil
}

func (b *Bot) getGroupInfo() (chat tgbotapi.Chat, err error) {
	chatConfig := tgbotapi.ChatInfoConfig{ChatConfig: tgbotapi.ChatConfig{SuperGroupUsername: config.GroupUserName}}
	chat, err = b.Bot.GetChat(chatConfig)
	if err != nil {
		log.Panic(err)
	}
	return chat, nil
}
