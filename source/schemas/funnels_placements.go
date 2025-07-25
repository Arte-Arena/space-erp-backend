package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type FunnelPlacement struct {
	ID          bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	FunnelID    bson.ObjectID `json:"funnel_id,omitempty" bson:"funnel_id,omitempty"`
	RelatedLead bson.ObjectID `json:"related_lead,omitempty" bson:"related_lead,omitempty"`
	StageName   string        `json:"stage_name,omitempty" bson:"stage_name,omitempty"`
	CreatedAt   time.Time     `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt   time.Time     `json:"updated_at" bson:"updated_at,omitempty"`
}
