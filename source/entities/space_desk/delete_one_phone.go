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

func DeletePhoneConfig(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload struct {
		Numero string `json:"numero"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	numero := onlyDigits(payload.Numero)
	if numero == "" {
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

	collection := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CONFIG)

	// Remove o telefone do array phones pelo número
	filter := bson.M{"type": "global"}
	update := bson.M{
		"$pull": bson.M{"phones": bson.M{"numero": numero}},
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

	utils.SendResponse(w, http.StatusOK, "", numero, 0)
}
