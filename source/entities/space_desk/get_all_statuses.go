package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetAllStatuses(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	untilStr := r.URL.Query().Get("until")
	minTimestamp := int64(0)
	if untilStr != "" {
		if untilDays, err := strconv.Atoi(untilStr); err == nil && untilDays > 0 {
			minTimestamp = time.Now().Add(-time.Duration(untilDays) * 24 * time.Hour).Unix()
		}
	}

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_EVENTS_WHATSAPP)

	findOptions := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}})
	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEADS_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	allStatuses := []bson.M{}
	for cursor.Next(ctx) {
		doc := bson.M{}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		entries, ok := doc["entry"].([]any)
		if !ok {
			continue
		}
		for _, entryRaw := range entries {
			entry, ok := entryRaw.(map[string]any)
			if !ok {
				continue
			}
			changes, ok := entry["changes"].([]any)
			if !ok {
				continue
			}
			for _, changeRaw := range changes {
				change, ok := changeRaw.(map[string]any)
				if !ok {
					continue
				}
				value, ok := change["value"].(map[string]any)
				if !ok {
					continue
				}
				statuses, ok := value["statuses"].([]any)
				if !ok {
					continue
				}
				for _, statusRaw := range statuses {
					status, ok := statusRaw.(map[string]any)
					if !ok {
						continue
					}
					tsStr, ok := status["timestamp"].(string)
					if !ok {
						continue
					}
					tsInt, err := strconv.ParseInt(tsStr, 10, 64)
					if err != nil {
						continue
					}
					if minTimestamp == 0 || tsInt >= minTimestamp {
						allStatuses = append(allStatuses, status)
					}
				}
			}
		}
	}

	utils.SendResponse(w, http.StatusOK, "", allStatuses, 0)
}
