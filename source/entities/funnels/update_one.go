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

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func UpdateOne(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.INVALID_FUNNEL_ID_FORMAT),
		})
		return
	}

	funnel := &schemas.Funnel{}
	if err := json.NewDecoder(r.Body).Decode(&funnel); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.FUNNELS_INVALID_REQUEST_DATA),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CANNOT_CONNECT_TO_MONGODB),
		})
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_FUNNELS)

	filter := bson.D{{Key: "_id", Value: id}}

	updateDoc := bson.D{}

	if funnel.Name != "" {
		updateDoc = append(updateDoc, bson.E{Key: "name", Value: funnel.Name})
	}
	if funnel.Type != "" {
		updateDoc = append(updateDoc, bson.E{Key: "type", Value: funnel.Type})
	}
	if len(funnel.Stages) > 0 {
		updateDoc = append(updateDoc, bson.E{Key: "stages", Value: funnel.Stages})
	}

	updateDoc = append(updateDoc, bson.E{Key: "updated_at", Value: time.Now()})

	if len(updateDoc) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Nenhum campo para atualizar foi fornecido",
		})
		return
	}

	update := bson.D{{Key: "$set", Value: updateDoc}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CANNOT_UPDATE_FUNNEL_IN_MONGODB),
		})
		return
	}

	if result.MatchedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Funil não encontrado",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
}
