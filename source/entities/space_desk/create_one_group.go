package spacedesk

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"api/database"
	"api/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type groupPayload struct {
	Name    string          `json:"name"`
	UserIDs []bson.ObjectID `json:"user_ids"`
	Status  string          `json:"status"`
	Type    string          `json:"type"`
}

func CreateOneGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload groupPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	if payload.Name == "" || len(payload.UserIDs) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	mongoURI := os.Getenv(utils.MONGODB_URI)
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	filterIDs := make([]any, len(payload.UserIDs))
	for i, id := range payload.UserIDs {
		filterIDs[i] = id
	}
	filter := bson.M{"user_id": bson.M{"$in": filterIDs}}
	log.Printf("🔍 filter: %+v", filter)

	chatCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)

	cursor, err := chatCol.Find(ctx, filter)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
		return
	}
	defer cursor.Close(ctx)

	var rawDocs []bson.M
	if err := cursor.All(ctx, &rawDocs); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
		return
	}
	var chatIDs []string
	for _, doc := range rawDocs {
		if oid, ok := doc["_id"].(bson.ObjectID); ok {
			chatIDs = append(chatIDs, oid.Hex())
		}
	}

	groupDoc := bson.M{
		"_id":        bson.NewObjectID(),
		"name":       payload.Name,
		"user_ids":   payload.UserIDs,
		"status":     payload.Status,
		"type":       payload.Type,
		"chats":      chatIDs,
		"created_at": time.Now(),
	}
	groupCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_GROUPS)
	if _, err := groupCol.InsertOne(ctx, groupDoc); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_SPACE_DESK_GROUP_TO_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusCreated, "", groupDoc, 0)
}
