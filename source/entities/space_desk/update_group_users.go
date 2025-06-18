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

func UpdateGroupUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload struct {
		ID      string   `json:"id"`
		UserIDs []string `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	groupObjID, err := bson.ObjectIDFromHex(payload.ID)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.FUNNELS_INVALID_REQUEST_DATA)
		return
	}

	update := bson.M{"$set": bson.M{"user_ids": payload.UserIDs}}

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	collection := client.Database(database.GetDB()).Collection("groups")

	_, err = collection.UpdateOne(ctx, bson.M{"_id": groupObjID}, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_UPDATE_SPACE_DESK_GROUP_TO_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Grupo atualizado com usuários", nil, 0)
}
