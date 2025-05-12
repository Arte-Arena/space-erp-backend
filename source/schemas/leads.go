package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Lead struct {
	ID             bson.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Name           string          `json:"name,omitempty" bson:"name,omitempty"`
	Phone          string          `json:"phone,omitempty" bson:"phone,omitempty"`
	Document       string          `json:"document,omitempty" bson:"document,omitempty"`
	Type           string          `json:"type,omitempty" bson:"type,omitempty"`
	Segment        string          `json:"segment,omitempty" bson:"segment,omitempty"`
	CreatedAt      time.Time       `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt      time.Time       `json:"updated_at" bson:"updated_at,omitempty"`
	Status         string          `json:"status,omitempty" bson:"status,omitempty"`
	Source         string          `json:"source,omitempty" bson:"source,omitempty"`
	UniqueID       string          `json:"unique_id,omitempty" bson:"unique_id,omitempty"`
	RelatedQuotes  []bson.ObjectID `json:"related_quotes,omitempty" bson:"related_quotes,omitempty"`
	RelatedOrders  []bson.ObjectID `json:"related_orders,omitempty" bson:"related_orders,omitempty"`
	Classification string          `json:"classification,omitempty" bson:"classification,omitempty"`
	Notes          string          `json:"notes,omitempty" bson:"notes,omitempty"`
}
