package funnelsplacements

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"math"
	"net/http"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetAll(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	pageSize := r.URL.Query().Get("page_size")

	pageNumber := 1
	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageNumber = p
		}
	}

	pageSizeNumber := 20
	if pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			pageSizeNumber = ps
		}
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

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_FUNNELS_PLACEMENTS)

	filter := bson.D{}

	funnelID := r.URL.Query().Get("funnel_id")
	if funnelID != "" {
		if objID, err := bson.ObjectIDFromHex(funnelID); err == nil {
			filter = append(filter, bson.E{Key: "funnel_id", Value: objID})
		}
	}

	relatedLead := r.URL.Query().Get("related_lead")
	if relatedLead != "" {
		if objID, err := bson.ObjectIDFromHex(relatedLead); err == nil {
			filter = append(filter, bson.E{Key: "related_lead", Value: objID})
		}
	}

	stageName := r.URL.Query().Get("stage_name")
	if stageName != "" {
		filter = append(filter, bson.E{Key: "stage_name", Value: stageName})
	}

	totalItems, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, CANNOT_FIND_FUNNEL_PLACEMENTS_IN_MONGODB)
		return
	}

	totalPages := int64(math.Ceil(float64(totalItems) / float64(pageSizeNumber)))

	findOptions := options.Find().
		SetSkip(int64((pageNumber - 1) * pageSizeNumber)).
		SetLimit(int64(pageSizeNumber)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, CANNOT_FIND_FUNNEL_PLACEMENTS_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	var funnelPlacements []schemas.FunnelPlacement
	if err = cursor.All(ctx, &funnelPlacements); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, CANNOT_FIND_FUNNEL_PLACEMENTS_IN_MONGODB)
		return
	}

	response := map[string]any{
		"data":         funnelPlacements,
		"current_page": pageNumber,
		"page_size":    pageSizeNumber,
		"total_items":  totalItems,
		"total_pages":  totalPages,
	}

	utils.SendResponse(w, http.StatusOK, "", response, 0)
}
