package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	REPORT_TYPE_CLIENTS = "clients"
	REPORT_TYPE_BUDGETS = "budgets"
	REPORT_TYPE_LEADS   = "leads"
	REPORT_TYPE_ORDERS  = "orders"

	REPORT_GOAL_TYPE_MONTHLY = "monthly"
	REPORT_GOAL_TYPE_DAILY   = "daily"
	REPORT_GOAL_TYPE_YEARLY  = "yearly"

	REPORT_GOAL_RELATED_BUDGETS = "budgets"
	REPORT_GOAL_RELATED_ORDERS  = "orders"
	REPORT_GOAL_RELATED_CLIENTS = "clients"
	REPORT_GOAL_RELATED_LEADS   = "leads"
)

type ReportCommercialGoals struct {
	ID          bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string        `json:"name,omitempty" bson:"name,omitempty"`
	GoalType    string        `json:"goal_type,omitempty" bson:"goal_type,omitempty"`
	RelatedTo   string        `json:"related_to,omitempty" bson:"related_to,omitempty"`
	TargetValue float64       `json:"target_value,omitempty" bson:"target_value,omitempty"`
	CreatedAt   time.Time     `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt   time.Time     `json:"updated_at" bson:"updated_at,omitempty"`
}
