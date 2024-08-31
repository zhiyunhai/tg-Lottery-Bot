package bot

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

// 初始化数据库
func initDB() (*sql.DB, error) {
	// 确保 .db 文件夹存在
	dbFolderPath := "./.db"
	if _, err := os.Stat(dbFolderPath); os.IsNotExist(err) {
		err = os.Mkdir(dbFolderPath, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("无法创建数据库文件夹: %v", err)
		}
	}

	// 数据库文件路径
	dbFilePath := filepath.Join(dbFolderPath, "info.db")

	// 连接到 SQLite 数据库
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开数据库连接: %v", err)
	}

	// 创建抽奖活动表
	sqlStmtEvents := `
	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		group_name TEXT,
		prize_name TEXT,
		prize_result_method TEXT,
		prize_result TEXT,
		how_to_participate TEXT,
		participate TEXT,
		key_word TEXT,
		prizes_list TEXT,
		time_of_winners TEXT,
		all_prizes TEXT,
		choose_prizes TEXT,
		prize_count INTEGER,
		number_of_winners INTEGER,
		open_status BOOLEAN,
		cancel_status BOOLEAN
	);
	`

	_, err = db.Exec(sqlStmtEvents)
	if err != nil {
		err = db.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("无法创建活动表: %v", err)
	}

	// 创建参与者表
	sqlStmtParticipants := `
	CREATE TABLE IF NOT EXISTS participants (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		user_name TEXT,
		event_id TEXT,
		FOREIGN KEY (event_id) REFERENCES events(id)
	);
	`

	_, err = db.Exec(sqlStmtParticipants)
	if err != nil {
		err = db.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("无法创建参与者表: %v", err)
	}

	// 创建中奖者表
	sqlStmtLuckyUser := `
	CREATE TABLE IF NOT EXISTS luckyUser (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		user_name TEXT,
		prize_info TEXT,
		event_id TEXT,
		FOREIGN KEY (event_id) REFERENCES events(id)
	);
	`

	_, err = db.Exec(sqlStmtLuckyUser)
	if err != nil {
		err = db.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("无法创建中奖者表: %v", err)
	}
	return db, nil
}
