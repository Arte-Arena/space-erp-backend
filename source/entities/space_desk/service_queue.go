package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"api/schemas"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetServiceQueue(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	query := r.URL.Query()
	untilStr := query.Get("until")
	limitStr := query.Get("limit")
	pageStr := query.Get("page")

	limit := 20
	page := 1
	if limitParsed, err := strconv.Atoi(limitStr); err == nil && limitParsed > 0 {
		if limitParsed > 100 {
			limit = 100
		} else {
			limit = limitParsed
		}
	}
	if pageParsed, err := strconv.Atoi(pageStr); err == nil && pageParsed > 0 {
		page = pageParsed
	}
	skip := (page - 1) * limit

	minTimestamp := int64(0)
	if untilStr != "" {
		if untilDays, err := strconv.Atoi(untilStr); err == nil && untilDays > 0 {
			minTimestamp = time.Now().Add(-time.Duration(untilDays) * 24 * time.Hour).Unix()
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

	chatCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT_METADATA)
	eventsCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_EVENTS_WHATSAPP)

	findOpts := options.Find().SetSort(bson.D{{Key: "last_message_timestamp", Value: 1}}).SetSkip(int64(skip)).SetLimit(int64(limit))
	filter := bson.M{"closed": false}
	if minTimestamp > 0 {
		filter["last_message_timestamp"] = bson.M{"$gte": time.Unix(minTimestamp, 0)}
	}
	cursor, err := chatCol.Find(ctx, filter, findOpts)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEADS_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	chats := []schemas.SpaceDeskChatMetadata{}
	if err := cursor.All(ctx, &chats); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_PARSE_SPACE_DESK_GROUPS)
		return
	}

	priorityQueue := make([]schemas.SpaceDeskChatMetadata, 0)
	for _, chat := range chats {
		lastMsgUser := ""
		lastMsgFrom := ""

		eventFilter := bson.M{"entry.changes.value.messages.to": chat.ClientPhoneNumber}
		findEventOpts := options.FindOne().SetSort(bson.D{{Key: "entry.changes.value.messages.timestamp", Value: -1}})
		event := schemas.SpaceDeskMessageEvent{}
		err := eventsCol.FindOne(ctx, eventFilter, findEventOpts).Decode(&event)
		if err == nil {
			for _, entry := range event.Entry {
				for _, change := range entry.Changes {
					if len(change.Value.Messages) > 0 {
						msg := change.Value.Messages[len(change.Value.Messages)-1]
						lastMsgUser = msg.User
						lastMsgFrom = msg.From
					}
				}
			}
		}

		if lastMsgFrom == "space-erp-backend" || lastMsgUser != "" {
			continue
		}
		priorityQueue = append(priorityQueue, chat)
	}

	utils.SendResponse(w, http.StatusOK, "Service queue", priorityQueue, 0)
}
