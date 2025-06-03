package spacedesk

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"api/database"
	"api/schemas"
	"api/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateOneGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload struct {
		Name   string   `json:"name"`
		Status string   `json:"status"`
		Type   string   `json:"type"`
		Chats  []string `json:"chats"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	group := schemas.Group{
		ID:     bson.NewObjectID(),
		Name:   payload.Name,
		Status: payload.Status,
		Type:   payload.Type,
		Chats:  payload.Chats,
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

	_, err = collection.InsertOne(ctx, group)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_SPACE_DESK_GROUP_TO_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusCreated, "", group, 0)
}
