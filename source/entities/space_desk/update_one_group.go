package spacedesk

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"api/database"
	"api/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func UpdateOneGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload struct {
		ID      string   `json:"id"`
		Name    string   `json:"name"`
		UserIds []string `json:"user_ids"`
		Status  string   `json:"status"`
		Type    string   `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	objID, err := bson.ObjectIDFromHex(payload.ID)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.FUNNELS_INVALID_REQUEST_DATA)
		return
	}

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	// Buscar os chats com base nos user_ids
	chatCol := client.Database(database.GetDB()).Collection("chats")
	filter := bson.M{"user_id": bson.M{"$in": payload.UserIds}}
	cursor, err := chatCol.Find(ctx, filter)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
		return
	}
	defer cursor.Close(ctx)

	var chats []string
	if err = cursor.All(ctx, &chats); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
		return
	}

	update := bson.M{"$set": bson.M{
		"name":     payload.Name,
		"user_ids": payload.UserIds,
		"status":   payload.Status,
		"type":     payload.Type,
		"chats":    chats,
	}}

	collection := client.Database(database.GetDB()).Collection("groups")

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_UPDATE_SPACE_DESK_GROUP_TO_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Group updated", nil, 0)
}
