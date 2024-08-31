package bot

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
)

// 保存活动信息到数据库
func saveEventsInformation(db *sql.DB, info EventInformation) error {
	// 序列化奖品列表和选择的奖品为JSON字符串
	allPrizesJSON, err := json.Marshal(info.AllPrizes)
	if err != nil {
		return fmt.Errorf("error marshalling all prizes: %v", err)
	}

	choosePrizesJSON, err := json.Marshal(info.ChoosePrizes)
	if err != nil {
		return fmt.Errorf("error marshalling choose prizes: %v", err)
	}

	stmt, err := db.Prepare(`
	INSERT INTO events (
		id, group_name, prize_name, prize_result_method, prize_result, how_to_participate,
		participate, key_word, prizes_list, time_of_winners, all_prizes, choose_prizes,
		prize_count, number_of_winners, open_status, cancel_status
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		group_name=excluded.group_name, prize_name=excluded.prize_name, prize_result_method=excluded.prize_result_method,
		prize_result=excluded.prize_result, how_to_participate=excluded.how_to_participate, participate=excluded.participate,
		key_word=excluded.key_word, prizes_list=excluded.prizes_list, time_of_winners=excluded.time_of_winners,
		all_prizes=excluded.all_prizes, choose_prizes=excluded.choose_prizes, prize_count=excluded.prize_count,
		number_of_winners=excluded.number_of_winners, open_status=excluded.open_status, cancel_status=excluded.cancel_status
	`)
	if err != nil {
		return fmt.Errorf("error saving create information: %v", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("error closing stmt: %v", err)
		}
	}()

	_, err = stmt.Exec(
		info.ID, info.GroupName, info.PrizeName, info.PrizeResultMethod, info.PrizeResult,
		info.HowToParticipate, info.Participate, info.KeyWord, info.PrizesList, info.TimeOfWinners,
		string(allPrizesJSON), string(choosePrizesJSON), info.PrizeCount, info.NumberOfWinners,
		info.OpenStatus, info.CancelStatus,
	)
	if err != nil {
		return fmt.Errorf("error saving create information: %v", err)
	}

	return nil
}

// 加载所有活动信息
func loadAllEvents(db *sql.DB) (AllEvent []EventInformation, err error) {
	rows, err := db.Query("SELECT * FROM events")
	if err != nil {
		return nil, fmt.Errorf("loadCreateInformation ERROR: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close ERROR: %v", err)
		}
	}()

	for rows.Next() {
		var info EventInformation
		var allPrizesJSON, choosePrizesJSON string

		err = rows.Scan(
			&info.ID, &info.GroupName, &info.PrizeName, &info.PrizeResultMethod, &info.PrizeResult,
			&info.HowToParticipate, &info.Participate, &info.KeyWord, &info.PrizesList,
			&info.TimeOfWinners, &allPrizesJSON, &choosePrizesJSON,
			&info.PrizeCount, &info.NumberOfWinners, &info.OpenStatus, &info.CancelStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("scan events ERROR: %v", err)
		}

		// 反序列化奖品列表和选择的奖品
		if err = json.Unmarshal([]byte(allPrizesJSON), &info.AllPrizes); err != nil {
			return nil, fmt.Errorf("json unmarshal allPrizes ERROR: %v", err)
		}
		if err = json.Unmarshal([]byte(choosePrizesJSON), &info.ChoosePrizes); err != nil {
			return nil, fmt.Errorf("json unmarshal choosePrizes ERROR: %v", err)
		}

		AllEvent = append(AllEvent, info)
	}

	sort.Slice(AllEvent, func(i, j int) bool {
		return AllEvent[i].ID < AllEvent[j].ID
	})

	return AllEvent, nil
}

