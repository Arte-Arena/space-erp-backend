package spacedesk

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"api/database"
	"api/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func AddUsersToGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	// 1) Decodifica payload
	var payload struct {
		GroupID string          `json:"group_id"`
		UserIDs []bson.ObjectID `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_SPACE_DESK_GROUP_REQUEST_DATA)
		return
	}

	// 2) Converte GroupID para ObjectID
	groupObjID, err := bson.ObjectIDFromHex(payload.GroupID)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		return
	}

	// 3) Conecta no Mongo
	client, err := mongo.Connect(options.Client().ApplyURI(os.Getenv(utils.MONGODB_URI)))
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	groupsCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_GROUPS)

	// 4) Busca o grupo atual
	var grp struct {
		UserIDs []string `bson:"user_ids"`
	}
	if err := groupsCol.FindOne(ctx, bson.M{"_id": groupObjID}).Decode(&grp); err != nil {
		utils.SendResponse(w, http.StatusNotFound, "", nil, utils.CANNOT_FIND_SPACE_DESK_GROUPS)
		return
	}

	// 5) Merge de user_ids (old_ids)
	idSet := make(map[string]struct{}, len(grp.UserIDs)+len(payload.UserIDs))
	for _, old := range grp.UserIDs {
		idSet[old] = struct{}{}
	}
	for _, old := range payload.UserIDs {
		idSet[old.Hex()] = struct{}{}
	}
	mergedIDs := make([]string, 0, len(idSet))
	for old := range idSet {
		mergedIDs = append(mergedIDs, old)
	}

	// 6) Atualiza o grupo: novos user_ids
	update := bson.M{"$set": bson.M{
		"user_ids":   mergedIDs,
		"updated_at": time.Now(),
	}}
	if _, err := groupsCol.UpdateOne(ctx, bson.M{"_id": groupObjID}, update); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_SPACE_DESK_GROUP_TO_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Usuários adicionados ao grupo", nil, 0)
}
