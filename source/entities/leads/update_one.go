package leads

import (
	"api/source/database"
	"api/source/schemas"
	"api/source/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpdateOne(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.INVALID_LEAD_ID_FORMAT),
		})
		return
	}

	lead := &schemas.Lead{}
	if err := json.NewDecoder(r.Body).Decode(&lead); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.LEADS_INVALID_REQUEST_DATA),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(ctx, opts)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.CANNOT_CONNECT_TO_MONGODB),
		})
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_LEADS)

	filter := bson.D{{Key: "_id", Value: id}}

	updateDoc := bson.D{}

	if lead.Name != "" {
		updateDoc = append(updateDoc, bson.E{Key: "name", Value: lead.Name})
	}
	if lead.Phone != "" {
		updateDoc = append(updateDoc, bson.E{Key: "phone", Value: lead.Phone})
	}
	if lead.Document != "" {
		updateDoc = append(updateDoc, bson.E{Key: "document", Value: lead.Document})
	}
	if lead.Type != "" {
		updateDoc = append(updateDoc, bson.E{Key: "type", Value: lead.Type})
	}
	if lead.Segment != "" {
		updateDoc = append(updateDoc, bson.E{Key: "segment", Value: lead.Segment})
	}
	if lead.Status != "" {
		updateDoc = append(updateDoc, bson.E{Key: "status", Value: lead.Status})
	}
	if lead.Source != "" {
		updateDoc = append(updateDoc, bson.E{Key: "source", Value: lead.Source})
	}
	if lead.UniqueID != "" {
		updateDoc = append(updateDoc, bson.E{Key: "unique_id", Value: lead.UniqueID})
	}
	if lead.Classification != "" {
		updateDoc = append(updateDoc, bson.E{Key: "classification", Value: lead.Classification})
	}
	if lead.Notes != "" {
		updateDoc = append(updateDoc, bson.E{Key: "notes", Value: lead.Notes})
	}
	if len(lead.RelatedQuotes) > 0 {
		updateDoc = append(updateDoc, bson.E{Key: "related_quotes", Value: lead.RelatedQuotes})
	}
	if len(lead.RelatedOrders) > 0 {
		updateDoc = append(updateDoc, bson.E{Key: "related_orders", Value: lead.RelatedOrders})
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
			Message: utils.SendInternalError(utils.CANNOT_UPDATE_LEAD_IN_MONGODB),
		})
		return
	}

	if result.MatchedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: "Lead não encontrado",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
}
