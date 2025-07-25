package funnelsplacements

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetOne(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, INVALID_FUNNEL_PLACEMENT_ID_FORMAT)
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

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_FUNNELS_PLACEMENTS)

	var funnelPlacement schemas.FunnelPlacement
	err = collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&funnelPlacement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.SendResponse(w, http.StatusNotFound, "Posicionamento de funil não encontrado", nil, 0)
			return
		}
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, CANNOT_FIND_FUNNEL_PLACEMENT_BY_ID_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", funnelPlacement, 0)
}
