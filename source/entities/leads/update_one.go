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

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func UpdateOne(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_LEAD_ID_FORMAT)
		return
	}

	lead := &schemas.Lead{}
	if err := json.NewDecoder(r.Body).Decode(&lead); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.LEADS_INVALID_REQUEST_DATA)
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

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_LEADS)

	filter := bson.D{{Key: "_id", Value: id}}

	updateDoc := bson.D{}

	if lead.Name != "" {
		updateDoc = append(updateDoc, bson.E{Key: "name", Value: lead.Name})
	}
	if lead.Nickname != "" {
		updateDoc = append(updateDoc, bson.E{Key: "nickname", Value: lead.Nickname})
	}
	if lead.Phone != "" {
		updateDoc = append(updateDoc, bson.E{Key: "phone", Value: lead.Phone})
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
	if lead.Rating != "" {
		updateDoc = append(updateDoc, bson.E{Key: "classification", Value: lead.Rating})
	}
	if lead.Notes != "" {
		updateDoc = append(updateDoc, bson.E{Key: "notes", Value: lead.Notes})
	}
	if len(lead.RelatedBudgets) > 0 {
		updateDoc = append(updateDoc, bson.E{Key: "related_budgets", Value: lead.RelatedBudgets})
	}
	if len(lead.RelatedOrders) > 0 {
		updateDoc = append(updateDoc, bson.E{Key: "related_orders", Value: lead.RelatedOrders})
	}
	if !lead.RelatedClient.IsZero() {
		updateDoc = append(updateDoc, bson.E{Key: "related_client", Value: lead.RelatedClient})
	}
	if !lead.Responsible.IsZero() {
		updateDoc = append(updateDoc, bson.E{Key: "responsible", Value: lead.Responsible})
	}

	updateDoc = append(updateDoc, bson.E{Key: "updated_at", Value: time.Now()})

	if len(updateDoc) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "Nenhum campo para atualizar foi fornecido", nil, 0)
		return
	}

	update := bson.D{{Key: "$set", Value: updateDoc}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_UPDATE_LEAD_IN_MONGODB)
		return
	}

	if result.MatchedCount == 0 {
		utils.SendResponse(w, http.StatusNotFound, "Lead não encontrado", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", nil, 0)
}
