package spacedesk

import (
	"context"
	"net/http"
	"os"

	"api/database"
	"api/schemas"
	"api/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetChatsByGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	groupID := r.URL.Query().Get("group_id")
	if groupID == "" {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	groupObjID, err := bson.ObjectIDFromHex(groupID)
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

	chatCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)

	filter := bson.M{"group_ids": groupObjID}
	cursor, err := chatCol.Find(ctx, filter)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHATS_BY_GROUP)
		return
	}
	defer cursor.Close(ctx)

	var chats []schemas.SpaceDeskChatMetadata
	for cursor.Next(ctx) {
		var chat schemas.SpaceDeskChatMetadata
		if err := cursor.Decode(&chat); err != nil {
			continue
		}
		chats = append(chats, chat)
	}

	utils.SendResponse(w, http.StatusOK, "", chats, 0)
}
