package database

import (
	"api/utils"
	"os"
	"time"
)

const (
	MONGO_TIMEOUT                         = 20 * time.Second
	COLLECTION_LEADS                      = "leads"
	COLLECTION_LEADS_TIERS                = "leads_tiers"
	COLLECTION_FUNNELS                    = "funnels"
	COLLECTION_BUDGETS                    = "budgets"
	COLLECTION_ORDERS                     = "orders"
	COLLECTION_USERS                      = "users"
	COLLECTION_CLIENTS                    = "clients"
	COLLECTION_FUNNELS_HISTORY            = "funnels_history"
	COLLECTION_SPACE_DESK_EVENTS_WHATSAPP = "space_desk_events_whatsapp"
	COLLECTION_SPACE_DESK_CHAT            = "space_desk_chat"
	COLLECTION_SPACE_DESK_MESSAGE         = "space_desk_message"
	COLLECTION_SPACE_DESK_CONFIG          = "space_desk_config"
	COLLECTION_SPACE_DESK_READY_MESSAGE   = "ready_chat_messages"
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
