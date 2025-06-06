package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UpdateChatStatusBody struct {
	ID     string `json:"id"`     // pode ser _id do Mongo ou cliente_phone_number
	Status string `json:"status"` // "active" ou "inactive"
}

func UpdateChatStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var body UpdateChatStatusBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "JSON inválido", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	if body.ID == "" || body.Status == "" {
		utils.SendResponse(w, http.StatusBadRequest, "Campos obrigatórios ausentes", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

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

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT_METADATA)

	objectID, _ := bson.ObjectIDFromHex(body.ID)
	filter := bson.M{"_id": objectID}

	update := bson.M{
		"$set": bson.M{
			"status":     body.Status,
			"updated_at": time.Now().UTC(),
		},
	}

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
