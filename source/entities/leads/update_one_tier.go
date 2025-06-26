package leads

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func UpdateOneTier(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_LEAD_ID_FORMAT)
		return
	}

	tier := &schemas.LeadTier{}
	if err := json.NewDecoder(r.Body).Decode(&tier); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.LEADS_INVALID_REQUEST_DATA)
		return
	}

	tier.Label = strings.TrimSpace(tier.Label)

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

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_LEADS_TIERS)

	if tier.Label != "" {
		labelFilter := bson.D{
			{Key: "label", Value: bson.D{{Key: "$regex", Value: "^" + strings.ReplaceAll(tier.Label, " ", "") + "$"}, {Key: "$options", Value: "i"}}},
			{Key: "_id", Value: bson.D{{Key: "$ne", Value: id}}},
		}
		cursor, err := collection.Find(ctx, labelFilter)
		if err == nil {
			var existing []schemas.LeadTier
			_ = cursor.All(ctx, &existing)
			if len(existing) > 0 {
				utils.SendResponse(w, http.StatusBadRequest, "Já existe um tier com esse label", nil, 0)
				return
			}
		}
	}

	valueOr := bson.A{}
	if tier.MinValue != 0 {
		valueOr = append(valueOr, bson.D{{Key: "min_value", Value: tier.MinValue}})
	}
	if tier.MaxValue != 0 {
		valueOr = append(valueOr, bson.D{{Key: "max_value", Value: tier.MaxValue}})
	}
	if len(valueOr) > 0 {
		valueFilter := bson.D{
			{Key: "$or", Value: valueOr},
			{Key: "_id", Value: bson.D{{Key: "$ne", Value: id}}},
		}
		cursor, err := collection.Find(ctx, valueFilter)
		if err == nil {
			existing := []schemas.LeadTier{}
			_ = cursor.All(ctx, &existing)
			if len(existing) > 0 {
				utils.SendResponse(w, http.StatusBadRequest, "Já existe um tier com esse min_value ou max_value", nil, 0)
				return
			}
		}
	}

	updateDoc := bson.D{}
	if tier.Label != "" {
		updateDoc = append(updateDoc, bson.E{Key: "label", Value: tier.Label})
	}
	if tier.MinValue != 0 {
		updateDoc = append(updateDoc, bson.E{Key: "min_value", Value: tier.MinValue})
	}
	if tier.MaxValue != 0 {
		updateDoc = append(updateDoc, bson.E{Key: "max_value", Value: tier.MaxValue})
	}
	if tier.Icon != "" {
		updateDoc = append(updateDoc, bson.E{Key: "icon", Value: tier.Icon})
	}
	updateDoc = append(updateDoc, bson.E{Key: "updated_at", Value: time.Now()})

	if len(updateDoc) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "Nenhum campo para atualizar foi fornecido", nil, 0)
		return
	}

	update := bson.D{{Key: "$set", Value: updateDoc}}

	result, err := collection.UpdateOne(ctx, bson.D{{Key: "_id", Value: id}}, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_UPDATE_LEAD_IN_MONGODB)
		return
	}

	if result.MatchedCount == 0 {
		utils.SendResponse(w, http.StatusNotFound, "Tier não encontrado", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", nil, 0)
}
