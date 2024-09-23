package models

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/wholesome-ghoul/persona-prototype-6/constants"
	"github.com/wholesome-ghoul/persona-prototype-6/storage/database"
)

const DAYS_IN_A_WEEK = 7

type WeekdaysFloat [DAYS_IN_A_WEEK]float64
type WeekdaysInt [DAYS_IN_A_WEEK]int

type WeeklyActivity struct {
	ID            int
	PersonaID     uuid.UUID
	Sun2MonProbs  WeekdaysFloat
	Sun2MonCounts WeekdaysInt
}

func NewWeeklyActivity() *WeeklyActivity {
	return &WeeklyActivity{}
}

func (w *WeekdaysFloat) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		strValue := string(src)
		strValue = strings.Trim(strValue, "{}")
		values := strings.Split(strValue, ",")
		for i, value := range values {
			w[i], _ = strconv.ParseFloat(value, 64)
		}
		return nil
	default:
		return errors.New("unsupported source type")
	}
}

func (w *WeekdaysInt) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		strValue := string(src)
		strValue = strings.Trim(strValue, "{}")
		values := strings.Split(strValue, ",")
		for i, value := range values {
			w[i], _ = strconv.Atoi(value)
		}
		return nil
	default:
		return errors.New("unsupported source type")
	}
}

func (w *WeeklyActivity) CalculateProbs(startEpoch, endEpoch int64) {
	if endEpoch-startEpoch < constants.EPOCH_DIFF_THRESHOLD {
		return
	}
	var weekdayCounts = [DAYS_IN_A_WEEK]int{}

	startDate := time.Unix(startEpoch, 0).UTC()
	endDate := time.Unix(endEpoch, 0).UTC()

	for date := startDate; date.Before(endDate); date = date.AddDate(0, 0, 1) {
		weekdayCounts[date.Weekday()]++
	}

	for i, totalWeekdayCount := range weekdayCounts {
		if totalWeekdayCount == 0 {
			continue
		}

		activeWeekdayCount := float64(w.Sun2MonCounts[i])
		newProb := activeWeekdayCount / float64(totalWeekdayCount)
		w.Sun2MonProbs[i] = newProb
	}
}

func (w *WeeklyActivity) UpdateWeeklyActivityTx(tx *database.MultiInstruction) error {
	_, err := tx.Exec(`
UPDATE weekly_activity
SET
	sun_to_mon_probs = $1,
	sun_to_mon_counts = $2
WHERE weekly_activity_id = $3
`, pq.Array(w.Sun2MonProbs), pq.Array(w.Sun2MonCounts), w.ID)

	return err
}
