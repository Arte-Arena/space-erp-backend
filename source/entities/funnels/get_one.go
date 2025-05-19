package funnels

import (
	"api/source/database"
	"api/source/schemas"
	"api/source/utils"
	"context"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetOne(w http.ResponseWriter, r *http.Request) {
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

	funnel := &schemas.Funnel{}
	err = collection.FindOne(ctx, filter).Decode(&funnel)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.SendResponse(w, http.StatusNotFound, "Funil não encontrado", nil, 0)
		} else {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_FUNNEL_BY_ID_IN_MONGODB)
		}
		return
	}

	utils.SendResponse(w, http.StatusOK, "", funnel, 0)
}
