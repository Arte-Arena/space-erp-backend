package spacedesk

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateOneWebhookWhatsapp(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	event := make(map[string]any)
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_EVENTS_WHATSAPP)

	_, err = collection.InsertOne(ctx, event)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}

	collection2 := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT_METADATA)

	entryArr, ok := event["entry"].([]interface{})
	if !ok || len(entryArr) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}
	entry, ok := entryArr[0].(map[string]interface{})
	if !ok {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}
	changesArr, ok := entry["changes"].([]interface{})
	if !ok || len(changesArr) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}
	changes, ok := changesArr[0].(map[string]interface{})
	if !ok {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}
	value, ok := changes["value"].(map[string]interface{})
	if !ok {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}
	contactsArr, ok := value["contacts"].([]interface{})
	if !ok || len(contactsArr) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}
	contacts, ok := contactsArr[0].(map[string]interface{})
	if !ok {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}
	messagesArr, ok := value["messages"].([]interface{})
	if !ok || len(messagesArr) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}
	messages, ok := messagesArr[0].(map[string]interface{})
	if !ok {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}

	clientPhoneNumber, ok := contacts["wa_id"].(string)
	if !ok {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}
	profile, ok := contacts["profile"].(map[string]interface{})
	if !ok {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}
	name, ok := profile["name"].(string)
	if !ok {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}
	timestampStr, ok := messages["timestamp"].(string)
	if !ok {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}
	lastMessageTimestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.CANNOT_INSERT_SPACE_DESK_EVENT_TO_MONGODB)
		return
	}

	lastMessageTimestamp := time.Unix(lastMessageTimestampInt, 0)
	updatedAt := time.Now()
	var lastMessage string
	if msgType, ok := messages["type"].(string); ok && msgType == "text" {
		if textObj, ok := messages["text"].(map[string]interface{}); ok {
			if body, ok := textObj["body"].(string); ok {
				lastMessage = body
			}
		}
	}

	filter := bson.M{"cliente_phone_number": clientPhoneNumber}
	update := bson.M{
		"$set": bson.M{
			"name":                   name,
			"updated_at":             updatedAt,
			"last_message_timestamp": lastMessageTimestamp,
			"last_message":           lastMessage,
			"need_template":          false,
		},
		"$setOnInsert": bson.M{
			"nick_name":            "",
			"cliente_phone_number": clientPhoneNumber,
			"user_id":              "",
			"description":          "",
			"type":                 "",
			"created_at":           updatedAt,
		},
	}

	updateOpts := options.UpdateOne().SetUpsert(true)
	updateResult, err := collection2.UpdateOne(ctx, filter, update, updateOpts)

	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_SPACE_DESK_CHAT_METADATA_TO_MONGODB)
		return
	}

	broadcastSpaceDeskMessage(event)

	if statusesArr, ok := value["statuses"].([]any); ok && len(statusesArr) > 0 {
		broadcastSpaceDeskMessage(map[string]any{"statuses": statusesArr})
	}

	chatMetadataID := bson.ObjectID{}
	if updateResult.UpsertedID != nil {
		chatMetadataID = updateResult.UpsertedID.(bson.ObjectID)
	} else {
		var metadata struct {
			ID bson.ObjectID `bson:"_id"`
		}
		if err := collection2.FindOne(ctx, filter).Decode(&metadata); err == nil {
			chatMetadataID = metadata.ID
		}
	}

	if !chatMetadataID.IsZero() {
		var rdb *redis.Client
		redisURI := os.Getenv("REDIS_URI")
		opts, err := redis.ParseURL(redisURI)
		if err == nil {
			rdb = redis.NewClient(opts)
			defer rdb.Close()
		}

		redisKey := "spacedesk:lead:phone:" + clientPhoneNumber
		cached := false
		if rdb != nil {
			if err := rdb.Get(ctx, redisKey).Err(); err == nil {
				cached = true
			}
		}

		if !cached {
			collectionLeads := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_LEADS)
			leadFilter := bson.M{"phone": clientPhoneNumber}
			count, err := collectionLeads.CountDocuments(ctx, leadFilter)

			if err == nil {
				shouldCache := false
				if count == 0 {
					newLead := schemas.Lead{
						Name:       name,
						Phone:      clientPhoneNumber,
						Source:     "SpaceDesk",
						PlatformId: chatMetadataID.Hex(),
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}
					if _, err := collectionLeads.InsertOne(ctx, newLead); err == nil {
						shouldCache = true
					}
				} else {
					shouldCache = true
				}

				if shouldCache && rdb != nil {
					expiration := 90 * 24 * time.Hour
					rdb.Set(ctx, redisKey, "1", expiration)
				}
			}
		}
	}

	utils.SendResponse(w, http.StatusCreated, "", nil, 0)
}
