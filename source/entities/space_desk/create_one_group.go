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

type Group struct {
	ID      bson.ObjectID           `bson:"_id,omitempty" json:"id"`
	Name    string                  `bson:"name" json:"name"`
	UserIDs []string                `bson:"user_ids" json:"user_ids"`
	Status  string                  `bson:"status" json:"status"`
	Type    string                  `bson:"type" json:"type"`
	Chats   []schemas.SpaceDeskChat `bson:"chats" json:"chats"`
}

func CreateOneGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload struct {
		Name    string   `json:"name"`
		UserIDs []string `json:"user_ids"`
		Status  string   `json:"status"`
		Type    string   `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
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

	chatCol := client.Database(database.GetDB()).Collection("chats")
	filter := bson.M{"user_id": bson.M{"$in": payload.UserIDs}}
	cursor, err := chatCol.Find(ctx, filter)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
		return
	}
	defer cursor.Close(ctx)

	group := schemas.Group{
		ID:      bson.NewObjectID(),
		Name:    payload.Name,
		UserIDs: payload.UserIDs,
		Status:  payload.Status,
		Type:    payload.Type,
		Chats:   []string{},
	}

	collection := client.Database(database.GetDB()).Collection("groups")
	_, err = collection.InsertOne(ctx, group)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_SPACE_DESK_GROUP_TO_MONGODB)
		return
	}

	var chats []string
	if err = cursor.All(ctx, &chats); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
		return
	}

	group.Chats = chats

	utils.SendResponse(w, http.StatusCreated, "", group, 0)
}
