package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func UpdateChatUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload struct {
		ChatID string `json:"chat_id"`
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "Corpo da requisição JSON inválido.", nil, 1)
		return
	}

	if payload.ChatID == "" {
		utils.SendResponse(w, http.StatusBadRequest, "O campo 'chat_id' é obrigatório.", nil, 1)
		return
	}
	if payload.UserID == "" {
		utils.SendResponse(w, http.StatusBadRequest, "O campo 'user_id' é obrigatório.", nil, 1)
		return
	}

	chatObjectID, err := bson.ObjectIDFromHex(payload.ChatID)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "O 'chat_id' fornecido é inválido.", nil, 1)
		return
	}

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)

	filter := bson.M{"_id": chatObjectID}
	update := bson.M{
		"$set": bson.M{
			"user_id":    payload.UserID,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao atualizar o chat no banco de dados.", nil, 1)
		return
	}

	if result.MatchedCount == 0 {
		utils.SendResponse(w, http.StatusNotFound, "Nenhum chat encontrado com o 'chat_id' fornecido.", nil, 1)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Usuário atribuído ao chat com sucesso.", nil, 0)
}
