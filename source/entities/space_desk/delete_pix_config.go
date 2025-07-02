package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DeletePixConfig(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	objID, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	collection := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CONFIG)

	filter := bson.M{"type": "global"}
	update := bson.M{
		"$pull":        bson.M{"pix": bson.M{"_id": objID}},
		"$currentDate": bson.M{"updatedAt": true},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}
	if result.ModifiedCount == 0 {
		utils.SendResponse(w, http.StatusNotFound, "", nil, utils.NOT_FOUND)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", map[string]string{"id": idParam}, 0)
}
