package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

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

	// Paginação e filtros
	query := r.URL.Query()
	page := 1
	limit := 80
	if p := query.Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	skip := (page - 1) * limit

	filter := bson.M{}
	if title := query.Get("titulo"); title != "" {
		filter["titulo"] = bson.M{"$regex": title, "$options": "i"}
	}

	col := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_READY_MESSAGE)
	findOpts := options.Find().SetLimit(int64(limit)).SetSkip(int64(skip))
	cursor, err := col.Find(ctx, filter, findOpts)
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

	// Retornar também o total de documentos para paginação
	total, _ := col.CountDocuments(ctx, filter)

	resp := bson.M{
		"data":  readyMsgs,
		"page":  page,
		"limit": limit,
		"total": total,
	}

	utils.SendResponse(w, http.StatusOK, "Mensagens prontas recuperadas com sucesso", resp, 0)
}
