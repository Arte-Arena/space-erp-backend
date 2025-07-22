package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type updateGroupPayload struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	UserIDs []bson.ObjectID `json:"user_ids"`
	Status  string          `json:"status"`
	Type    string          `json:"type"`
}

func UpdateOneGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	// 1) Decodifica payload
	var payload updateGroupPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	// 2) Converte ID do grupo
	groupOID, err := bson.ObjectIDFromHex(payload.ID)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.FUNNELS_INVALID_REQUEST_DATA)
		return
	}

	// 3) Conecta no Mongo
	client, err := mongo.Connect(options.Client().ApplyURI(os.Getenv(utils.MONGODB_URI)))
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	// 4) Monta filtro $in para user_ids
	filterIDs := make([]interface{}, len(payload.UserIDs))
	for i, uid := range payload.UserIDs {
		filterIDs[i] = uid
	}
	chatFilter := bson.M{"user_id": bson.M{"$in": filterIDs}}

	// 5) Busca só os documentos de chat correspondentes
	chatCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
	cursor, err := chatCol.Find(ctx, chatFilter)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
		return
	}
	defer cursor.Close(ctx)

	// 6) Extrai apenas os hex IDs
	var raw []bson.M
	if err := cursor.All(ctx, &raw); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
		return
	}
	var chatIDs []string
	for _, doc := range raw {
		if oid, ok := doc["_id"].(bson.ObjectID); ok {
			chatIDs = append(chatIDs, oid.Hex())
		}
	}

	// 7) Atualiza o grupo apenas definindo o slice de IDs
	update := bson.M{"$set": bson.M{
		"name":     payload.Name,
		"user_ids": payload.UserIDs,
		"status":   payload.Status,
		"type":     payload.Type,
		"chats":    chatIDs,
	}}
	groupsCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_GROUPS)
	if _, err := groupsCol.UpdateOne(ctx, bson.M{"_id": groupOID}, update); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_UPDATE_SPACE_DESK_GROUP_TO_MONGODB)
		return
	}

	// 8) Responde sucesso (sem body ou com chatIDs, se quiser)
	utils.SendResponse(w, http.StatusOK, "Group updated", bson.M{"chats": chatIDs}, 0)
}
