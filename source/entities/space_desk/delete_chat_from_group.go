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

func DeleteChatFromGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload struct {
		GroupID string `json:"group_id"`
		ChatID  string `json:"chat_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.LEADS_INVALID_REQUEST_DATA)
		return
	}
	objID, err := bson.ObjectIDFromHex(payload.GroupID)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
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

	collection := client.Database(database.GetDB()).Collection("groups")

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$pull": bson.M{"chats": payload.ChatID}},
	)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_DELETE_CHAT_FROM_SPACE_DESK_GROUP)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Chat removed from group", nil, 0)
}
