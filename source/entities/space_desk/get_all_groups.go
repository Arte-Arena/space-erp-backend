package spacedesk

import (
	"context"
	"net/http"
	"os"

	"api/database"
	"api/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetAllGroups(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	collection := client.Database(database.GetDB()).Collection("groups")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_GROUPS)
		return
	}
	defer cursor.Close(ctx)

	var groups []bson.M
	if err := cursor.All(ctx, &groups); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_PARSE_SPACE_DESK_GROUPS)
		return
	}

	chatCol := client.Database(database.GetDB()).Collection("chats")
	for i, group := range groups {
		userIDs, ok := group["user_ids"].([]any)
		if !ok {
			continue
		}
		var ids []string
		for _, id := range userIDs {
			if strID, ok := id.(string); ok {
				ids = append(ids, strID)
			}
		}
		filter := bson.M{"user_id": bson.M{"$in": ids}}
		cursor, err := chatCol.Find(ctx, filter)
		if err != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
			return
		}
		defer cursor.Close(ctx)
		var chats []bson.M
		if err := cursor.All(ctx, &chats); err != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
			return
		}
		group["chats_list"] = chats
		groups[i] = group
	}

	utils.SendResponse(w, http.StatusOK, "Groups retrieved successfully", groups, 0)
}
