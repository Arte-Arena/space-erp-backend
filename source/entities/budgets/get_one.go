package budgets

import (
	"api/source/database"
	"api/source/utils"
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
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_BUDGET_ID_FORMAT)
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

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_BUDGETS)

	filter := bson.D{{Key: "_id", Value: id}}

	var result bson.M
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		utils.SendResponse(w, http.StatusNotFound, "Orçamento não encontrado", nil, 0)
		return
	}
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_BUDGET_BY_ID_IN_MONGODB)
		return
	}

	if oldID, hasOldID := result["old_id"]; hasOldID {
		if oldIDInt, canConvert := oldID.(int64); canConvert {
			oldBudget, err := GetOneOld(int(oldIDInt))
			if err == nil && oldBudget != nil {
				result["old_data"] = oldBudget
			}
		}
	}

	utils.SendResponse(w, http.StatusOK, "", result, 0)
}
