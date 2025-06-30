package spacedesk

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"log"
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
	// Primeiro, valida o body ANTES de responder 200
	event := make(map[string]any)
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	// Responde 200 OK imediatamente
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

	// Processamento real em background
	go func(event map[string]any) {
		// Evita panics de travar o servidor silenciosamente
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC][WEBHOOK] %v\n", r)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
		defer cancel()

		mongoURI := os.Getenv(utils.MONGODB_URI)
		opts := options.Client().ApplyURI(mongoURI)
		mongoClient, err := mongo.Connect(opts)
		if err != nil {
			log.Printf("[WEBHOOK][ERROR] Cannot connect to MongoDB: %v", err)
			return
		}
		defer mongoClient.Disconnect(ctx)

		collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_EVENTS_WHATSAPP)
		_, err = collection.InsertOne(ctx, event)
		if err != nil {
			log.Printf("[WEBHOOK][ERROR] Cannot insert event to MongoDB: %v", err)
			return
		}

		// Notifica websocket, ignora retorno/erro
		defer func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[WEBHOOK][PANIC] broadcastSpaceDeskMessage: %v", r)
				}
			}()
			broadcastSpaceDeskMessage(event)
		}()

		collection2 := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT_METADATA)

		entryArr, ok := event["entry"].([]interface{})
		if !ok || len(entryArr) == 0 {
			log.Printf("[WEBHOOK][ERROR] missing entry")
			return
		}
		entry, ok := entryArr[0].(map[string]interface{})
		if !ok {
			log.Printf("[WEBHOOK][ERROR] entry not map")
			return
		}
		changesArr, ok := entry["changes"].([]interface{})
		if !ok || len(changesArr) == 0 {
			log.Printf("[WEBHOOK][ERROR] missing changes")
			return
		}
		changes, ok := changesArr[0].(map[string]interface{})
		if !ok {
			log.Printf("[WEBHOOK][ERROR] changes not map")
			return
		}
		value, ok := changes["value"].(map[string]interface{})
		if !ok {
			log.Printf("[WEBHOOK][ERROR] value not map")
			return
		}

		metadata, ok := value["metadata"].(map[string]interface{})
		if !ok {
			log.Printf("[WEBHOOK][ERROR] missing metadata")
			return
		}

		companyPhoneNumber, ok := metadata["display_phone_number"].(string)
		if !ok {
			log.Printf("[WEBHOOK][ERROR] missing companyPhoneNumber")
			return
		}

		contactsArr, ok := value["contacts"].([]interface{})
		if !ok || len(contactsArr) == 0 {
			log.Printf("[WEBHOOK][ERROR] missing contacts")
			return
		}
		contacts, ok := contactsArr[0].(map[string]interface{})
		if !ok {
			log.Printf("[WEBHOOK][ERROR] contacts not map")
			return
		}
		messagesArr, ok := value["messages"].([]interface{})
		if !ok || len(messagesArr) == 0 {
			log.Printf("[WEBHOOK][ERROR] missing messages")
			return
		}
		messages, ok := messagesArr[0].(map[string]interface{})
		if !ok {
			log.Printf("[WEBHOOK][ERROR] messages not map")
			return
		}

		clientPhoneNumber, ok := contacts["wa_id"].(string)
		if !ok {
			log.Printf("[WEBHOOK][ERROR] missing clientPhoneNumber")
			return
		}
		profile, ok := contacts["profile"].(map[string]interface{})
		if !ok {
			log.Printf("[WEBHOOK][ERROR] profile not map")
			return
		}
		name, ok := profile["name"].(string)
		if !ok {
			log.Printf("[WEBHOOK][ERROR] missing name")
			return
		}
		timestampStr, ok := messages["timestamp"].(string)
		if !ok {
			log.Printf("[WEBHOOK][ERROR] missing timestamp")
			return
		}
		lastMessageTimestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			log.Printf("[WEBHOOK][ERROR] invalid timestamp: %v", err)
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
				"company_phone_number":   companyPhoneNumber,
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
			log.Printf("[WEBHOOK][ERROR] Cannot insert/update chat metadata: %v", err)
			return
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

		// --- Caching lead no Redis e inserção na collection de leads, igual ao original ---
		if !chatMetadataID.IsZero() {
			var rdb *redis.Client
			redisURI := os.Getenv("REDIS_URI")
			redisOpts, err := redis.ParseURL(redisURI)
			if err == nil {
				rdb = redis.NewClient(redisOpts)
				defer rdb.Close()
			} else {
				log.Printf("[WEBHOOK][WARN] Failed to parse REDIS_URI: %v", err)
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

				if err != nil {
					log.Printf("[WEBHOOK][ERROR] Lead count fail: %v", err)
				} else {
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
						} else {
							log.Printf("[WEBHOOK][ERROR] Lead insert fail: %v", err)
						}
					} else {
						shouldCache = true
					}

					if shouldCache && rdb != nil {
						expiration := 90 * 24 * time.Hour
						if err := rdb.Set(ctx, redisKey, "1", expiration).Err(); err != nil {
							log.Printf("[WEBHOOK][ERROR] Redis set fail: %v", err)
						}
					}
				}
			}
		}
	}(event)
}
