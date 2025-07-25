package funnelsplacements

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateOne(w http.ResponseWriter, r *http.Request) {
	funnelPlacement := &schemas.FunnelPlacement{}
	if err := json.NewDecoder(r.Body).Decode(&funnelPlacement); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, FUNNELS_PLACEMENTS_INVALID_REQUEST_DATA)
		return
	}

	funnelPlacement.CreatedAt = time.Now()
	funnelPlacement.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_FUNNELS_PLACEMENTS)

	_, err = collection.InsertOne(ctx, funnelPlacement)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, CANNOT_INSERT_FUNNEL_PLACEMENT_TO_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusCreated, "", funnelPlacement, 0)
}
