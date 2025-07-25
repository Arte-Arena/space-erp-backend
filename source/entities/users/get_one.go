package users

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"net/http"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetOne(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	oldID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_USER_ID_FORMAT)
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

	filter := bson.D{{Key: "old_id", Value: oldID}}

	user := schemas.User{}
	err = collection.FindOne(ctx, filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		utils.SendResponse(w, http.StatusNotFound, "Usuário não encontrado", nil, 0)
		return
	}
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_USER_BY_ID_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", user, 0)
}
