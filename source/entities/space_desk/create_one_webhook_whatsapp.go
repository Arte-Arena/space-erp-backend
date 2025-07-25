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
		broadcastSpaceDeskMessage(event)
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
			last_message        string `bson:"last_message_id"`
			last_message_sender string `bson:"last_message_sender"`
			Name                string `bson:"name"`
		}
		err := collection_chat.FindOne(ctx, filter).Decode(&chatDoc)
		lastMessageIdFromDB := ""
		if err == nil {
			lastMessageIdFromDB = chatDoc.last_message
			// Log do lastMessageId já gravado
			fmt.Println("lastMessageId já gravado:", lastMessageIdFromDB)
		}

		// Atualizar a collection de chat
		if statusMessageId == lastMessageIdFromDB && chatDoc.last_message_sender == "company" {
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
			_, err := collection_chat.UpdateOne(ctx, filter, update, updateOpts)

			if err != nil {
				log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
			}
		}

		// Atualizar a collection de message
		collection_message := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_MESSAGE)
		filter = bson.M{"message_id": messageId}

		update := bson.M{
			"$set": bson.M{
				"status":     status,
				"updated_at": updatedAt,
			},
		}

		updateOpts := options.UpdateOne()
		_, err = collection_message.UpdateOne(ctx, filter, update, updateOpts)

		if err != nil {
			log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
		}

	} else {
		event["from"] = "client"
		broadcastSpaceDeskMessage(event)

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

		now := time.Now().UTC()
		var lastMessageFromClient string
		msgType, _ := messages["type"].(string)

		if msgType, ok := messages["type"].(string); ok && msgType == "text" {
			if textObj, ok := messages["text"].(map[string]interface{}); ok {
				if body, ok := textObj["body"].(string); ok {
					lastMessageFromClient = body
				}
			}
		}

		// colocar nos chats o id do lead (caso não encotre ele cria um novo lead) e salva o id do lead no chat
		collectionLeads := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_LEADS)
		leadFilter := bson.M{"phone": clientPhoneNumber}

		var leadDoc struct {
			ID bson.ObjectID `bson:"_id"`
		}
		err := collectionLeads.FindOne(ctx, leadFilter).Decode(&leadDoc)
		isNewLead := false

		if err != nil {
			// Criar novo lead se não existir
			newLead := schemas.Lead{
				Phone:     clientPhoneNumber,
				Name:      name,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Source:    "SpaceDesk",
			}
			insertResult, err := collectionLeads.InsertOne(ctx, newLead)
			if err != nil {
				log.Printf("Erro ao criar lead: %v", err)
			} else {
				leadDoc.ID = insertResult.InsertedID.(bson.ObjectID)
				isNewLead = true
			}
		}

		// 2. Atualizar chat COM leadID
		filter := bson.M{"cliente_phone_number": clientPhoneNumber, "company_phone_number": companyPhoneNumber}
		update := bson.M{
			"$set": bson.M{
				"name":                               name,
				"last_message_id":                    messageFromClientId,
				"last_message_timestamp":             fmt.Sprint(now.Unix()),
				"last_message_excerpt":               lastMessageFromClient,
				"last_message_type":                  msgType,
				"last_message_sender":                "client",
				"last_message_from_client_timestamp": fmt.Sprint(now.Unix()),
				"updated_at":                         updatedAt,
				"closed":                             false,
				"lead_id":                            leadDoc.ID, // Usar ID real
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
			log.Printf("Erro ao atualizar chat: %v", err)
		}

		// 3. Atualizar lead com platformId se for novo
		if isNewLead {
			var chatID bson.ObjectID
			if updateResult.UpsertedID != nil {
				chatID = updateResult.UpsertedID.(bson.ObjectID)
			} else {
				var chatDoc struct {
					ID bson.ObjectID `bson:"_id"`
				}
				if err := collection_chat.FindOne(ctx, filter).Decode(&chatDoc); err == nil {
					chatID = chatDoc.ID
				}
			}

			if !chatID.IsZero() {
				_, err := collectionLeads.UpdateByID(
					ctx,
					leadDoc.ID,
					bson.M{"$set": bson.M{"platform_id": chatID.Hex()}},
				)
				if err != nil {
					log.Printf("Erro ao atualizar lead: %v", err)
				}
			}
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

		if msgType, ok := messages["type"].(string); ok {
			filter = bson.M{"message_id": messageFromClientId}

			// monta o update de acordo com o tipo
			switch msgType {
			case "text":
				if textObj, ok := messages["text"].(map[string]interface{}); ok {
					if body, ok := textObj["body"].(string); ok {
						update = bson.M{
							"$set": bson.M{
								"chat_id": chatId,
								"type":    "text",
								"body":    body,
							},
							"$setOnInsert": bson.M{"created_at": updatedAt},
						}
					}
				}

			case "image", "video", "audio", "document":
				media := messages[msgType].(map[string]interface{})
				base := bson.M{
					"chat_id":   chatId,
					"type":      msgType,
					"media_id":  media["id"].(string),
					"mime_type": media["mime_type"].(string),
					"sha256":    media["sha256"].(string),
				}
				if cap, ok := media["caption"].(string); ok {
					base["caption"] = cap
				}
				update = bson.M{
					"$set":         base,
					"$setOnInsert": bson.M{"created_at": updatedAt},
				}

			case "sticker":
				st := messages["sticker"].(map[string]interface{})
				update = bson.M{
					"$set": bson.M{
						"chat_id":   chatId,
						"type":      "sticker",
						"media_id":  st["id"].(string),
						"mime_type": st["mime_type"].(string),
						"sha256":    st["sha256"].(string),
					},
					"$setOnInsert": bson.M{"created_at": updatedAt},
				}

			case "reaction":
				react := messages["reaction"].(map[string]interface{})
				update = bson.M{
					"$set": bson.M{
						"chat_id":             chatId,
						"type":                "reaction",
						"reaction_emoji":      react["emoji"].(string),
						"reaction_message_id": react["message_id"].(string),
					},
					"$setOnInsert": bson.M{"created_at": updatedAt},
				}

			case "interactive":
				inter := messages["interactive"].(map[string]interface{})
				base := bson.M{
					"chat_id":          chatId,
					"type":             "interactive",
					"interactive_type": inter["type"].(string),
				}
				if btn, ok := inter["button_reply"].(map[string]interface{}); ok {
					base["reply_id"] = btn["id"].(string)
					base["reply_text"] = btn["title"].(string)
				}
				if lst, ok := inter["list_reply"].(map[string]interface{}); ok {
					base["reply_id"] = lst["id"].(string)
					base["reply_text"] = lst["title"].(string)
				}
				update = bson.M{"$set": base, "$setOnInsert": bson.M{"created_at": updatedAt}}

			case "location":
				loc := messages["location"].(map[string]interface{})
				update = bson.M{
					"$set": bson.M{
						"chat_id":   chatId,
						"type":      "location",
						"latitude":  loc["latitude"],
						"longitude": loc["longitude"],
						"name":      loc["name"].(string),
						"address":   loc["address"].(string),
					},
					"$setOnInsert": bson.M{"created_at": updatedAt},
				}

			case "contacts":
				update = bson.M{
					"$set": bson.M{
						"chat_id":  chatId,
						"type":     "contacts",
						"contacts": messages["contacts"],
					},
					"$setOnInsert": bson.M{"created_at": updatedAt},
				}

			case "template":
				tmpl := messages["template"].(map[string]interface{})
				update = bson.M{
					"$set": bson.M{
						"chat_id":       chatId,
						"type":          "template",
						"template_name": tmpl["name"].(string),
						"language":      tmpl["language"].(string),
						"components":    tmpl["components"],
					},
					"$setOnInsert": bson.M{"created_at": updatedAt},
				}

			default:
				log.Printf("[CreateOneWebhookWhatsapp] Tipo não tratado: %s", msgType)
				return
			}

			// campos comuns a todos os tipos
			s := update["$set"].(bson.M)
			s["message_from_client_timestamp"] = fmt.Sprint(now.Unix())
			s["message_id"] = messageFromClientId
			s["by"] = clientPhoneNumber
			s["updated_at"] = updatedAt

			// Se houver contexto (reply/citação), salva também
			if ctxObj, ok := messages["context"].(map[string]interface{}); ok {
				ctxSet := bson.M{
					"message_id": ctxObj["id"].(string),
					"from":       ctxObj["from"].(string),
				}
				if t, ok2 := ctxObj["type"].(string); ok2 {
					ctxSet["type"] = t
				}
				s["context"] = ctxSet
			}

			updateOpts := options.UpdateOne().SetUpsert(true)
			updateResult, err := collection_message.UpdateOne(ctx, filter, update, updateOpts)
			fmt.Println("UpdateOne result (message):", updateResult)
			if err != nil {
				log.Printf("[CreateOneWebhookWhatsapp] Error inserting event into MongoDB: %v", err)
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
