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

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func UpdateOne(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, INVALID_FUNNEL_PLACEMENT_ID_FORMAT)
		return
	}

	funnelPlacement := &schemas.FunnelPlacement{}
	if err := json.NewDecoder(r.Body).Decode(&funnelPlacement); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, FUNNELS_PLACEMENTS_INVALID_REQUEST_DATA)
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

	filter := bson.D{{Key: "_id", Value: id}}

	updateDoc := bson.D{}

	if !funnelPlacement.FunnelID.IsZero() {
		updateDoc = append(updateDoc, bson.E{Key: "funnel_id", Value: funnelPlacement.FunnelID})
	}
	if !funnelPlacement.RelatedLead.IsZero() {
		updateDoc = append(updateDoc, bson.E{Key: "related_lead", Value: funnelPlacement.RelatedLead})
	}
	if funnelPlacement.StageName != "" {
		updateDoc = append(updateDoc, bson.E{Key: "stage_name", Value: funnelPlacement.StageName})
	}

	updateDoc = append(updateDoc, bson.E{Key: "updated_at", Value: time.Now()})

	if len(updateDoc) == 1 {
		utils.SendResponse(w, http.StatusBadRequest, "Nenhum campo para atualizar foi fornecido", nil, 0)
		return
	}

	update := bson.D{{Key: "$set", Value: updateDoc}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, CANNOT_UPDATE_FUNNEL_PLACEMENT_IN_MONGODB)
		return
	}

	if result.MatchedCount == 0 {
		utils.SendResponse(w, http.StatusNotFound, "Posicionamento de funil não encontrado", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", nil, 0)
}
