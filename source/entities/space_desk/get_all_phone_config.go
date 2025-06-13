package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PhoneSettings struct {
	Type      string        `bson:"type" json:"type"`
	Phones    []PhoneConfig `bson:"phones" json:"phones"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
}

func GetAllPhoneConfig(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	collection := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CONFIG)

	// Busca o documento "global" de configurações
	filter := bson.M{"type": "global"}
	var settings PhoneSettings
	err = collection.FindOne(ctx, filter).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.SendResponse(w, http.StatusNotFound, "", nil, 0)
			return
		}
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", settings, 0)
}
