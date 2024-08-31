package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"sync"
	"time"
)

type Bot struct {
	Bot            *tgbotapi.BotAPI
	allEvents      []EventInformation         // 全部活动
	onEvent        []EventInformation         // 正在进行的活动
	cancelEvents   []EventInformation         // 取消的活动
	userJoinEvents []EventInformation         //用户参与过的所有活动信息
	winInfoList    []winInfo                  //用户的中奖信息
	prizeList      []string                   //奖品列表
	UserStates     map[int64]string           // 用于跟踪用户的状态
	userStatesMu   sync.Mutex                 // 用于保护 UserStates 的并发访问
	EventInfoMap   map[int64]EventInformation // 用于暂存创建的活动信息
	eventInfoMapMu sync.Mutex                 // 用于保护 EventInfoMap 的并发访问
	drawTimers     map[string]*time.Timer     // 用于管理开奖的定时任务
	timersMu       sync.Mutex                 // 用于保护 drawTimers 的并发访问
}

func NewBot() (*Bot, error) {
	readConfig() //加载配置文件
	botInstance, err := tgbotapi.NewBotAPI(config.ApiToken)
	if err != nil {
		return nil, err
	}
	// 设置命令列表
	commands := []tgbotapi.BotCommand{
		{Command: "id", Description: "查看你自己的用户ID"},
		{Command: "start", Description: "查看帮助"},
		{Command: "see", Description: "查看已参与的活动"},
		{Command: "join", Description: "参加抽奖活动"},
		{Command: "prize", Description: "查看中奖历史信息"},
	}

	// 设置命令到 Telegram 服务器
	_, err = botInstance.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Fatalf("Failed to set commands: %v", err)
	}
	bot := &Bot{
		Bot:          botInstance,
		drawTimers:   make(map[string]*time.Timer),
		UserStates:   make(map[int64]string),
		EventInfoMap: make(map[int64]EventInformation),
	}

	return bot, nil
}

func (b *Bot) Start() {
	log.Println("Bot started...")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.Bot.GetUpdatesChan(u)

	// 刷新开奖时间定时器
	err := b.regularPrizeDraw()
	if err != nil {
		log.Printf("Failed to regular prize draw: %v", err)
	}

	for update := range updates {
		// 检查 Bot 是否已经初始化
		if b.Bot == nil {
			log.Println("Bot is nil, skipping update handling.")
			continue
		}

		if update.Message != nil {
			// 处理命令
			b.handleUpdate(update.Message)

			// 检查消息是否为 nil，并且不是命令
			if update.Message.Text != "" && !update.Message.IsCommand() {
				err := b.listenKeyWordMsg(update.Message)
				if err != nil {
					log.Printf("Failed to handle text message: %v", err)
				}
			}
		} else if update.CallbackQuery != nil {
			// 处理回调查询
			b.handleCallbackQuery(update.CallbackQuery)
		}
	}

}

func (b *Bot) Stop() {
	log.Println("Bot stopped.")
}
