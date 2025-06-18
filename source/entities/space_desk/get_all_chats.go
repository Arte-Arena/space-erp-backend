package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"api/schemas"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetAllChats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	untilStr := r.URL.Query().Get("until")

	minTimestamp := int64(0)
	if untilStr != "" {
		if untilDays, err := strconv.Atoi(untilStr); err == nil && untilDays > 0 {
			minTimestamp = time.Now().Add(-time.Duration(untilDays) * 24 * time.Hour).Unix()
		}
	}

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT_METADATA)

	filter := bson.M{}
	if minTimestamp > 0 {
		filter["created_at"] = bson.M{"$gte": time.Unix(minTimestamp, 0)}
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}})
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEADS_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	chats := []schemas.SpaceDeskChatMetadata{}
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	for cursor.Next(ctx) {
		var chat schemas.SpaceDeskChatMetadata
		if err := cursor.Decode(&chat); err != nil {
			continue
		}

		//Atualiza o status do chat caso tenha mais de 24 horas
		if !chat.LastMessage.IsZero() && chat.LastMessage.Before(twentyFourHoursAgo) && chat.Status != "inactive" {
			chat.Status = "inactive"
			updateFilter := bson.M{"_id": chat.ID}
			updateData := bson.M{"$set": bson.M{
				"status":     "inactive",
				"updated_at": time.Now(),
			}}
			go func(chatId bson.ObjectID) {
				updateCtx, updateCancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer updateCancel()
				_, err := collection.UpdateOne(updateCtx, updateFilter, updateData)
				if err != nil {
					log.Printf("Erro ao tentar encerrar o chat %s: %v", chatId.Hex(), err)
				}
			}(chat.ID)
		}
		chats = append(chats, chat)
	}

	utils.SendResponse(w, http.StatusOK, "", chats, 0)
}
