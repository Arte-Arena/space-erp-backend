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
	ID         bson.ObjectID    `json:"_id,omitempty" bson:"_id,omitempty"`
	OldID      uint64           `json:"old_id" bson:"old_id"`
	Name       string           `json:"name" bson:"name"`
	Email      string           `json:"email" bson:"email"`
	Role       []string         `json:"role" bson:"role"`
	Commission []CommissionRule `json:"commission" bson:"commission"`
	CreatedAt  time.Time        `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at" bson:"updated_at"`
	Goals      []IndividualGoal `json:"goals" bson:"goals"`
}

type GoalType string

const (
	GoalTypeBudgets GoalType = "budgets"
	GoalTypeOrders  GoalType = "orders"
	GoalTypeLeads   GoalType = "leads"
)

type GoalPeriod string

const (
	GoalPeriodDaily   GoalPeriod = "daily"
	GoalPeriodWeekly  GoalPeriod = "weekly"
	GoalPeriodMonthly GoalPeriod = "monthly"
)

type IndividualGoal struct {
	Name   string     `json:"name" bson:"name"`
	Type   GoalType   `json:"type" bson:"type"`
	Value  uint64     `json:"value" bson:"value"`
	Period GoalPeriod `json:"period" bson:"period"`
}
type CommissionRule struct {
	MinSales   uint64  `json:"min_sales" bson:"min_sales"`
	MaxSales   uint64  `json:"max_sales" bson:"max_sales"`
	Percentage float64 `json:"percentage" bson:"percentage"`
}
