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

func UpdateOneGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	idStr := r.URL.Query().Get("id")
	field := r.URL.Query().Get("field")
	value := r.URL.Query().Get("value")

	if idStr == "" || field == "" || value == "" {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	objID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.FUNNELS_INVALID_REQUEST_DATA)
		return
	}

	allowed := map[string]bool{"name": true, "status": true, "type": true}
	if !allowed[field] {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	update := bson.M{"$set": bson.M{field: value}}

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	collection := client.Database(database.GetDB()).Collection("groups")

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_UPDATE_SPACE_DESK_GROUP_TO_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Group updated", nil, 0)
}
