package funnels

import (
	"api/source/database"
	"api/source/schemas"
	"api/source/utils"
	"context"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAll(w http.ResponseWriter, r *http.Request) {
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

	filter := bson.D{}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_FUNNELS_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	funnels := []schemas.Funnel{}
	if err := cursor.All(ctx, &funnels); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_FUNNELS_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", funnels, 0)
}
