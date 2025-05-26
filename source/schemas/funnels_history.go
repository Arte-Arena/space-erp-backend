package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type FunnelsHistory struct {
	ID              bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	RelatedUserName bson.ObjectID `json:"related_user,omitempty" bson:"related_user,omitempty"`
	RelatedFunnel   bson.ObjectID `json:"related_funnel,omitempty" bson:"related_funnel,omitempty"`
	Action          string        `json:"action,omitempty" bson:"action,omitempty"`
	CreatedAt       time.Time     `json:"created_at" bson:"created_at,omitempty"`
}
