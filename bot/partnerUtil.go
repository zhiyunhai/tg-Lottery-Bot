package bot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// 保存参与者信息到数据库
func saveParticipant(db *sql.DB, eventID string, partner Partner) error {
	stmt, err := db.Prepare("INSERT INTO participants(user_id, user_name, event_id) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("stmt.Close error: %v", err)
		}
	}()

	_, err = stmt.Exec(partner.UserID, partner.UserName, eventID)
	if err != nil {
		return fmt.Errorf("error saving participant: %v", err)
	}

	return nil
}

// 查找指定活动ID下的所有参与者
func getParticipantsByEventID(db *sql.DB, eventID string) ([]Partner, error) {
	query := `
	SELECT user_id, user_name 
	FROM participants 
	WHERE event_id = ?;
	`

	rows, err := db.Query(query, eventID)
	if err != nil {
		return nil, fmt.Errorf("getParticipantsByEventID: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close err: %v", err)
		}
	}()

	var participants []Partner
	for rows.Next() {
		var partner Partner
		err := rows.Scan(&partner.UserID, &partner.UserName)
		if err != nil {
			return nil, fmt.Errorf("getParticipantsByEventID: %w", err)
		}
		participants = append(participants, partner)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("getParticipantsByEventID: %w", err)
	}

	return participants, nil
}

// GetUserEventsByUserID 通过用户ID获取此用户参与过的所有活动信息
func GetUserEventsByUserID(db *sql.DB, userID int64) ([]EventInformation, error) {
	// 准备查询语句，获取用户参与的活动ID
	query := `
		SELECT events.id, events.group_name, events.prize_name, events.prize_result_method, events.prize_result,
		       events.how_to_participate, events.participate, events.key_word, events.prizes_list, events.time_of_winners,
		       events.all_prizes, events.choose_prizes, events.prize_count, events.number_of_winners, 
		       events.open_status, events.cancel_status
		FROM participants
		INNER JOIN events ON participants.event_id = events.id
		WHERE participants.user_id = ?
	`

	// 执行查询
	rows, err := db.Query(query, userID)
	if err != nil {
		log.Printf("Error executing query to retrieve events for user %d: %v", userID, err)
		return nil, fmt.Errorf("error executing query to retrieve events for user %d: %v", userID, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close err: %v", err)
		}
	}()

	var events []EventInformation

	// 解析查询结果
	for rows.Next() {
		var event EventInformation
		var allPrizes, choosePrizes string

		err := rows.Scan(&event.ID, &event.GroupName, &event.PrizeName, &event.PrizeResultMethod, &event.PrizeResult,
			&event.HowToParticipate, &event.Participate, &event.KeyWord, &event.PrizesList, &event.TimeOfWinners,
			&allPrizes, &choosePrizes, &event.PrizeCount, &event.NumberOfWinners, &event.OpenStatus, &event.CancelStatus)
		if err != nil {
			log.Printf("Error scanning event data for user %d: %v", userID, err)
			return nil, fmt.Errorf("error scanning event data for user %d: %v", userID, err)
		}

		// Convert the string representations of the allPrizes and choosePrizes fields back to slices
		event.AllPrizes = strings.Split(allPrizes, ",")
		event.ChoosePrizes = strings.Split(choosePrizes, ",")

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error encountered during rows iteration for user %d: %v", userID, err)
		return nil, fmt.Errorf("error encountered during rows iteration for user %d: %v", userID, err)
	}

	return events, nil
}

// 检查用户是否已经参与过该活动
func hasParticipated(db *sql.DB, eventID string, userID int64) (bool, error) {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM participants WHERE event_id = ? AND user_id = ?")
	if err != nil {
		return false, fmt.Errorf("error preparing statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("stmt.Close error: %v", err)
		}
	}()

	var count int
	err = stmt.QueryRow(eventID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error scanning participated event for user %d: %v", userID, err)
	}

	return count > 0, nil
}

// 查询给定活动ID下的参与者数量
func getParticipantCountByEventID(db *sql.DB, eventID string) (count int, err error) {
	query := `
	SELECT COUNT(*) 
	FROM participants 
	WHERE event_id = ?;
	`

	err = db.QueryRow(query, eventID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("getParticipantCountByEventID ERROR: %v", err)
	}

	return count, nil
}
