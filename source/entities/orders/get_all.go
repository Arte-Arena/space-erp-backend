package orders

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

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_ORDERS)

	filter := buildFilterFromQueryParams(r)

	totalItems, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_ORDERS_IN_MONGODB)
		return
	}

	totalPages := int64(math.Ceil(float64(totalItems) / float64(pageSize)))

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: pageSize}},

		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_USERS},
			{Key: "localField", Value: "created_by"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "created_by_data"},
		}}},

		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_USERS},
			{Key: "localField", Value: "related_seller"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "related_seller_data"},
		}}},

		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_USERS},
			{Key: "localField", Value: "related_designer"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "related_designer_data"},
		}}},

		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_BUDGETS},
			{Key: "localField", Value: "related_budget"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "related_budget_data"},
		}}},

		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$created_by_data"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$related_seller_data"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$related_designer_data"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$related_budget_data"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "created_by", Value: "$created_by_data"},
			{Key: "related_seller", Value: "$related_seller_data"},
			{Key: "related_designer", Value: "$related_designer_data"},
			{Key: "related_budget", Value: "$related_budget_data"},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "created_by_data", Value: 0},
			{Key: "related_seller_data", Value: 0},
			{Key: "related_designer_data", Value: 0},
			{Key: "related_budget_data", Value: 0},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_ORDERS_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	orders := []bson.M{}
	if err := cursor.All(ctx, &orders); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_ORDERS_IN_MONGODB)
		return
	}

	response := map[string]any{
		"items": orders,
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

	if oldID := queryParams.Get("old_id"); oldID != "" {
		if parsedOldID, err := strconv.ParseUint(oldID, 10, 64); err == nil {
			filter = append(filter, bson.E{Key: "old_id", Value: parsedOldID})
		}
	}

	if tinyID := queryParams.Get("tiny_id"); tinyID != "" {
		filter = append(filter, bson.E{Key: "tiny.id", Value: tinyID})
	}

	if tinyNumber := queryParams.Get("tiny_number"); tinyNumber != "" {
		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "tiny.number", Value: bson.D{{Key: "$regex", Value: tinyNumber}, {Key: "$options", Value: "i"}}}},
			bson.D{{Key: "tiny.numero", Value: bson.D{{Key: "$regex", Value: tinyNumber}, {Key: "$options", Value: "i"}}}},
		}})
	}

	objectIDFields := map[string]string{
		"created_by":       "created_by",
		"related_seller":   "related_seller",
		"related_designer": "related_designer",
		"related_budget":   "related_budget",
	}
	for param, field := range objectIDFields {
		if value := queryParams.Get(param); value != "" {
			if objectID, err := bson.ObjectIDFromHex(value); err == nil {
				filter = append(filter, bson.E{Key: field, Value: objectID})
			}
		}
	}

	stringFields := []string{"tracking_code", "url_trello", "products_list_legacy", "notes"}
	for _, field := range stringFields {
		if value := queryParams.Get(field); value != "" {
			if queryParams.Get(field+"_exact") == "true" {
				filter = append(filter, bson.E{Key: field, Value: value})
			} else {
				filter = append(filter, bson.E{Key: field, Value: bson.D{{Key: "$regex", Value: value}, {Key: "$options", Value: "i"}}})
			}
		}
	}

	enumFields := []string{"status", "stage", "type"}
	for _, field := range enumFields {
		if value := queryParams.Get(field); value != "" {
			filter = append(filter, bson.E{Key: field, Value: value})
		}
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

	dateFields := []string{"created_at", "updated_at", "expected_date", "payment_date"}
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

	return filter
}
