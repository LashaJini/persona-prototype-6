package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wholesome-ghoul/persona-prototype-6/constants"
	"github.com/wholesome-ghoul/persona-prototype-6/reddit"
)

var ErrColumnNotExists = errors.New("column does not exist")

const ACTIVITY_MULTIPLIER = 100
const LIMIT_NUMBER_OF_USERS = 100
const MAX_DATE_DIFF_SCORE = 10
const PENALTY = -100

type Persona struct {
	ID                   uuid.UUID
	Username             string
	SocialNetID          int16
	LastScannedAt        int64
	FirstScannedAt       int64
	TimesScanned         int64
	LatestActivityAt     int64
	PriorityScore        int64
	TotalContentCount    int
	AverageSpam          float32
	AverageEmotions      jsonb
	AveragePersonalities jsonb
	Thirdparty           *reddit.User
}

func NewPersona() *Persona {
	return &Persona{
		Thirdparty: reddit.NewUser(),
	}
}

func (p *Persona) IsGold() bool      { return p.Thirdparty.IsGold() }
func (p *Persona) IsEmployee() bool  { return p.Thirdparty.IsEmployee() }
func (p *Persona) IsSuspended() bool { return p.Thirdparty.IsSuspended() }
func (p *Persona) IsVerified() bool  { return p.Thirdparty.IsVerified() }
func (p *Persona) TotalKarma() int   { return p.Thirdparty.TotalKarma() }

func (p *Persona) Columns() map[string]interface{} {
	columns := map[string]interface{}{
		"persona_id":            p.ID,
		"username":              p.Username,
		"social_net_id":         p.SocialNetID,
		"last_scanned_at":       p.LastScannedAt,
		"first_scanned_at":      p.FirstScannedAt,
		"times_scanned":         p.TimesScanned,
		"latest_activity_at":    p.LatestActivityAt,
		"priority_score":        p.PriorityScore,
		"total_content_count":   p.TotalContentCount,
		"average_spam":          p.AverageSpam,
		"average_emotions":      p.AverageEmotions,
		"average_personalities": p.AveragePersonalities,
		"thirdparty":            p.Thirdparty,
	}
	return columns
}

type PersonaWithActivitiesMap map[string]PersonaWithActivities

