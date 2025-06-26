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

func CreateOneTier(w http.ResponseWriter, r *http.Request) {
	tier := &schemas.LeadTier{}
	if err := json.NewDecoder(r.Body).Decode(&tier); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.LEADS_INVALID_REQUEST_DATA)
		return
	}

	tier.Label = strings.TrimSpace(tier.Label)

	if tier.Label == "" {
		utils.SendResponse(w, http.StatusBadRequest, "Label não pode ser vazio", nil, 0)
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

	collection := mongoClient.Database(database.GetDB()).Collection("lead_tiers")

	labelFilter := bson.D{{Key: "label", Value: bson.D{{Key: "$regex", Value: "^" + strings.ReplaceAll(tier.Label, " ", "") + "$"}, {Key: "$options", Value: "i"}}}}
	cursor, err := collection.Find(ctx, labelFilter)
	if err == nil {
		var existing []schemas.LeadTier
		_ = cursor.All(ctx, &existing)
		if len(existing) > 0 {
			utils.SendResponse(w, http.StatusBadRequest, "Já existe um tier com esse label", nil, 0)
			return
		}
	}

	valueFilter := bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "min_value", Value: tier.MinValue}},
		bson.D{{Key: "max_value", Value: tier.MaxValue}},
	}}}
	cursor, err = collection.Find(ctx, valueFilter)
	if err == nil {
		var existing []schemas.LeadTier
		_ = cursor.All(ctx, &existing)
		if len(existing) > 0 {
			utils.SendResponse(w, http.StatusBadRequest, "Já existe um tier com esse min_value ou max_value", nil, 0)
			return
		}
	}

	tier.CreatedAt = time.Now()
	tier.UpdatedAt = time.Now()

	_, err = collection.InsertOne(ctx, tier)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_LEAD_TO_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusCreated, "", nil, 0)
}
