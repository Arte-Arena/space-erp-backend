package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type LeadTier struct {
	ID        bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Label     string        `json:"label" bson:"label"`
	MinValue  float64       `json:"min_value" bson:"min_value"`
	MaxValue  float64       `json:"max_value" bson:"max_value"`
	Icon      string        `json:"icon" bson:"icon"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at"`
}
