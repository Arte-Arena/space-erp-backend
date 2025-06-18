package users

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
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

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao buscar usuários.", nil, utils.NOT_FOUND)
		return
	}
	defer cursor.Close(ctx)

	var users []schemas.User
	for cursor.Next(ctx) {
		var user schemas.User
		if err := cursor.Decode(&user); err != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "Erro ao decodificar usuário.", nil, utils.NOT_FOUND)
			return
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro no cursor.", nil, utils.NOT_FOUND)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", users, 0)
}
