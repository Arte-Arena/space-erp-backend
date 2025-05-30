package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

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

	// 3) upsert em chat_metadata
	metaCol := mongoClient.Database(database.GetDB()).Collection("chat_metadata")

	// extrai phoneNumber (wa_id) e timestamp da última mensagem
	var phoneNumber string
	var lastTs time.Time

	if entries, ok := event["entry"].([]any); ok && len(entries) > 0 {
		if entry, ok := entries[0].(map[string]any); ok {
			if changes, ok := entry["changes"].([]any); ok && len(changes) > 0 {
				if change, ok := changes[0].(map[string]any); ok {
					if value, ok := change["value"].(map[string]any); ok {
						// phoneNumber
						if contacts, ok := value["contacts"].([]any); ok && len(contacts) > 0 {
							if c0, ok := contacts[0].(map[string]any); ok {
								if wa, ok := c0["wa_id"].(string); ok {
									phoneNumber = wa
								}
							}
						}
						// lastMessageTimestamp
						if msgs, ok := value["messages"].([]any); ok && len(msgs) > 0 {
							if m := msgs[len(msgs)-1].(map[string]any); ok {
								if tsStr, ok := m["timestamp"].(string); ok {
									if secs, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
										lastTs = time.Unix(secs, 0).UTC()
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// monta filtro e upsert
	filter := bson.M{"phoneNumber": phoneNumber}
	update := bson.M{
		"$setOnInsert": bson.M{
			"name":        "",
			"nickName":    "",
			"thumbUrl":    "",
			"phoneNumber": phoneNumber,
			"description": "",
			"status":      "",
			"type":        "",
			"groupId":     "",
			"createdAt":   time.Now().UTC(),
		},
		"$set": bson.M{
			"lastMessageTimestamp": lastTs,
			"updatedAt":            time.Now().UTC(),
		},
	}
	upsertOpts := options.UpdateOne().SetUpsert(true)
	if _, err := metaCol.UpdateOne(ctx, filter, update, upsertOpts); err != nil {
		log.Printf("[CreateOneWebhookWhatsapp] erro no upsert chat_metadata: %v", err)
	}

	// 4) broadcast e resposta
	broadcastSpaceDeskMessage(event)

	utils.SendResponse(w, http.StatusCreated, "", nil, 0)
}
