package funnels

import (
	"api/source/database"
	"api/source/schemas"
	"api/source/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateOne(w http.ResponseWriter, r *http.Request) {
	funnel := &schemas.Funnel{}
	if err := json.NewDecoder(r.Body).Decode(&funnel); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.FUNNELS_INVALID_REQUEST_DATA)
		return
	}

	funnel.CreatedAt = time.Now()
	funnel.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(ctx, opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_FUNNELS)

	_, err = collection.InsertOne(ctx, funnel)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_FUNNEL_TO_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusCreated, "", nil, 0)
}
