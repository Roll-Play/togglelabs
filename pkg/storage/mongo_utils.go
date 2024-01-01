package storage

import "go.mongodb.org/mongo-driver/bson/primitive"

type Timestamps struct {
	CreatedAt primitive.DateTime `json:"created_at" bson:"created_at"`
	UpadtedAt primitive.DateTime `json:"updated_at" bson:"updated_at"`
	DeletedAt primitive.DateTime `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}
