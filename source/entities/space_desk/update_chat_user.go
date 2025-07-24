package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func UpdateChatUser(w http.ResponseWriter, r *http.Request) {

	type payload struct {
		ChatID  string `json:"chat_id" bson:"chat_id"`
		UserID  string `json:"user_id" bson:"user_id"`
		GroupID string `json:"group_id" bson:"group_id"`
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "Erro ao ler o body", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	var body payload
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "JSON inválido", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	if body.ChatID == "" {
		utils.SendResponse(w, http.StatusBadRequest, "O campo 'chat_id' é obrigatório.", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	chatObjectID, err := bson.ObjectIDFromHex(body.ChatID)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "O 'chat_id' fornecido é inválido.", nil, utils.INVALID_CHAT_ID_FORMAT)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()
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

	updateFields := bson.M{
		"updated_at": time.Now(),
	}

	if body.UserID != "" {
		updateFields["user_id"] = body.UserID
	}

	if body.GroupID != "" {
		updateFields["group_id"] = body.GroupID
	}

	if len(updateFields) == 1 {
		utils.SendResponse(w, http.StatusBadRequest, "Nenhum campo para atualizar foi fornecido", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	update := bson.M{"$set": updateFields}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao atualizar o chat no banco de dados.", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}

	if result.MatchedCount == 0 {
		utils.SendResponse(w, http.StatusNotFound, "Nenhum chat encontrado com o 'chat_id' fornecido.", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Usuário atribuído ao chat com sucesso.", nil, 0)
}
