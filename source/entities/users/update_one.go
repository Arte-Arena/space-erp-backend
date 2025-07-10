package users

import (
	"api/database"
	"api/middlewares"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func UpdateOne(w http.ResponseWriter, r *http.Request) {
	ctxUserRaw := r.Context().Value(middlewares.UserContextKey)
	if ctxUserRaw == nil {
		utils.SendResponse(w, http.StatusUnauthorized, "Usuário não autenticado", nil, 0)
		return
	}
	laravelUser, ok := ctxUserRaw.(middlewares.LaravelUser)
	if !ok {
		utils.SendResponse(w, http.StatusUnauthorized, "Usuário inválido", nil, 0)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	var authenticatedUserDoc schemas.User
	err = client.Database(database.GetDB()).Collection(database.COLLECTION_USERS).
		FindOne(ctx, bson.M{"old_id": laravelUser.ID}).Decode(&authenticatedUserDoc)
	if err != nil {
		utils.SendResponse(w, http.StatusUnauthorized, "Usuário não encontrado", nil, utils.NOT_FOUND)
		return
	}

	isSuperAdmin := slices.Contains(authenticatedUserDoc.Role, schemas.USERS_ROLE_SUPER_ADMIN)
	if !isSuperAdmin {
		utils.SendResponse(w, http.StatusForbidden, "Usuário não possui permissão super_admin", nil, 0)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	oldID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_USER_ID_FORMAT)
		return
	}

	user := &schemas.User{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	if user.Commission > 100 {
		utils.SendResponse(w, http.StatusBadRequest, "Comissão deve ser entre 0 e 100", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	collection := client.Database(database.GetDB()).Collection(database.COLLECTION_USERS)

	filter := bson.D{{Key: "old_id", Value: oldID}}

	updateDoc := bson.D{}

	if user.Name != "" {
		updateDoc = append(updateDoc, bson.E{Key: "name", Value: user.Name})
	}

	if user.Email != "" {
		updateDoc = append(updateDoc, bson.E{Key: "email", Value: user.Email})
	}

	if len(user.Role) > 0 {
		updateDoc = append(updateDoc, bson.E{Key: "role", Value: user.Role})
	}

	updateDoc = append(updateDoc, bson.E{Key: "commission", Value: user.Commission})

	updateDoc = append(updateDoc, bson.E{Key: "updated_at", Value: time.Now()})

	if len(updateDoc) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "Nenhum campo para atualizar foi fornecido", nil, 0)
		return
	}

	update := bson.D{{Key: "$set", Value: updateDoc}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}

	if result.MatchedCount == 0 {
		utils.SendResponse(w, http.StatusNotFound, "Usuário não encontrado", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", nil, 0)
}
