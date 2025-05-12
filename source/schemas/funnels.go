package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type FunnelStage struct {
	Name         string          `json:"name,omitempty" bson:"name,omitempty"`
	RelatedLeads []bson.ObjectID `json:"related_leads,omitempty" bson:"related_leads,omitempty"`
}

type Funnel struct {
	ID        bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name      string        `json:"name,omitempty" bson:"name,omitempty"`
	Type      string        `json:"type,omitempty" bson:"type,omitempty"`
	Stages    []FunnelStage `json:"stages,omitempty" bson:"stages,omitempty"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at,omitempty"`
}
