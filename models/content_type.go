package models

import (
	"github.com/wholesome-ghoul/persona-prototype-6/storage/database"
)

const CONTENT_TYPE_COMMENT = "comment"
const CONTENT_TYPE_POST = "post"

type ContentType struct {
	ID   int16
	Name string
}

func findContentIDByNameTx(tx *database.MultiInstruction, name string) (int16, error) {
	var contentTypeID int16

	err := tx.QueryRow(`
SELECT content_type_id
FROM content_type
WHERE name = $1
`, name).Scan(&contentTypeID)

	return contentTypeID, err
}