func SelectByRedisScore(db *sql.DB, weekday int8, excludedUsernames ...string) (PersonaWithActivitiesMap, error) {
	if weekday < int8(time.Sunday) || weekday > int8(time.Saturday) {
		return nil, errors.New("invalid weekday")
	}
	psqlWeekdayArrayIndex := weekday + 1

	var whereQuery string
	for _, excludedUsername := range excludedUsernames {
		whereQuery += fmt.Sprintf("\t'%s',\n", excludedUsername)
	}
	whereQuery = strings.TrimSuffix(whereQuery, ",\n")

	if len(excludedUsernames) > 0 {
		whereQuery = fmt.Sprintf(`
WHERE username NOT IN (
%s
)`, whereQuery)
	}

	withQuery := fmt.Sprintf(`
WITH days_since_last_scanned AS (
  SELECT
    p.persona_id,
    (EXTRACT(epoch FROM CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::BIGINT - p.last_scanned_at) / %d as days_since
  FROM
    persona p
)
`, constants.DAYS_SINCE_LAST_SCAN)

	probsQuery := fmt.Sprintf(`%d * w.sun_to_mon_probs[%d]`, ACTIVITY_MULTIPLIER, psqlWeekdayArrayIndex)

	caseQuery := fmt.Sprintf(`
CASE
	WHEN p.last_scanned_at = 0 THEN 10
	WHEN days_since = 0 THEN %d
	ELSE days_since
END
`, PENALTY)

	q := fmt.Sprintf(`
%s
SELECT 
  p.persona_id,
	p.username,
	p.social_net_id,
	p.last_scanned_at,
	p.first_scanned_at,
	p.times_scanned,
	p.latest_activity_at,
	p.priority_score,
	p.total_content_count,
	p.average_spam,
	p.average_emotions,
	p.average_personalities,
	p.thirdparty,

	y.yearly_activity_id,
	y.yearly_activity,

	w.weekly_activity_id,
	w.sun_to_mon_probs,
	w.sun_to_mon_counts,

  p.priority_score
	  + %s
		+ %s
    AS redis_score
FROM persona p
INNER JOIN weekly_activity w USING(persona_id)
INNER JOIN yearly_activity y USING(persona_id)
INNER JOIN days_since_last_scanned d USING(persona_id)
 %s
ORDER BY
  redis_score DESC
LIMIT %d
`,
		withQuery,
		probsQuery,
		caseQuery,
		whereQuery,
		LIMIT_NUMBER_OF_USERS,
	)

	rows, err := db.Query(q)
	if err != nil {
		fmt.Println(q)
		return nil, err
	}
	defer rows.Close()

	personaWithActivitiesMap := make(PersonaWithActivitiesMap)
	for rows.Next() {
		persona := &Persona{}
		weeklyActivity := &WeeklyActivity{}
		yearlyActivity := &YearlyActivity{}
		var redisScore float64

		err := rows.Scan(
			&persona.ID,
			&persona.Username,
			&persona.SocialNetID,
			&persona.LastScannedAt,
			&persona.FirstScannedAt,
			&persona.TimesScanned,
			&persona.LatestActivityAt,
			&persona.PriorityScore,
			&persona.TotalContentCount,
			&persona.AverageSpam,
			&persona.AverageEmotions,
			&persona.AveragePersonalities,
			&persona.Thirdparty,

			&yearlyActivity.ID,
			&yearlyActivity.YearlyActivity,

			&weeklyActivity.ID,
			&weeklyActivity.Sun2MonProbs,
			&weeklyActivity.Sun2MonCounts,

			&redisScore,
		)
		if err != nil {
			fmt.Println("error in redis score calculation SCAN", err)
			continue
		}

		personaWithActivitiesMap[persona.Username] = PersonaWithActivities{
			persona,
			weeklyActivity,
			yearlyActivity,
			redisScore,
		}
	}

	return personaWithActivitiesMap, nil
}

func (p *Persona) UpdatePersona(db *sql.DB, includedColumns ...string) error {
	if len(includedColumns) == 0 {
		return nil
	}

	tableColumns := p.Columns()
	values := []interface{}{}
	for _, column := range includedColumns {
		if _, ok := tableColumns[column]; !ok {
			return ErrColumnNotExists
		}

		values = append(values, tableColumns[column])
	}
	values = append(values, p.ID)

	query := ""
	i := 1
	for _, column := range includedColumns {
		query += fmt.Sprintf("%s = $%d,", column, i)
		i++
	}
	query = strings.TrimSuffix(query, ",")

	q := fmt.Sprintf(`
UPDATE persona
SET %s
WHERE persona_id = $%d
`, query, i)

	_, err := db.Exec(q, values...)
	return err
}

func FindPersonaByID(db *sql.DB, personaID uuid.UUID) (*Persona, error) {
	return findPersona(db, "WHERE persona_id = $1", personaID)
}

func findPersona(db *sql.DB, query string, args ...interface{}) (*Persona, error) {
	persona := &Persona{}
	err := db.QueryRow(`
SELECT
	persona_id,
	username,
	social_net_id,
	last_scanned_at,
	first_scanned_at,
	times_scanned,
	latest_activity_at,
	priority_score,
	total_content_count,
	average_spam,
	average_emotions,
	average_personalities,
	thirdparty
FROM persona
`+query, args...).Scan(
		&persona.ID,
		&persona.Username,
		&persona.SocialNetID,
		&persona.LastScannedAt,
		&persona.FirstScannedAt,
		&persona.TimesScanned,
		&persona.LatestActivityAt,
		&persona.PriorityScore,
		&persona.TotalContentCount,
		&persona.AverageSpam,
		&persona.AverageEmotions,
		&persona.AveragePersonalities,
		&persona.Thirdparty,
	)

	return persona, err
}

