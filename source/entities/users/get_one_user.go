package users

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"sync"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var userCache sync.Map

func GetOneUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	oldID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "ID inválido, deve ser um número inteiro.", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	if cachedUser, ok := userCache.Load(oldID); ok {
		utils.SendResponse(w, http.StatusOK, "", cachedUser, 0)
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

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_USERS)

	var user schemas.User
	filter := bson.M{"old_id": oldID}

	err = collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.SendResponse(w, http.StatusNotFound, "Utilizador não encontrado.", nil, utils.NOT_FOUND)
		} else {
			utils.SendResponse(w, http.StatusInternalServerError, "Erro ao buscar utilizador.", nil, utils.NOT_FOUND)
		}
		return
	}

	userCache.Store(oldID, user)

	utils.SendResponse(w, http.StatusOK, "", user, 0)
}
