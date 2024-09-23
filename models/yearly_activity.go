package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/wholesome-ghoul/persona-prototype-6/constants"
	"github.com/wholesome-ghoul/persona-prototype-6/storage/database"
)

const HOURS_IN_A_DAY = constants.HOURS_IN_A_DAY
const DAYS_IN_A_YEAR = constants.DAYS_IN_A_YEAR

type YearsByHours [DAYS_IN_A_YEAR][HOURS_IN_A_DAY]int

func (y *YearsByHours) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		strValue := string(src)
		// Remove leading and trailing '{' and '}' characters
		strValue = strings.Trim(strValue, "{}")

		// Split the string into rows
		rowStrs := strings.Split(strValue, "},{")
		if len(rowStrs) != DAYS_IN_A_YEAR {
			return errors.New("invalid number of rows")
		}

		// Parse each row into integers and populate the YearlyActivity array
		for i, rowStr := range rowStrs {
			colStrs := strings.Split(rowStr, ",")
			if len(colStrs) != HOURS_IN_A_DAY {
				return fmt.Errorf("invalid number of columns in row %d", i)
			}

			for j, colStr := range colStrs {
				value, err := strconv.Atoi(colStr)
				if err != nil {
					return fmt.Errorf("failed to parse integer in row %d, column %d: %w", i, j, err)
				}
				y[i][j] = value
			}
		}
		return nil
	default:
		return errors.New("unsupported source type")
	}
}

type YearlyActivity struct {
	ID             int
	PersonaID      uuid.UUID
	YearlyActivity YearsByHours
}

func NewYearlyActivity() *YearlyActivity {
	return &YearlyActivity{}
}

func (y *YearlyActivity) UpdateYearlyActivityTx(tx *database.MultiInstruction) error {
	_, err := tx.Exec(`
UPDATE yearly_activity
SET
	yearly_activity = $1
WHERE yearly_activity_id = $2
`, pq.Array(y.YearlyActivity), y.ID)

	return err
}
