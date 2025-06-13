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

func onlyDigits(s string) string {
	out := ""
	for _, c := range s {
		if c >= '0' && c <= '9' {
			out += string(c)
		}
	}
	return out
}

func CreatePhoneConfig(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload PhoneConfig
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	payload.Numero = onlyDigits(payload.Numero)
	if payload.Numero == "" {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	if payload.Status == "" {
		payload.Status = "Ativo"
	}
	if payload.Nome == "" {
		payload.Nome = "Telefone"
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

	filterExists := bson.M{
		"type":          "global",
		"phones.numero": payload.Numero,
	}
	count, err := collection.CountDocuments(ctx, filterExists)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}
	if count > 0 {
		utils.SendResponse(w, http.StatusConflict, "", nil, utils.ALREADY_EXISTS)
		return
	}

	doc := collection.FindOne(ctx, bson.M{"type": "global"})
	var settings PhoneSettings
	if err := doc.Decode(&settings); err == nil {
		if len(settings.Phones) >= 3 {
			utils.SendResponse(w, http.StatusConflict, "", nil, utils.LIMIT_REACHED)
			return
		}
	}

	filter := bson.M{"type": "global"}
	update := bson.M{
		"$push":        bson.M{"phones": payload},
		"$setOnInsert": bson.M{"type": "global"},
		"$currentDate": bson.M{"updatedAt": true, "createdAt": true},
	}
	_, err = collection.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}
	utils.SendResponse(w, http.StatusCreated, "", payload, 0)
}
