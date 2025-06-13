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

type PhoneConfig struct {
	Nome   string `bson:"nome" json:"nome"`
	Numero string `bson:"numero" json:"numero"`
	Status string `bson:"status" json:"status"`
	Label  string `bson:"label" json:"label"`
}

func UpdatePhoneConfig(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload map[string]string // Aceita só strings, mais flexível para PATCH
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	numeroAntigo, ok := payload["numero"]
	if !ok || numeroAntigo == "" {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	// Montar os updates só para os campos enviados, exceto 'numero' identificador
	updateFields := bson.M{}
	if nome, ok := payload["nome"]; ok {
		updateFields["phones.$.nome"] = nome
	}
	if numero, ok := payload["novoNumero"]; ok {
		updateFields["phones.$.numero"] = numero
	}
	if status, ok := payload["status"]; ok {
		updateFields["phones.$.status"] = status
	}
	if label, ok := payload["label"]; ok {
		updateFields["phones.$.label"] = label
	}

	if len(updateFields) == 0 {
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

	filter := bson.M{
		"type":          "global", // personalize para multi-tenant se quiser
		"phones.numero": numeroAntigo,
	}

	update := bson.M{
		"$set":         updateFields,
		"$currentDate": bson.M{"updatedAt": true},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}

	if result.ModifiedCount == 0 {
		utils.SendResponse(w, http.StatusNotFound, "", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", updateFields, 0)
}
