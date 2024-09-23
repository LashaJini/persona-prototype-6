package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/wholesome-ghoul/persona-prototype-6/reddit"
	"github.com/wholesome-ghoul/persona-prototype-6/storage/database"
)

type Content[V reddit.PostOrComment] struct {
	ID                  int
	ContentTypeID       int16
	PersonaID           uuid.UUID
	Spam                float32
	Emotions            jsonb
	Personalities       jsonb
	ThirdpartyURL       string
	ThirdpartyCreatedAt int64
	Thirdparty          V
}

func (c *Content[V]) Columns() map[string]interface{} {
	columns := map[string]interface{}{
		"content_id":            c.ID,
		"content_type_id":       c.ContentTypeID,
		"persona_id":            c.PersonaID,
		"spam":                  c.Spam,
		"emotions":              c.Emotions,
		"personalities":         c.Personalities,
		"thirdparty_url":        c.ThirdpartyURL,
		"thirdparty_created_at": c.ThirdpartyCreatedAt,
		"thirdparty":            c.Thirdparty,
	}
	return columns
}

func InsertContentsTx[V reddit.PostOrComment](tx *database.MultiInstruction, personaID string, contents []V, spamProbs []float32) ([]string, error) {
	var contentIDs []string
	if len(contents) == 0 {
		return contentIDs, nil
	}

	if len(contents) != len(spamProbs) {
		return contentIDs, errors.New("contents and spamProbs should have same length")
	}

	contentTypeID, err := findContentIDByNameTx(tx, contents[0].ContentTypeName())
	if err != nil {
		return contentIDs, err
	}

	columns := []string{
		"content_type_id",
		"persona_id",
		"spam",
		"thirdparty_url",
		"thirdparty_created_at",
		"thirdparty",
	}
	numOfColumnsToInsert := len(columns)
	columnsQuery := ""
	for _, column := range columns {
		columnsQuery += fmt.Sprintf("%s,\n", column)
	}
	columnsQuery = strings.TrimSuffix(columnsQuery, ",\n")

	valuesQuery := ""
	i := 1
	values := []interface{}{}
	for index, content := range contents {
		valuesQuery += "("
		for j := i; j < i+numOfColumnsToInsert; j++ {
			valuesQuery += fmt.Sprintf("$%d, ", j)
		}
		valuesQuery = strings.TrimSuffix(valuesQuery, ", ")
		valuesQuery += "),\n"

		values = append(
			values,
			contentTypeID,
			personaID,
			spamProbs[index],
			content.ThirdpartyURL(),
			content.CreatedAtEpoch(),
			content,
		)

		i += numOfColumnsToInsert
	}
	valuesQuery = strings.TrimSuffix(valuesQuery, ",\n")

	q := fmt.Sprintf(`
INSERT INTO content (
	%s
)
VALUES
	%s
RETURNING content_id
`, columnsQuery, valuesQuery)

	rows, err := tx.Query(q, values...)
	if err != nil {
		return contentIDs, err
	}
	defer rows.Close()

	for rows.Next() {
		var contentID string
		_ = rows.Scan(&contentID)
		contentIDs = append(contentIDs, contentID)
	}

	return contentIDs, nil
}
