package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UpdateChatStatusBody struct {
	ID           string `json:"id"` 
	Closed       bool   `json:"closed" bson:"closed"`
	NeedTemplate bool   `json:"need_template" bson:"need_template"`
	Blocked      *bool  `json:"blocked,omitempty"`
}

func UpdateChatStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "Erro ao ler o body", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	var body UpdateChatStatusBody
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "JSON inválido", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	if body.ID == "" {
		utils.SendResponse(w, http.StatusBadRequest, "Campo 'id' obrigatório ausente", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	var raw map[string]any
	_ = json.Unmarshal(bodyBytes, &raw)

	ctx, cancel := context.WithTimeout(r.Context(), database.MONGO_TIMEOUT)
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

	objectID, _ := bson.ObjectIDFromHex(body.ID)
	filter := bson.M{"_id": objectID}

	updateFields := bson.M{
		"updated_at": time.Now().UTC(),
	}
	if body.Blocked != nil {
		updateFields["blocked"] = *body.Blocked
	}
	if _, ok := raw["closed"]; ok {
		updateFields["closed"] = body.Closed
	}
	if _, ok := raw["need_template"]; ok {
		updateFields["need_template"] = body.NeedTemplate
	}
	if len(updateFields) == 1 {
		utils.SendResponse(w, http.StatusBadRequest, "Nenhum campo para atualizar foi fornecido", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	update := bson.M{"$set": updateFields}

	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("Erro ao atualizar status do chat:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao atualizar status do chat", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}

	if res.MatchedCount == 0 {
		log.Printf("Chat não encontrado para filtro: %+v\n", filter)
		utils.SendResponse(w, http.StatusNotFound, "Chat não encontrado", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Status do chat atualizado com sucesso", nil, 0)
}
