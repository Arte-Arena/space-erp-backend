package database

import (
	"api/source/utils"
	"os"
	"time"
)

const (
	MONGO_TIMEOUT      = 20 * time.Second
	COLLECTION_LEADS   = "leads"
	COLLECTION_FUNNELS = "funnels"
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
