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

type UpdateReadyMessage struct {
	ID       string   `json:"_id"`
	Title    string   `json:"title"`
	Messages []string `json:"messages"`
}

func UpdateOneReadyMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var body UpdateReadyMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Println("Erro ao decodificar corpo da requisição:", err)
		utils.SendResponse(w, http.StatusBadRequest, "JSON inválido: "+err.Error(), nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	if body.ID == "" || body.Title == "" || len(body.Messages) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "Campos obrigatórios ausentes", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), database.MONGO_TIMEOUT)
	defer cancel()

	clientOpts := options.Client().ApplyURI(os.Getenv(utils.MONGODB_URI))
	dbClient, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Println("Erro ao conectar ao MongoDB:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao conectar ao MongoDB: "+err.Error(), nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer dbClient.Disconnect(ctx)

	col := dbClient.Database(database.GetDB()).Collection("ReadyChatMessages")

	update := bson.M{
		"$set": bson.M{
			"titulo":     body.Title,
			"menssagens": body.Messages,
			"updatedAt":  time.Now().UTC(),
		},
	}

	filter := bson.M{"_id": body.ID}
	res, err := col.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("Erro ao atualizar mensagem pronta:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao atualizar mensagem pronta", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Mensagem pronta atualizada com sucesso", res.ModifiedCount, 0)
}
