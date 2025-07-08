package spacedesk

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"fmt"
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

// CreateOneWebhookWhatsapp cria um novo webhook do WhatsApp
func CreateOneWebhookWhatsapp(w http.ResponseWriter, r *http.Request) {

	utils.SendResponse(w, http.StatusOK, "", nil, 0)

	event := make(map[string]any)
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Printf("[CreateOneWebhookWhatsapp] Error decoding request body: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()
	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		log.Printf("[CreateOneWebhookWhatsapp] Error connecting to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	collectionEvents := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_EVENTS_WHATSAPP)

	_, err = collectionEvents.InsertOne(ctx, event)
	if err != nil {
		log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
	}

	_, err = collectionEvents.InsertOne(ctx, event)
	if err != nil {
		log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
	}

	broadcastSpaceDeskMessage(event)

	collection_chat := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)

	entryArr, ok := event["entry"].([]interface{})
	if !ok || len(entryArr) == 0 {
		log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
	}
	entry, ok := entryArr[0].(map[string]interface{})
	if !ok {
		log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
	}
	changesArr, ok := entry["changes"].([]interface{})
	if !ok || len(changesArr) == 0 {
		log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
	}
	changes, ok := changesArr[0].(map[string]interface{})
	if !ok {
		log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
	}
	value, ok := changes["value"].(map[string]interface{})
	if !ok {
		log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
	}

	metadata, ok := value["metadata"].(map[string]interface{})
	if !ok {
		log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
	}

	companyPhoneNumber, ok := metadata["display_phone_number"].(string)
	if !ok {
		log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
	}

	updatedAt := time.Now()

	// Definir name e clientPhoneNumber antes dos blocos condicionais
	name := ""
	clientPhoneNumber := ""

	log.Println("---------------------->")
	log.Println(value)
	log.Println("<---------------------")

	if statusesArr, ok := value["statuses"].([]any); ok {

		fmt.Println("passei pelo status.")

		fmt.Println(statusesArr)

		// Parseando os dados de status

		statuses, ok := statusesArr[0].(map[string]any)
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}

		messageId, ok := statuses["id"].(string)
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}

		clientPhoneNumber, ok = statuses["recipient_id"].(string)
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}

		status, ok := statuses["status"].(string)
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}

		statusTimestamp, ok := statuses["timestamp"].(string)
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}

		statusMessageId, ok := statuses["id"].(string)
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}

		filter := bson.M{"cliente_phone_number": clientPhoneNumber, "company_phone_number": companyPhoneNumber}

		// Buscar o lastMessageId já gravado na collection de chat
		var chatDoc struct {
			LastMessageIdFromClient string `bson:"last_message_id"`
			Name                    string `bson:"name"`
		}
		err := collection_chat.FindOne(ctx, filter).Decode(&chatDoc)
		lastMessageIdFromDB := ""
		if err == nil {
			lastMessageIdFromDB = chatDoc.LastMessageIdFromClient
			// Log do lastMessageId já gravado
			fmt.Println("lastMessageId já gravado:", lastMessageIdFromDB)
		}

		// Atualizar a collection de chat

		// fmt.Println("============>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> statusMessageId:", statusMessageId)

		fmt.Printf("statusMessageId: %T\n", statusMessageId)
		fmt.Printf("lastMessageIdFromDB: %T\n", lastMessageIdFromDB)

		if statusMessageId == lastMessageIdFromDB {
			update := bson.M{
				"$set": bson.M{
					"last_status_from_company_related_to_message_id": bson.M{
						"timestamp": statusTimestamp,
						"value":     status,
					},
					"updated_at": updatedAt,
				},
			}
			updateOpts := options.UpdateOne().SetUpsert(true)
			updateResult, err := collection_chat.UpdateOne(ctx, filter, update, updateOpts)

			fmt.Println("UpdateOne result (status):", updateResult)
			fmt.Println("------------------------>>> status:", status)

			if err != nil {
				log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
			}
		}

		// Atualizar a collection de message

		fmt.Println("------------------>>        passei pelo message para atualizar a collection de message")

		collection_message := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_MESSAGE)

		filter = bson.M{"message_id": messageId}

		update := bson.M{
			"$set": bson.M{
				"status":     status,
				"updated_at": updatedAt,
			},
		}

		updateOpts := options.UpdateOne()
		updateResult, err := collection_message.UpdateOne(ctx, filter, update, updateOpts)

		log.Println("---------------->>>>>            UpdateOne result (message):", updateResult)

		if err != nil {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}

	} else {

		// Criar ou Atualizar a collection de chat (quando chega uma mensagem pelo webhook)

		fmt.Println("passei pela mensagem.")

		contactsArr, ok := value["contacts"].([]interface{})
		if !ok || len(contactsArr) == 0 {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}
		contacts, ok := contactsArr[0].(map[string]interface{})
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}
		messagesArr, ok := value["messages"].([]interface{})
		if !ok || len(messagesArr) == 0 {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}
		messages, ok := messagesArr[0].(map[string]interface{})
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}
		messageFromClientId, ok := messages["id"].(string)
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}

		clientPhoneNumber, ok = contacts["wa_id"].(string)
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}
		profile, ok := contacts["profile"].(map[string]interface{})
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}
		name, ok = profile["name"].(string)
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}
		timestampStr, ok := messages["timestamp"].(string)
		if !ok {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}
		lastMessageTimestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}

		lastMessageFromClientTimestamp := time.Unix(lastMessageTimestampInt, 0)

		var lastMessageFromClient string

		if msgType, ok := messages["type"].(string); ok && msgType == "text" {
			if textObj, ok := messages["text"].(map[string]interface{}); ok {
				if body, ok := textObj["body"].(string); ok {
					lastMessageFromClient = body
				}
			}
		}

		filter := bson.M{"cliente_phone_number": clientPhoneNumber, "company_phone_number": companyPhoneNumber}
		update := bson.M{
			"$set": bson.M{
				"name":                               name,
				"last_message_id":                    messageFromClientId,
				"last_message_excerpt":               lastMessageFromClient,
				"last_message_sender":                "client",
				"last_message_from_client_timestamp": lastMessageFromClientTimestamp,
				"updated_at":                         updatedAt,
			},
			"$setOnInsert": bson.M{
				"company_phone_number": companyPhoneNumber,
				"cliente_phone_number": clientPhoneNumber,
				"nick_name":            "",
				"user_id":              "",
				"description":          "",
				"created_at":           updatedAt,
			},
		}
		updateOpts := options.UpdateOne().SetUpsert(true)
		updateResult, err := collection_chat.UpdateOne(ctx, filter, update, updateOpts)

		if err != nil {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}

		// Criar ou Atualizar a collection de message

		collection_message := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_MESSAGE)

		var chatId interface{}

		if updateResult.UpsertedID != nil {
			chatId = updateResult.UpsertedID
		} else {
			// Buscar o _id do chat pelo filtro (sempre vai achar se não foi upsert)
			var chat struct {
				ID interface{} `bson:"_id"`
			}
			err := collection_chat.FindOne(ctx, filter).Decode(&chat)
			if err == nil {
				chatId = chat.ID
			}
		}

		if textObj, ok := messages["text"].(map[string]interface{}); ok {
			if body, ok := textObj["body"].(string); ok {

				filter = bson.M{"message_id": messageFromClientId}

				update = bson.M{
					"$set": bson.M{
						"chat_id":                       chatId,
						"type":                          "text",
						"message_from_client_timestamp": lastMessageFromClientTimestamp,
						"message_id":                    messageFromClientId,
						"body":                          body,
						"from":                          "client",
						"by":                            clientPhoneNumber,
					},
					"$setOnInsert": bson.M{
						"created_at": updatedAt,
					},
				}

				updateOpts := options.UpdateOne().SetUpsert(true)
				updateResult, err := collection_message.UpdateOne(ctx, filter, update, updateOpts)
				fmt.Println("UpdateOne result (message):", updateResult)

				if err != nil {
					log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
				}
			}
		}

		// Verificar se o chat é um novo Lead

		chatID := bson.ObjectID{}
		if updateResult.UpsertedID != nil {
			chatID = updateResult.UpsertedID.(bson.ObjectID)
		} else {
			var metadata struct {
				ID bson.ObjectID `bson:"_id"`
			}
			if err := collection_chat.FindOne(ctx, filter).Decode(&metadata); err == nil {
				chatID = metadata.ID
			}
		}

		if !chatID.IsZero() {
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
							PlatformId: chatID.Hex(),
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

	}

}
