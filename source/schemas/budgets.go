package schemas

import "go.mongodb.org/mongo-driver/v2/bson"

type Budget struct {
	ID    bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	OldID uint64        `json:"old_id" bson:"old_id"`
}
