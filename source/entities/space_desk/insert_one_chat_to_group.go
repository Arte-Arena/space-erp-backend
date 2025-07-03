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

func AddChatToGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload struct {
		GroupIDs []string `json:"group_ids"` // <- array agora
		ChatID   string   `json:"chat_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_SPACE_DESK_GROUP_REQUEST_DATA)
		return
	}

	// Converte os IDs para ObjectID usando bson.ObjectID
	var groupObjIDs []bson.ObjectID // Usando bson.ObjectID
	for _, id := range payload.GroupIDs {
		objID, err := bson.ObjectIDFromHex(id) // Usando bson.ObjectIDFromHex
		if err == nil {
			groupObjIDs = append(groupObjIDs, objID)
		}
	}
	if len(groupObjIDs) == 0 {
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

	db := client.Database(database.GetDB())
	groupCol := db.Collection("groups")
	chatCol := db.Collection(database.COLLECTION_SPACE_DESK_CHAT)

	// Atualiza todos os grupos
	for _, groupID := range groupObjIDs {
		_, err := groupCol.UpdateOne(
			ctx,
			bson.M{"_id": groupID},
			bson.M{"$addToSet": bson.M{"chats": payload.ChatID}},
		)
		if err != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_SPACE_DESK_GROUP_TO_MONGODB)
			return
		}
	}

	// Atualiza o chat com todos os group_ids
	_, err = chatCol.UpdateOne(
		ctx,
		bson.M{"cliente_phone_number": payload.ChatID},
		bson.M{"$addToSet": bson.M{"group_ids": bson.M{"$each": groupObjIDs}}},
	)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_SPACE_DESK_GROUP_TO_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Chat added to groups", nil, 0)
}
