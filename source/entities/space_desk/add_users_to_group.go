package spacedesk

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"api/database"
	"api/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func AddUsersToGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload struct {
		GroupIDs []string `json:"group_ids"`
		UserIDs  []string `json:"user_ids"` // <- NOVO
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_SPACE_DESK_GROUP_REQUEST_DATA)
		return
	}

	// Converte group_ids para ObjectID
	var groupObjIDs []bson.ObjectID
	for _, id := range payload.GroupIDs {
		objID, err := bson.ObjectIDFromHex(id)
		if err == nil {
			groupObjIDs = append(groupObjIDs, objID)
		}
	}
	if len(groupObjIDs) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		return
	}

	// Conecta ao MongoDB
	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	groupCol := client.Database(database.GetDB()).Collection("groups")

	// Atualiza todos os grupos: adiciona usuários ao array `user_ids`
	for _, groupID := range groupObjIDs {
		_, err := groupCol.UpdateOne(
			ctx,
			bson.M{"_id": groupID},
			bson.M{"$addToSet": bson.M{"user_ids": bson.M{"$each": payload.UserIDs}}},
		)
		if err != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_SPACE_DESK_GROUP_TO_MONGODB)
			return
		}
	}

	utils.SendResponse(w, http.StatusOK, "Usuários adicionados aos grupos", nil, 0)
}
