# Telegram 抽奖机器人 | Telegram Lottery Bot

---

## 广告

美国T-Mobile USA原生顶级蜂窝网络ip专线节点，新用户首次购买仅25RMB(CNY) 1GB
，服务介绍 https://sooon.cc/
订购网址 https://tmo.sooon.cc/
示例IP：172.58.27.130，适合各类美国业务，登录美国银行，paypal保号,gv保号,gv注册等。
官方电报  https://t.me/zhiyh8

### 项目简介

此项目主要是为了解决群主给群友公平发福利的难题。欢迎加入群组 https://t.me/zhiyunhai 体验此机器人。
要使用此程序，您需要一个 Telegram 帐户和一个 Telegram 机器人。
您可以通过与[BotFather](https://t.me/BotFather)交互并按照提供的说明创建 Telegram 机器人。
我已经彻底测试了这个程序，发现它运行良好。如果您在使用中发现了bug，欢迎您提交issue。 或通过电子邮件与我联络。

### 功能介绍

#### 管理员指令 📜

- **/id** - 查看你自己的用户ID。
- **/create** - 创建一个新的抽奖活动。
    - **参数**：
        - `活动名称` - 设置抽奖活动的名称。
        - `奖品数量` - 设置奖品的数量。
        - `开奖方法1/2` - 选择开奖方法：
            1. 按时间开奖
            2. 按人数开奖
        - `选1填时间，选2填人数` - 根据选择的开奖方法填入相应的信息。
        - `参与方法1/2` - 选择参与方式：
            1. 群组内发送关键词参与
            2. 私聊机器人参与
        - `选1填关键词，选2填 私聊机器人参与` - 根据选择的参与方式填入相应的信息。
    - **命令示例**：
        - `/create 我要抽奖 10 1 20240823-23:07 1 抽奖`
        - `/create 我要抽奖 10 1 20240823-23:07 2 私聊机器人参与`
        - `/create 我要抽奖 10 2 30 1 抽奖`
        - `/create 我要抽奖 10 2 30 2 私聊机器人参与`
- **/add** - 添加奖品。
- **/delete** - 删除奖品。
- **/list** - 查看库存中的所有奖品，支持分页展示。
- **/on** - 查看正在进行的活动，支持指定页码（可选）。
- **/cancel** - 查看已取消的活动，支持指定页码（可选）。
- **/history** - 查看历史抽奖活动，支持指定页码（可选）。
- **/open** - 手动开奖，需传入活动ID。
- **/close** - 关闭正在进行的活动，需传入活动ID。

#### 参与者指令 📋

- **/see** - 查看已参与的活动，支持指定页码（可选）。
- **/join** - 参加参与方式为“私聊机器人参与”的抽奖活动。
- **/join 关键词** - 参加参与方式为“群组内发送关键词”的抽奖活动。
- **/prize** - 查看中奖历史，支持指定页码（可选）。

### 部署指南

1. 前往 [GitHub Releases](https://github.com/zhiyunhai/tg-Lottery-Bot/releases) 页面下载适合你系统的最新版本二进制文件。

2. 在项目根目录下创建 `config.yaml` 文件，按照以下格式进行配置：

```yaml
api_token: "YOUR-TG-BOT-API-TOKEN"
admin_user_id: 123456789
group_user_name: "@example"
prize_txt_file_path: "example.txt"
timezone: "Asia/Shanghai"  # 可选，不指定则使用UTC世界标准时间
```

如需在后台运行此程序，可以使用以下 `systemd` 服务文件进行配置：

```ini
[Unit]
Description=tgLotteryBot
After=network.target

[Service]
Type=simple
User=your_user_name
WorkingDirectory=/your/binary/path/here/
ExecStart=/your/binary/path/here/tgLotteryBot
Restart=on-failure
RestartSec=10s
TimeoutStopSec=30s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target

```

将此文件保存为 `/etc/systemd/system/tgLotteryBot.service`，然后运行以下命令启动和启用服务：

```bash
sudo systemctl daemon-reload
sudo systemctl start tgLotteryBot  # 启动服务
sudo systemctl enable tgLotteryBot # 设置开机自启
sudo systemctl status tgLotteryBot # 检查服务状态
```

### 自行编译

由于项目依赖于 `github.com/mattn/go-sqlite3` 包，该包依赖于 C 库，因此在编译时需要在目标平台上安装相应的依赖。

#### 安装 Go 语言环境

1. **下载 Go**：
    - 前往 [Go 官方网站](https://go.dev/dl/) 下载适用于您的操作系统的 Go 安装包。

2. **安装 Go**：
    - 根据下载的安装包，按照操作系统的要求进行安装。

3. **验证安装**：
    - 打开终端（或命令提示符），输入以下命令来检查 Go 是否安装成功：
      ```bash
      go version
      ```

#### 安装依赖库

- **Linux**（Debian/Ubuntu）:
  ```bash
  sudo apt-get update
  sudo apt-get install gcc
  ```

- **macOS**:
  ```bash
  xcode-select --install
  ```

#### 编译和运行

1. **获取代码**：
   ```bash
   git clone https://github.com/zhiyunhai/tg-Lottery-Bot.git
   cd tg-Lottery-Bot
   ```

2. **编译代码**：
   ```bash
   go build -o tgLotteryBot
   ```

3. **运行程序**：
   ```bash
   ./tgLotteryBot
   ```


### 捐赠

喜欢这个项目吗？请考虑支持开发者，通过以下方式捐赠：

- 支付宝、微信：访问 [支持我们](https://www.zyh8.com/%e6%94%af%e6%8c%81%e6%88%91%e4%bb%ac/)
- USDT (TRC20): `TJsH3fGwtmr1nSbyBfFN6uXvBArpQjebJ6`

### 联系方式

- Email: support@zyh8.com
- 我的博客: [https://www.zyh8.com](https://www.zyh8.com)

---
