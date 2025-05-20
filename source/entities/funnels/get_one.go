package funnels

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
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_FUNNEL_ID_FORMAT)
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

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: id}}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$stages"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_LEADS},
			{Key: "localField", Value: "stages.related_leads"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "stages.related_leads_data"},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_BUDGETS},
			{Key: "localField", Value: "stages.related_budgets"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "stages.related_budgets_data"},
		}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "stages.related_leads", Value: "$stages.related_leads_data"},
			{Key: "stages.related_budgets", Value: "$stages.related_budgets_data"},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "stages.related_leads_data", Value: 0},
			{Key: "stages.related_budgets_data", Value: 0},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$_id"},
			{Key: "name", Value: bson.D{{Key: "$first", Value: "$name"}}},
			{Key: "type", Value: bson.D{{Key: "$first", Value: "$type"}}},
			{Key: "stages", Value: bson.D{{Key: "$push", Value: "$stages"}}},
			{Key: "created_at", Value: bson.D{{Key: "$first", Value: "$created_at"}}},
			{Key: "updated_at", Value: bson.D{{Key: "$first", Value: "$updated_at"}}},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_FUNNEL_BY_ID_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_FUNNEL_BY_ID_IN_MONGODB)
		return
	}

	if len(results) == 0 {
		utils.SendResponse(w, http.StatusNotFound, "Funil não encontrado", nil, 0)
		return
	}

	result := results[0]
	stages, hasStages := result["stages"].(primitive.A)
	if hasStages {
		for i, stageObj := range stages {
			if stage, isStage := stageObj.(bson.M); isStage {
				if relatedBudgets, ok := stage["related_budgets"].(primitive.A); ok && len(relatedBudgets) > 0 {
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
							stage["related_budgets_old_data"] = oldBudgets
						}
					}
				}

				if relatedOrders, ok := stage["related_orders"].(primitive.A); ok && len(relatedOrders) > 0 {
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
							stage["related_orders_old_data"] = oldOrders
						}
					}
				}

				stages[i] = stage
			}
		}
		result["stages"] = stages
	}

	utils.SendResponse(w, http.StatusOK, "", result, 0)
}
