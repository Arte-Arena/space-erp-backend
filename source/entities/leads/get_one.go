package leads

import (
	"api/source/database"
	"api/source/utils"
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetOne(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := bson.ObjectIDFromHex(idStr)
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

	filter := buildFilterForGetOne(r, id)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
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

	utils.SendResponse(w, http.StatusOK, "", result, 0)
}

func buildFilterForGetOne(r *http.Request, id bson.ObjectID) bson.D {
	filter := bson.D{{Key: "_id", Value: id}}

	queryParams := r.URL.Query()

	textFields := map[string]string{
		"name":        "name",
		"nickname":    "nickname",
		"phone":       "phone",
		"type":        "type",
		"segment":     "segment",
		"status":      "status",
		"source":      "source",
		"platform_id": "platform_id",
		"rating":      "rating",
	}

	for param, field := range textFields {
		if value := queryParams.Get(param); value != "" {
			if queryParams.Get(param+"_exact") == "true" {
				filter = append(filter, bson.E{Key: field, Value: value})
			} else {
				filter = append(filter, bson.E{Key: field, Value: bson.D{{Key: "$regex", Value: value}, {Key: "$options", Value: "i"}}})
			}
		}
	}

	dateFields := []string{"created_at", "updated_at"}
	for _, field := range dateFields {
		if startDate := queryParams.Get(field + "_start"); startDate != "" {
			if parsedDate, err := time.Parse(time.RFC3339, startDate); err == nil {
				filter = append(filter, bson.E{Key: field, Value: bson.D{{Key: "$gte", Value: parsedDate}}})
			}
		}

		if endDate := queryParams.Get(field + "_end"); endDate != "" {
			if parsedDate, err := time.Parse(time.RFC3339, endDate); err == nil {
				filter = append(filter, bson.E{Key: field, Value: bson.D{{Key: "$lte", Value: parsedDate}}})
			}
		}
	}

	objectIDFields := map[string]string{
		"related_client": "related_client",
		"responsible":    "responsible",
	}

	for param, field := range objectIDFields {
		if value := queryParams.Get(param); value != "" {
			if objectID, err := bson.ObjectIDFromHex(value); err == nil {
				filter = append(filter, bson.E{Key: field, Value: objectID})
			}
		}
	}

	multiValueFields := []string{"status", "type", "segment", "source", "rating"}
	for _, field := range multiValueFields {
		if values := queryParams.Get(field + "_in"); values != "" {
			valuesList := strings.Split(values, ",")
			if len(valuesList) > 1 {
				orConditions := bson.A{}
				for _, val := range valuesList {
					orConditions = append(orConditions, bson.D{{Key: field, Value: strings.TrimSpace(val)}})
				}
				filter = append(filter, bson.E{Key: "$or", Value: orConditions})
			}
		}
	}

	return filter
}
