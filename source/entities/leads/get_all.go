package leads

import (
	"api/database"
	"api/utils"
	"context"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

	filter := buildFilterFromQueryParams(r)

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
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "lead_tiers"},
			{Key: "localField", Value: "related_tier"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "related_tier_data"},
		}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$responsible_data"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$related_client_data"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$related_tier_data"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "responsible", Value: "$responsible_data"},
			{Key: "related_client", Value: "$related_client_data"},
			{Key: "related_tier", Value: "$related_tier_data"},
		}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "tier", Value: bson.D{
				{Key: "$ifNull", Value: bson.A{"$tier", ""}},
			}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "responsible_data", Value: 0},
			{Key: "related_client_data", Value: 0},
			{Key: "related_tier_data", Value: 0},
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

func buildFilterFromQueryParams(r *http.Request) bson.D {
	filter := bson.D{}

	queryParams := r.URL.Query()

	if id := queryParams.Get("id"); id != "" {
		if objectID, err := bson.ObjectIDFromHex(id); err == nil {
			filter = append(filter, bson.E{Key: "_id", Value: objectID})
		}
	}

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
