package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DeleteUserFromGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload struct {
		GroupID string `json:"group_id"`
		UserID  string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.LEADS_INVALID_REQUEST_DATA)
		return
	}

	groupObjID, err := bson.ObjectIDFromHex(payload.GroupID)
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

	groupCol := client.Database(database.GetDB()).Collection("groups")

	// Remove o usuário do array user_ids
	_, err = groupCol.UpdateOne(
		ctx,
		bson.M{"_id": groupObjID},
		bson.M{"$pull": bson.M{"user_ids": payload.UserID}},
	)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_DELETE_CHAT_FROM_SPACE_DESK_GROUP)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Usuário removido do grupo", nil, 0)
}
