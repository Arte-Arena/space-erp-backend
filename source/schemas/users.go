package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	USERS_ROLE_SUPER_ADMIN          = "super_admin"
	USERS_ROLE_IT                   = "it"
	USERS_ROLE_ADMIN                = "admin"
	USERS_ROLE_LEADER               = "leader"
	USERS_ROLE_COLLABORATOR         = "collaborator"
	USERS_ROLE_DESIGNER             = "designer"
	USERS_ROLE_DESIGNER_COORDINATOR = "designer_coordinator"
	USERS_ROLE_PRODUCTION           = "production"
	USERS_ROLE_COMMERCIAL           = "commercial"
)

type User struct {
	ID        bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	OldID     uint64        `json:"old_id" bson:"old_id"`
	Name      string        `json:"name" bson:"name"`
	Email     string        `json:"email" bson:"email"`
	Role      []string      `json:"role" bson:"role"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at"`
}
