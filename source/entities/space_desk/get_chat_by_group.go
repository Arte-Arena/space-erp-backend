package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"net/http"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetChatsFromGroup(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	groupId := r.PathValue("groupId")
	if groupId == "" {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	groupObjID, err := bson.ObjectIDFromHex(groupId)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		return
	}

	limit := int64(10)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database(database.GetDB())
	groupCol := db.Collection("groups")
	chatCol := db.Collection(database.COLLECTION_SPACE_DESK_CHAT)

	// Busca o grupo
	var group struct {
		Chats []string `bson:"chats"`
	}
	err = groupCol.FindOne(ctx, bson.M{"_id": groupObjID}).Decode(&group)
	if err != nil {
		utils.SendResponse(w, http.StatusNotFound, "Grupo não encontrado.", nil, utils.NOT_FOUND)
		return
	}
	if len(group.Chats) == 0 {
		utils.SendResponse(w, http.StatusOK, "Sem chats neste grupo.", []any{}, 0)
		return
	}

	var chatObjIDs []bson.ObjectID
	for _, idStr := range group.Chats {
		objID, err := bson.ObjectIDFromHex(idStr)
		if err == nil {
			chatObjIDs = append(chatObjIDs, objID)
		}
	}
	if len(chatObjIDs) == 0 {
		utils.SendResponse(w, http.StatusOK, "Nenhum chat válido encontrado no grupo.", []any{}, 0)
		return
	}

	filter := bson.M{"_id": bson.M{"$in": chatObjIDs}}

	cur, err := chatCol.Find(ctx, filter, options.Find().SetLimit(limit))
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.NOT_FOUND)
		return
	}
	defer cur.Close(ctx)

	var chats []bson.M
	if err := cur.All(ctx, &chats); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.NOT_FOUND)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", chats, 0)
}