// 加载未开奖和未取消的所有活动信息
func loadNoCancelAndNoOpenEvents(db *sql.DB) (onEvent []EventInformation, err error) {
	rows, err := db.Query("SELECT * FROM events WHERE open_status = 0 AND cancel_status = 0")
	if err != nil {
		return nil, fmt.Errorf("loadCreateInformation ERROR: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close ERROR: %v", err)
		}
	}()

	for rows.Next() {
		var info EventInformation
		var allPrizesJSON, choosePrizesJSON string

		err = rows.Scan(
			&info.ID, &info.GroupName, &info.PrizeName, &info.PrizeResultMethod, &info.PrizeResult,
			&info.HowToParticipate, &info.Participate, &info.KeyWord, &info.PrizesList,
			&info.TimeOfWinners, &allPrizesJSON, &choosePrizesJSON,
			&info.PrizeCount, &info.NumberOfWinners, &info.OpenStatus, &info.CancelStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("scan events ERROR: %v", err)
		}

		// 反序列化奖品列表和选择的奖品
		if err = json.Unmarshal([]byte(allPrizesJSON), &info.AllPrizes); err != nil {
			return nil, fmt.Errorf("json unmarshal allPrizes ERROR: %v", err)
		}
		if err = json.Unmarshal([]byte(choosePrizesJSON), &info.ChoosePrizes); err != nil {
			return nil, fmt.Errorf("json unmarshal choosePrizes ERROR: %v", err)
		}

		onEvent = append(onEvent, info)
	}

	sort.Slice(onEvent, func(i, j int) bool {
		return onEvent[i].ID < onEvent[j].ID
	})

	return onEvent, nil
}

// 加载已取消的所有活动信息
func loadCancelEvents(db *sql.DB) (cancelEvent []EventInformation, err error) {
	rows, err := db.Query("SELECT * FROM events WHERE open_status = 0 AND cancel_status = 1")
	if err != nil {
		return nil, fmt.Errorf("loadCreateInformation ERROR: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("rows.Close ERROR: %v", err)
		}
	}()

	for rows.Next() {
		var info EventInformation
		var allPrizesJSON, choosePrizesJSON string

		err = rows.Scan(
			&info.ID, &info.GroupName, &info.PrizeName, &info.PrizeResultMethod, &info.PrizeResult,
			&info.HowToParticipate, &info.Participate, &info.KeyWord, &info.PrizesList,
			&info.TimeOfWinners, &allPrizesJSON, &choosePrizesJSON,
			&info.PrizeCount, &info.NumberOfWinners, &info.OpenStatus, &info.CancelStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("scan events ERROR: %v", err)
		}

		// 反序列化奖品列表和选择的奖品
		if err = json.Unmarshal([]byte(allPrizesJSON), &info.AllPrizes); err != nil {
			return nil, fmt.Errorf("json unmarshal allPrizes ERROR: %v", err)
		}
		if err = json.Unmarshal([]byte(choosePrizesJSON), &info.ChoosePrizes); err != nil {
			return nil, fmt.Errorf("json unmarshal choosePrizes ERROR: %v", err)
		}

		cancelEvent = append(cancelEvent, info)
	}

	sort.Slice(cancelEvent, func(i, j int) bool {
		return cancelEvent[i].ID < cancelEvent[j].ID
	})

	return cancelEvent, nil
}

// 检查特定活动 ID 的数据
func checkEventInformationFromId(db *sql.DB, id string) (info EventInformation, err error) {
	var allPrizesJSON, choosePrizesJSON string

	stmt, err := db.Prepare("SELECT * FROM events WHERE id = ?")
	if err != nil {
		return EventInformation{}, fmt.Errorf("checkCreateInformation ERROR: %v", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("stmt.Close ERROR: %v", err)
		}
	}()

	err = stmt.QueryRow(id).Scan(
		&info.ID, &info.GroupName, &info.PrizeName, &info.PrizeResultMethod, &info.PrizeResult,
		&info.HowToParticipate, &info.Participate, &info.KeyWord, &info.PrizesList,
		&info.TimeOfWinners, &allPrizesJSON, &choosePrizesJSON,
		&info.PrizeCount, &info.NumberOfWinners, &info.OpenStatus, &info.CancelStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EventInformation{}, fmt.Errorf("event id does not exist")
		}
		return EventInformation{}, fmt.Errorf("checkCreateInformation ERROR: %v", err)
	}

	// 反序列化奖品列表和选择的奖品
	if err = json.Unmarshal([]byte(allPrizesJSON), &info.AllPrizes); err != nil {
		return EventInformation{}, fmt.Errorf("json Unmarshal allPrize err: %v", err)
	}
	if err = json.Unmarshal([]byte(choosePrizesJSON), &info.ChoosePrizes); err != nil {
		return EventInformation{}, fmt.Errorf("json Unmarshal choosePrize err: %v", err)
	}

	return info, nil
}
