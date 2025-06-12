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
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$unwind", Value: "$entry"}},
		bson.D{{Key: "$unwind", Value: "$entry.changes"}},
		bson.D{{Key: "$unwind", Value: "$entry.changes.value.statuses"}},
		bson.D{
			{Key: "$replaceRoot", Value: bson.D{
				{Key: "newRoot", Value: "$entry.changes.value.statuses"},
			}},
		},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: -1}}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEADS_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	allStatuses := []bson.M{}
	uniqueStatuses := map[string]bson.M{}

	for cursor.Next(ctx) {
		doc := bson.M{}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		// Filtro pelo timestamp, se necessário
		tsStr, ok := doc["timestamp"].(string)
		if !ok {
			continue
		}
		tsInt, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			continue
		}
		if minTimestamp != 0 && tsInt < minTimestamp {
			continue
		}

		// Deduplicação pelo id
		idStr, ok := doc["id"].(string)
		if !ok {
			continue
		}
		// Só adiciona se ainda não tem no mapa
		if _, found := uniqueStatuses[idStr]; !found {
			uniqueStatuses[idStr] = doc
		}
		// Se quiser sempre o último, faça: uniqueStatuses[idStr] = doc
	}

	for _, status := range uniqueStatuses {
		allStatuses = append(allStatuses, status)
	}

	utils.SendResponse(w, http.StatusOK, "", allStatuses, 0)
}
