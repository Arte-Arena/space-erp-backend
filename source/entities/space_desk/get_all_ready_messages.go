package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"log"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetAllReadyMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
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

	col := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_READY_MESSAGE)
	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		log.Println("Erro ao buscar mensagens prontas:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao buscar mensagens prontas", nil, utils.ERROR_TO_FIND_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	var readyMsgs []bson.M
	if err := cursor.All(ctx, &readyMsgs); err != nil {
		log.Println("Erro ao decodificar mensagens prontas:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao decodificar mensagens prontas", nil, utils.ERROR_TO_FIND_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Mensagens prontas recuperadas com sucesso", readyMsgs, 0)
}
