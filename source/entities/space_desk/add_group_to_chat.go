package spacedesk

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"api/database"
	"api/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func AddGroupToChat(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	// 1) Decodifica payload
	var payload struct {
		ChatID  string `json:"chat_id"`
		GroupID string `json:"group_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_SPACE_DESK_GROUP_REQUEST_DATA)
		return
	}

	// 2) Valida se os campos obrigatórios estão presentes
	if payload.ChatID == "" {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_SPACE_DESK_GROUP_REQUEST_DATA)
		return
	}
	if payload.GroupID == "" {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_SPACE_DESK_GROUP_REQUEST_DATA)
		return
	}

	// 3) Converte ChatID para ObjectID
	chatObjID, err := bson.ObjectIDFromHex(payload.ChatID)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_GET_SPACE_DESK_CHAT_ID)
		return
	}

	// 4) Conecta no Mongo
	client, err := mongo.Connect(options.Client().ApplyURI(os.Getenv(utils.MONGODB_URI)))
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	chatCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)

	// 5) Verifica se o chat existe
	var chatDoc struct {
		ID string `bson:"_id"`
	}
	if err := chatCol.FindOne(ctx, bson.M{"_id": chatObjID}).Decode(&chatDoc); err != nil {
		utils.SendResponse(w, http.StatusNotFound, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
		return
	}

	// 6) Atualiza o chat: adiciona group_id
	update := bson.M{"$set": bson.M{
		"group_id":   payload.GroupID,
		"updated_at": time.Now(),
	}}
	if _, err := chatCol.UpdateOne(ctx, bson.M{"_id": chatObjID}, update); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_SPACE_DESK_CHAT_TO_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Grupo adicionado ao chat com sucesso", nil, 0)
}
