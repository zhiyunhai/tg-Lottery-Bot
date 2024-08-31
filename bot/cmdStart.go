package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) cmdStart(msg *tgbotapi.Message) error {
	response := `🎉 *欢迎使用抽奖机器人！* 🎉

/id - 查看你自己的用户ID

📜 **管理员指令**
*开奖方法：*
1. 按时间开奖  
2. 按人数开奖

*参与方法：*
1. 群组内发送关键词参与  
2. 私聊机器人参与

/create [活动名称] [奖品数量] [开奖方法1/2] [选1填时间，选2填人数] [参与方法1/2] [选1填关键词，选2填 私聊机器人参与] - 创建一个新的抽奖活动

*命令示例：*
` + "`" + `/create 我要抽奖 10 1 20240823-23:07 1 抽奖` + "`" + `
` + "`" + `/create 我要抽奖 10 1 20240823-23:07 2 私聊机器人参与` + "`" + `
` + "`" + `/create 我要抽奖 10 2 30 1 抽奖` + "`" + `
` + "`" + `/create 我要抽奖 10 2 30 2 私聊机器人参与` + "`" + `

/add - 添加奖品  
/delete - 删除奖品  
/list - 查看库存中的所有奖品  
/on [指定页码（可选）]- 查看正在进行的活动
/cancel [指定页码（可选）] - 查看已取消的活动
/history [指定页码（可选）] - 查看历史抽奖活动
/open [活动ID] - 手动开奖  
/close [活动ID] - 关闭正在进行的活动

📋 **参与者指令**
/see [指定页码（可选）] - 查看已参与的活动
/join - 参加参与方式为“私聊机器人参与”的抽奖活动  
/join [关键词] - 参加参与方式为“群组内发送关键词”的抽奖活动

🎁 **领取奖品**  
/prize [指定页码（可选）] - 查看中奖历史`

	err := b.sendMarkDown(msg, response)
	if err != nil {
		return err
	}
	return nil
}
