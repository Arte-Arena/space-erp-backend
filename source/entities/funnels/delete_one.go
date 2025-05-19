package funnels

import (
	"api/source/database"
	"api/source/utils"
	"context"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DeleteOne(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_FUNNEL_ID_FORMAT)
		return
	}

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

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_FUNNELS)

	filter := bson.D{{Key: "_id", Value: id}}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_DELETE_FUNNEL_FROM_MONGODB)
		return
	}

	if result.DeletedCount == 0 {
		utils.SendResponse(w, http.StatusNotFound, "Funil não encontrado", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", nil, 0)
}
