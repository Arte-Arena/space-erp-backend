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

func buildFilterFromQueryParams(r *http.Request) (bson.M, bool) {
	query := r.URL.Query()
	filter := bson.M{}

	if name := query.Get("name"); name != "" {
		filter["name"] = bson.M{"$regex": name, "$options": "i"}
	}

	if number := query.Get("number"); number != "" {
		filter["cliente_phone_number"] = bson.M{"$regex": number, "$options": "i"}
	}

	if status := query.Get("status"); status != "" {
		switch status {
		case "closed":
			filter["closed"] = true
		case "opened":
			filter["closed"] = false
		default:
			return nil, true
		}
	}

	if untilStr := query.Get("until"); untilStr != "" {
		if untilDays, err := strconv.Atoi(untilStr); err == nil && untilDays > 0 {
			minTimestamp := time.Now().Add(-time.Duration(untilDays) * 24 * time.Hour)
			filter["created_at"] = bson.M{"$gte": minTimestamp}
		}
	}

	return filter, false
}

func GetAllChats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	query := r.URL.Query()
	limitStr := query.Get("limit")
	pageStr := query.Get("page")
	numbers := query["numbers[]"]

	// Definindo valores padrão
	limit := 100
	page := 1

	// Parse do "limit"
	if limitParsed, err := strconv.Atoi(limitStr); err == nil && limitParsed > 0 {
		if limitParsed > 100 {
			limit = 100
		} else {
			limit = limitParsed
		}
	}

	// Parse do "page"
	if pageParsed, err := strconv.Atoi(pageStr); err == nil && pageParsed > 0 {
		page = pageParsed
	}

	skip := (page - 1) * limit

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT_METADATA)

	filter, hasError := buildFilterFromQueryParams(r)
	if hasError {
		utils.SendResponse(w, http.StatusBadRequest, "Parâmetro 'status' inválido. Use 'closed' ou 'opened'", nil, 0)
		return
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

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

		// Atualiza o status do chat caso tenha mais de 24 horas
		if !chat.LastMessage.IsZero() && chat.LastMessage.Before(twentyFourHoursAgo) {
			if !chat.NeedTemplate {
				chat.NeedTemplate = true
			}

			updateFilter := bson.M{"_id": chat.ID}
			updateData := bson.M{
				"$set": bson.M{
					"need_template": true,
					"updated_at":    time.Now(),
				},
			}

			_, err := collection.UpdateOne(ctx, updateFilter, updateData)
			if err != nil {
				log.Printf("Erro ao tentar encerrar o chat %s: %v", chat.ID.Hex(), err)
			}
		}

		chats = append(chats, chat)
	}

	// Se numbers[] foi passado, filtra apenas os chats cujos números, só com dígitos, batem com algum
	if len(numbers) > 0 {
		numbersMap := make(map[string]bool)
		for _, n := range numbers {
			numbersMap[n] = true
		}
		filteredChats := make([]schemas.SpaceDeskChatMetadata, 0, len(chats))
		for _, chat := range chats {
			phoneDigits := onlyDigits(chat.ClientPhoneNumber)
			if numbersMap[phoneDigits] {
				filteredChats = append(filteredChats, chat)
			}
		}
		chats = filteredChats
	}

	utils.SendResponse(w, http.StatusOK, "Chats encontrados com sucesso", chats, 0)
}

// GET /api/chats?limit=10&page=2&until=30
// GET /api/chats?numbers[]=11999998888&numbers[]=65999998888