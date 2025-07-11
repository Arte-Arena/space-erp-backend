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

func buildServiceQueueFilterFromQueryParams(r *http.Request) bson.M {
	query := r.URL.Query()
	filter := bson.M{"closed": false}

	if name := query.Get("name"); name != "" {
		filter["name"] = bson.M{"$regex": name, "$options": "i"}
	}

	if number := query.Get("number"); number != "" {
		filter["client_phone_number"] = bson.M{"$regex": number, "$options": "i"}
	}

	if untilStr := query.Get("until"); untilStr != "" {
		if untilDays, err := strconv.Atoi(untilStr); err == nil && untilDays > 0 {
			minTimestamp := time.Now().Add(-time.Duration(untilDays) * 24 * time.Hour).Unix()
			filter["last_message_timestamp"] = bson.M{"$gte": time.Unix(minTimestamp, 0)}
		}
	}

	return filter
}

func buildServiceQueueV2FilterFromQueryParams(r *http.Request) bson.M {
	query := r.URL.Query()
	filter := bson.M{
		"last_message_sender": "client",
	}

	if name := query.Get("name"); name != "" {
		filter["name"] = bson.M{"$regex": name, "$options": "i"}
	}

	if number := query.Get("number"); number != "" {
		filter["cliente_phone_number"] = bson.M{"$regex": number, "$options": "i"}
	}

	if untilStr := query.Get("until"); untilStr != "" {
		if untilDays, err := strconv.Atoi(untilStr); err == nil && untilDays > 0 {
			minTimestamp := time.Now().Add(-time.Duration(untilDays) * 24 * time.Hour)
			filter["last_message_from_client_timestamp"] = bson.M{"$gte": minTimestamp}
		}
	}

	return filter
}

func GetServiceQueue(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	query := r.URL.Query()
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

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	chatCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
	eventsCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_EVENTS_WHATSAPP)

	findOpts := options.Find().SetSort(bson.D{{Key: "last_message_timestamp", Value: -1}}).SetSkip(int64(skip)).SetLimit(int64(limit))
	filter := buildServiceQueueFilterFromQueryParams(r)
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

		eventFilter := bson.M{
			"$or": []bson.M{
				{"entry.changes.value.contacts.wa_id": chat.ClientPhoneNumber},
				{"entry.changes.value.messages.to": chat.ID.Hex()},
			},
		}

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

	for i, j := 0, len(priorityQueue)-1; i < j; i, j = i+1, j-1 {
		priorityQueue[i], priorityQueue[j] = priorityQueue[j], priorityQueue[i]
	}

	utils.SendResponse(w, http.StatusOK, "Service queue", priorityQueue, 0)
}

func GetServiceQueueV2(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	query := r.URL.Query()
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

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	chatCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)

	findOpts := options.Find().
		SetSort(bson.D{{Key: "last_message_from_client_timestamp", Value: 1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	filter := buildServiceQueueV2FilterFromQueryParams(r)
	cursor, err := chatCol.Find(ctx, filter, findOpts)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEADS_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	chats := []schemas.SpaceDeskChat{}
	if err := cursor.All(ctx, &chats); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_PARSE_SPACE_DESK_GROUPS)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", chats, 0)
}
