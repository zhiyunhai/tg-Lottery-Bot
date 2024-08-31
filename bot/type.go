package bot

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var config Config

// Config 配置文件
type Config struct {
	ApiToken         string `yaml:"api_token"`
	AdminUserID      int64  `yaml:"admin_user_id"`
	GroupUserName    string `yaml:"group_user_name"`
	PrizeTxtFilePath string `yaml:"prize_txt_file_path"`
	TimeZone         string `yaml:"timezone"`
}

// EventInformation 活动信息
type EventInformation struct {
	ID                string   `json:"id"`                //活动ID
	GroupName         string   `json:"groupName"`         //群组名称
	PrizeName         string   `json:"prizeName"`         //活动名称
	PrizeResultMethod string   `json:"prizeResultMethod"` //选择开奖方式
	PrizeResult       string   `json:"prizeResult"`       //开奖方式
	HowToParticipate  string   `json:"howToParticipate"`  //选择参与方式
	Participate       string   `json:"participate"`       //参与方式
	KeyWord           string   `json:"keyWord"`           //抽奖关键词
	PrizesList        string   `json:"prizesList"`        //返回奖品列表组成的字符串
	TimeOfWinners     string   `json:"timeOfWinners"`     //开奖时间
	AllPrizes         []string `json:"allPrizes"`         //全部奖品
	ChoosePrizes      []string `json:"choosePrizes"`      //选中的奖品
	PrizeCount        int      `json:"prizeCount"`        //奖品数量
	NumberOfWinners   int      `json:"numberOfWinners"`   //开奖人数
	OpenStatus        bool     `json:"openStatus"`        //开奖状态
	CancelStatus      bool     `json:"cancelStatus"`      //是否为取消的活动
}

// Partner 参与者
type Partner struct {
	UserID   int64  `json:"user_id"`
	UserName string `json:"user_name"`
}

// LuckyUser 中奖者名单
type LuckyUser struct {
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name"`
	PrizeInfo string `json:"prize_info"`
	EventID   string `json:"event_id"`
}

// winInfo 中奖信息
type winInfo struct {
	ID                string `json:"id"`                //活动ID
	GroupName         string `json:"groupName"`         //群组名称
	PrizeName         string `json:"prizeName"`         //活动名称
	PrizeResultMethod string `json:"prizeResultMethod"` //选择开奖方式
	HowToParticipate  string `json:"howToParticipate"`  //选择参与方式
	KeyWord           string `json:"keyWord"`           //抽奖关键词
	PrizesList        string `json:"prizesList"`        //返回抽奖列表
	TimeOfWinners     string `json:"timeOfWinners"`     //开奖时间
	PrizeCount        int    `json:"prizeCount"`        //奖品数量
	NumberOfWinners   int    `json:"numberOfWinners"`   //开奖人数
	PrizeInfo         string `json:"prizeInfo"`         //奖品
}

func readConfig() {
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Println("请在软件所在目录下创建 config.yaml 配置文件")
		log.Fatalf("Error reading config.yaml: %v", err)
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Error parsing config.yaml: %v", err)
	}
	if config.TimeZone == "" {
		config.TimeZone = "UTC"
	}
}
