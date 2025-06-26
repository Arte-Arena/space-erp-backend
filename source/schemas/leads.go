package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Lead struct {
	ID             bson.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Name           string          `json:"name,omitempty" bson:"name,omitempty"`
	Nickname       string          `json:"nickname,omitempty" bson:"nickname,omitempty"`
	Phone          string          `json:"phone,omitempty" bson:"phone,omitempty"`
	Type           string          `json:"type,omitempty" bson:"type,omitempty"`
	Segment        string          `json:"segment,omitempty" bson:"segment,omitempty"`
	CreatedAt      time.Time       `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt      time.Time       `json:"updated_at" bson:"updated_at,omitempty"`
	Status         string          `json:"status,omitempty" bson:"status,omitempty"`
	Source         string          `json:"source,omitempty" bson:"source,omitempty"`
	PlatformId     string          `json:"platform_id,omitempty" bson:"platform_id,omitempty"`
	RelatedBudgets []bson.ObjectID `json:"related_budgets,omitempty" bson:"related_budgets,omitempty"`
	RelatedOrders  []bson.ObjectID `json:"related_orders,omitempty" bson:"related_orders,omitempty"`
	RelatedClient  bson.ObjectID   `json:"related_client,omitempty" bson:"related_client,omitempty"`
	Rating         string          `json:"rating,omitempty" bson:"rating,omitempty"`
	Notes          string          `json:"notes,omitempty" bson:"notes,omitempty"`
	Responsible    bson.ObjectID   `json:"responsible,omitempty" bson:"responsible,omitempty"`
	UnlinkClient   bool            `json:"unlink_client,omitempty" bson:"-"`
	Blocked        bool            `json:"blocked,omitempty" bson:"blocked,omitempty"`
}

type LeadTier struct {
	ID        bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Label     string        `json:"label" bson:"label"`
	MinValue  float64       `json:"min_value" bson:"min_value"`
	MaxValue  float64       `json:"max_value" bson:"max_value"`
	Icon      string        `json:"icon" bson:"icon"`
	SumType   string        `json:"sum_type" bson:"sum_type"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at"`
}