func InsertPersonas(db *sql.DB, usernames []string, socialNetName string) (map[string]string, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	socialNetID := -1
	err = tx.QueryRow(`
SELECT social_net_id
FROM social_net
WHERE name = $1
`, socialNetName).Scan(&socialNetID)
	if err != nil {
		return nil, err
	}

	query := ""
	for i := range usernames {
		query += fmt.Sprintf("\t($%d, %d),\n", i+1, socialNetID)
	}
	query = strings.TrimSuffix(query, ",\n")

	q := fmt.Sprintf(`
INSERT INTO persona (
	username,
	social_net_id
)
VALUES %s
ON CONFLICT DO NOTHING
RETURNING persona_id
`, query)

	var anyUsernames []interface{}
	for _, username := range usernames {
		anyUsernames = append(anyUsernames, username)
	}
	rows, err := tx.Query(q, anyUsernames...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	personaIDsMap := make(map[string]string)
	ithUsername := -1
	for rows.Next() {
		ithUsername++
		var personaID string
		if err := rows.Scan(&personaID); err != nil {
			continue
		}

		personaIDsMap[usernames[ithUsername]] = personaID
	}

	query = ""
	i := 1
	var uniquePersonaIDs []interface{}
	for _, personaID := range personaIDsMap {
		query += fmt.Sprintf("\t($%d),\n", i)
		uniquePersonaIDs = append(uniquePersonaIDs, personaID)
		i++
	}
	query = strings.TrimSuffix(query, ",\n")

	if len(uniquePersonaIDs) > 0 {
		q := fmt.Sprintf(`
INSERT INTO weekly_activity (persona_id)
VALUES %s
`, query)

		weeklyStmt, err := tx.Prepare(q)
		if err != nil {
			return personaIDsMap, err
		}
		defer weeklyStmt.Close()

		_, err = weeklyStmt.Exec(uniquePersonaIDs...)
		if err != nil {
			return personaIDsMap, err
		}

		q = fmt.Sprintf(`
INSERT INTO yearly_activity (persona_id)
VALUES %s
`, query)

		yearlyStmt, err := tx.Prepare(q)
		if err != nil {
			return personaIDsMap, err
		}
		defer yearlyStmt.Close()

		_, err = yearlyStmt.Exec(uniquePersonaIDs...)
		if err != nil {
			return personaIDsMap, err
		}
	}

	if err = tx.Commit(); err != nil {
		return personaIDsMap, err
	}

	return personaIDsMap, err
}

func InsertPersona(db *sql.DB, username, socialNetName string) (string, error) {
	var personaID string
	tx, err := db.Begin()
	if err != nil {
		return personaID, err
	}
	defer tx.Rollback()

	socialNetID := -1
	err = tx.QueryRow(`
SELECT social_net_id
FROM social_net
WHERE name = $1
`, socialNetName).Scan(&socialNetID)
	if err != nil {
		return personaID, err
	}

	err = tx.QueryRow(`
INSERT INTO persona (
	username,
	social_net_id
)
VALUES (
	$1, $2
)
RETURNING persona_id
`, username, socialNetID).Scan(&personaID)
	if err != nil {
		return personaID, err
	}

	_, err = tx.Exec(`
INSERT INTO weekly_activity (persona_id)
VALUES ($1)
`, personaID)
	if err != nil {
		return personaID, err
	}

	_, err = tx.Exec(`
INSERT INTO yearly_activity (persona_id)
VALUES ($1)
`, personaID)
	if err != nil {
		return personaID, err
	}

	if err = tx.Commit(); err != nil {
		return personaID, err
	}

	return personaID, nil
}

func DeleteAllPersona(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM persona")

	return err
}
