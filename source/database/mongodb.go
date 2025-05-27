package database

import (
	"api/utils"
	"os"
	"time"
)

const (
	MONGO_TIMEOUT              = 20 * time.Second
	COLLECTION_LEADS           = "leads"
	COLLECTION_FUNNELS         = "funnels"
	COLLECTION_BUDGETS         = "budgets"
	COLLECTION_ORDERS          = "orders"
	COLLECTION_USERS           = "users"
	COLLECTION_CLIENTS         = "clients"
	COLLECTION_FUNNELS_HISTORY = "funnels_history"
)

func GetDB() string {
	environment := os.Getenv(utils.ENV)

	if environment == utils.ENV_RELEASE {
		return "production"
	}

	if environment == utils.ENV_HOMOLOG {
		return "homolog"
	}

	if environment == utils.ENV_DEVELOPMENT {
		return "development"
	}

	panic("[MongoDB] Invalid DB name")
}
