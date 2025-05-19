package funnels

import (
	"api/source/database"
	"api/source/schemas"
	"api/source/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func UpdateOneStage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_FUNNEL_ID_FORMAT)
		return
	}

	indexStr := r.PathValue("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "Índice de estágio inválido", nil, 0)
		return
	}

	stage := &schemas.FunnelStage{}
	if err := json.NewDecoder(r.Body).Decode(&stage); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.FUNNELS_INVALID_REQUEST_DATA)
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

	if index < 0 || index >= len(funnel.Stages) {
		utils.SendResponse(w, http.StatusBadRequest, "Índice de estágio fora dos limites", nil, 0)
		return
	}

	updateDoc := bson.D{}

	if stage.Name != "" {
		updateDoc = append(updateDoc, bson.E{Key: "stages." + indexStr + ".name", Value: stage.Name})
	}
	if len(stage.RelatedLeads) > 0 {
		updateDoc = append(updateDoc, bson.E{Key: "stages." + indexStr + ".related_leads", Value: stage.RelatedLeads})
	}

	updateDoc = append(updateDoc, bson.E{Key: "updated_at", Value: time.Now()})

	if len(updateDoc) == 1 {
		utils.SendResponse(w, http.StatusBadRequest, "Nenhum campo para atualizar foi fornecido", nil, 0)
		return
	}

	update := bson.D{{Key: "$set", Value: updateDoc}}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_UPDATE_FUNNEL_IN_MONGODB)
		return
	}

	if result.MatchedCount == 0 {
		utils.SendResponse(w, http.StatusNotFound, "Funil não encontrado", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", nil, 0)
}
