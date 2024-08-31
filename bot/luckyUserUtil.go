package bot

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

// 保存中奖者信息
func saveLuckyUser(db *sql.DB, luckyUser LuckyUser, eventID string) error {
	sqlStmt := `
	INSERT INTO luckyUser (user_id, user_name, prize_info, event_id) 
	VALUES (?, ?, ?, ?);
	`

	_, err := db.Exec(sqlStmt, luckyUser.UserID, luckyUser.UserName, luckyUser.PrizeInfo, eventID)
	if err != nil {
		log.Printf("无法保存活动ID %s 的中奖者信息: %v", eventID, err)
		return fmt.Errorf("无法保存中奖者信息，请稍后再试")
	}
	return nil
}

// 查找指定活动ID下的中奖者信息
func getLuckyUsersListByEventID(db *sql.DB, eventID string) ([]LuckyUser, error) {
	query := `
	SELECT user_id, user_name ,prize_info
	FROM luckyUser
	WHERE event_id = ?;
	`

	rows, err := db.Query(query, eventID)
	if err != nil {
		log.Printf("查询活动ID %s 的中奖者信息失败: %v", eventID, err)
		return nil, fmt.Errorf("无法获取中奖者信息，请稍后再试")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close 错误: %v", err)
		}
	}()

	var luckyUserList []LuckyUser
	for rows.Next() {
		var luckyUser LuckyUser
		err := rows.Scan(&luckyUser.UserID, &luckyUser.UserName, &luckyUser.PrizeInfo)
		if err != nil {
			log.Printf("读取活动ID %s 的中奖者信息失败: %v", eventID, err)
			return nil, fmt.Errorf("无法获取中奖者信息，请稍后再试")
		}
		luckyUserList = append(luckyUserList, luckyUser)
	}

	if err = rows.Err(); err != nil {
		log.Printf("遍历活动ID %s 的中奖者信息时出错: %v", eventID, err)
		return nil, fmt.Errorf("无法获取中奖者信息，请稍后再试")
	}

	return luckyUserList, nil
}

// 通过活动ID查询所有中奖者组成的字符串
func getAllLuckyUserName(db *sql.DB, eventID string) (AllLuckyUserName string, err error) {
	LuckyUserNameList, err := getLuckyUserNameListByEventID(db, eventID)
	if err != nil {
		return "", fmt.Errorf("getAllLuckyUserName ERROR: %v", err)
	}
	// 为每个用户名添加 '@' 前缀
	for i, userName := range LuckyUserNameList {
		LuckyUserNameList[i] = "@" + userName
	}
	// 将所有用户名连接成一个完整的字符串，每个用户名占一行
	AllLuckyUserName = strings.Join(LuckyUserNameList, "\n")
	return AllLuckyUserName, nil
}

// 通过活动ID查询中奖者用户名列表
func getLuckyUserNameListByEventID(db *sql.DB, eventID string) (luckyUserNameList []string, err error) {
	// 查询数据库，获取指定活动ID对应的所有中奖者用户名
	query := `
	SELECT user_name 
	FROM luckyUser 
	WHERE event_id = ?;
	`

	rows, err := db.Query(query, eventID)
	if err != nil {
		return nil, fmt.Errorf("error querying list of winner usernames for campaign ID %s: %v", eventID, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close err: %v", err)
		}
	}()

	for rows.Next() {
		var userName string
		if err := rows.Scan(&userName); err != nil {
			return nil, fmt.Errorf("error reading winner's username for campaign ID %s: %v", eventID, err)
		}
		luckyUserNameList = append(luckyUserNameList, userName)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error traversing list of winner usernames for campaign ID %s: %v", eventID, err)
	}

	return luckyUserNameList, nil
}

func getWinInfoByUserID(db *sql.DB, userID int64) ([]winInfo, error) {
	// 定义返回的切片
	var winInfos []winInfo

	// 查询用户的中奖记录
	queryLuckyUser := `
	SELECT user_id, user_name, prize_info, event_id
	FROM luckyUser
	WHERE user_id = ?
	`
	rows, err := db.Query(queryLuckyUser, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户中奖记录失败: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close err: %v", err)
		}
	}()

	// 遍历中奖记录
	for rows.Next() {
		var luckyUser LuckyUser
		err = rows.Scan(&luckyUser.UserID, &luckyUser.UserName, &luckyUser.PrizeInfo, &luckyUser.EventID)
		if err != nil {
			return nil, fmt.Errorf("扫描中奖记录失败: %v", err)
		}

		// 查询对应的活动信息
		var event EventInformation
		queryEvent := `
		SELECT id, group_name, prize_name, prize_result_method, how_to_participate, key_word, prizes_list, time_of_winners, prize_count, number_of_winners
		FROM events
		WHERE id = ?
		`
		err = db.QueryRow(queryEvent, luckyUser.EventID).Scan(
			&event.ID, &event.GroupName, &event.PrizeName, &event.PrizeResultMethod, &event.HowToParticipate,
			&event.KeyWord, &event.PrizesList, &event.TimeOfWinners, &event.PrizeCount, &event.NumberOfWinners)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, fmt.Errorf("查询活动信息失败: %w", err)
		}

		// 将数据填充到 winInfo 结构体中
		win := winInfo{
			ID:                event.ID,
			GroupName:         event.GroupName,
			PrizeName:         event.PrizeName,
			PrizeResultMethod: event.PrizeResultMethod,
			HowToParticipate:  event.HowToParticipate,
			KeyWord:           event.KeyWord,
			PrizesList:        event.PrizesList,
			TimeOfWinners:     event.TimeOfWinners,
			PrizeCount:        event.PrizeCount,
			NumberOfWinners:   event.NumberOfWinners,
			PrizeInfo:         luckyUser.PrizeInfo, // 将奖品信息添加到 winInfo
		}

		// 将 winInfo 添加到返回的切片中
		winInfos = append(winInfos, win)
	}

	return winInfos, nil
}
