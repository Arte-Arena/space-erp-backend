package leads

import (
	"api/source/database"
	"api/source/entities/budgets"
	"api/source/entities/orders"
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
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_LEAD_ID_FORMAT)
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

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: id}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_BUDGETS},
			{Key: "localField", Value: "related_budgets"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "related_budgets"},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_ORDERS},
			{Key: "localField", Value: "related_orders"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "related_orders"},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_USERS},
			{Key: "localField", Value: "responsible"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "responsible_data"},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_CLIENTS},
			{Key: "localField", Value: "related_client"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "related_client_data"},
		}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$responsible_data"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$related_client_data"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "responsible", Value: "$responsible_data"},
			{Key: "related_client", Value: "$related_client_data"},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "responsible_data", Value: 0},
			{Key: "related_client_data", Value: 0},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEAD_BY_ID_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEAD_BY_ID_IN_MONGODB)
		return
	}

	if len(results) == 0 {
		utils.SendResponse(w, http.StatusNotFound, "Lead não encontrado", nil, 0)
		return
	}

	result := results[0]

	if relatedOrders, ok := result["related_orders"].(primitive.A); ok && len(relatedOrders) > 0 {
		orderOldIDs := make([]int, 0)
		for _, orderObj := range relatedOrders {
			if orderMap, isMap := orderObj.(bson.M); isMap {
				if oldID, hasOldID := orderMap["old_id"]; hasOldID {
					if oldIDInt, canConvert := oldID.(int64); canConvert {
						orderOldIDs = append(orderOldIDs, int(oldIDInt))
					}
				}
			}
		}

		if len(orderOldIDs) > 0 {
			oldOrders, err := orders.GetManyOld(orderOldIDs)
			if err == nil && oldOrders != nil {
				result["related_orders_old_data"] = oldOrders
			}
		}
	}

	if relatedBudgets, ok := result["related_budgets"].(primitive.A); ok && len(relatedBudgets) > 0 {
		budgetOldIDs := make([]int, 0)
		for _, budgetObj := range relatedBudgets {
			if budgetMap, isMap := budgetObj.(bson.M); isMap {
				if oldID, hasOldID := budgetMap["old_id"]; hasOldID {
					if oldIDInt, canConvert := oldID.(int64); canConvert {
						budgetOldIDs = append(budgetOldIDs, int(oldIDInt))
					}
				}
			}
		}

		if len(budgetOldIDs) > 0 {
			oldBudgets, err := budgets.GetManyOld(budgetOldIDs)
			if err == nil && oldBudgets != nil {
				result["related_budgets_old_data"] = oldBudgets
			}
		}
	}

	utils.SendResponse(w, http.StatusOK, "", result, 0)
}
