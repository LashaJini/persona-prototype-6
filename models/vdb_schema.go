package models

type ContentVDBSchema struct {
	ContentID  string
	Text       string
	Embeddings []float64
}

// func GetContentVDBSchema(collectionName string) *entity.Schema {
// 	schema := &entity.Schema{
// 		CollectionName: collectionName,
// 		Description:    "Content search",
// 		Fields: []*entity.Field{
// 			{
// 				Name:       "content_id",
// 				DataType:   entity.FieldTypeVarChar,
// 				PrimaryKey: true,
// 				AutoID:     false,
// 				TypeParams: map[string]string{
// 					"max_length": "64",
// 				},
// 			},
// 			{
// 				Name:     "text",
// 				DataType: entity.FieldTypeVarChar,
// 				TypeParams: map[string]string{
// 					"max_length": "2048",
// 				},
// 			},
// 			{
// 				Name:     "embedding",
// 				DataType: entity.FieldTypeFloatVector,
// 				TypeParams: map[string]string{
// 					"dim": "384",
// 				},
// 			},
// 		},
// 	}
//
// 	return schema
// }
