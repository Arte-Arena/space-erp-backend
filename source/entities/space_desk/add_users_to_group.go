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
		GroupIDs []string `json:"group_ids"`
		UserIDs  []string `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_SPACE_DESK_GROUP_REQUEST_DATA)
		return
	}

	// 2) Converte GroupIDs para ObjectIDs
	var groupObjIDs []bson.ObjectID
	for _, gid := range payload.GroupIDs {
		if oid, err := bson.ObjectIDFromHex(gid); err == nil {
			groupObjIDs = append(groupObjIDs, oid)
		}
	}
	if len(groupObjIDs) == 0 {
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
	chatCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)

	for _, groupOID := range groupObjIDs {
		// 4) Busca o grupo atual
		var grp struct {
			UserIDs []string `bson:"user_ids"`
		}
		if err := groupsCol.FindOne(ctx, bson.M{"_id": groupOID}).Decode(&grp); err != nil {
			utils.SendResponse(w, http.StatusNotFound, "", nil, utils.CANNOT_FIND_SPACE_DESK_GROUPS)
			return
		}

		// 5) Merge de user_ids (old_ids)
		idSet := make(map[string]struct{}, len(grp.UserIDs)+len(payload.UserIDs))
		for _, old := range grp.UserIDs {
			idSet[old] = struct{}{}
		}
		for _, old := range payload.UserIDs {
			idSet[old] = struct{}{}
		}
		mergedIDs := make([]string, 0, len(idSet))
		for old := range idSet {
			mergedIDs = append(mergedIDs, old)
		}

		// 6) Busca chats para esses user_ids
		filterIDs := make([]interface{}, len(mergedIDs))
		for i, uid := range mergedIDs {
			filterIDs[i] = uid
		}
		cursor, err := chatCol.Find(ctx, bson.M{"user_id": bson.M{"$in": filterIDs}})
		if err != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
			return
		}
		var raw []bson.M
		if err := cursor.All(ctx, &raw); err != nil {
			cursor.Close(ctx)
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
			return
		}
		cursor.Close(ctx)

		// 7) Extrai apenas os _id em hex
		chatIDs := make([]string, 0, len(raw))
		for _, doc := range raw {
			if oid, ok := doc["_id"].(bson.ObjectID); ok {
				chatIDs = append(chatIDs, oid.Hex())
			}
		}

		// 8) Atualiza o grupo: novos user_ids + lista de chat IDs
		update := bson.M{"$set": bson.M{
			"user_ids":   mergedIDs,
			"chats":      chatIDs,
			"updated_at": time.Now(),
		}}
		if _, err := groupsCol.UpdateOne(ctx, bson.M{"_id": groupOID}, update); err != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_SPACE_DESK_GROUP_TO_MONGODB)
			return
		}
	}

	utils.SendResponse(w, http.StatusOK, "Usuários adicionados e chats atualizados", nil, 0)
}
