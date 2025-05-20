package leads

import (
	"api/source/database"
	"api/source/entities/budgets"
	"api/source/entities/orders"
	"api/source/utils"
	"context"
	"math"
	"net/http"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetAll(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := int64(1)
	pageSize := int64(25)

	if pageStr != "" {
		if parsedPage, err := strconv.ParseInt(pageStr, 10, 64); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if pageSizeStr != "" {
		if parsedPageSize, err := strconv.ParseInt(pageSizeStr, 10, 64); err == nil && parsedPageSize > 0 {
			pageSize = parsedPageSize
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}

	skip := (page - 1) * pageSize

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_LEADS)

	filter := bson.D{}

	totalItems, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEADS_IN_MONGODB)
		return
	}

	totalPages := int64(math.Ceil(float64(totalItems) / float64(pageSize)))

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: pageSize}},
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
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEADS_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	var leads []bson.M
	if err := cursor.All(ctx, &leads); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEADS_IN_MONGODB)
		return
	}

	for i, lead := range leads {
		if relatedOrders, ok := lead["related_orders"].(primitive.A); ok && len(relatedOrders) > 0 {
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
					leads[i]["related_orders_old_data"] = oldOrders
				}
			}
		}

		if relatedBudgets, ok := lead["related_budgets"].(primitive.A); ok && len(relatedBudgets) > 0 {
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
					leads[i]["related_budgets_old_data"] = oldBudgets
				}
			}
		}
	}

	response := map[string]any{
		"items": leads,
		"pagination": map[string]any{
			"page":        page,
			"page_size":   pageSize,
			"total_items": totalItems,
			"total_pages": totalPages,
		},
	}

	utils.SendResponse(w, http.StatusOK, "", response, 0)
}
