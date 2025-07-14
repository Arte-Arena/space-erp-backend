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

	if userId := query.Get("user_id"); userId != "" {
		filter["user_id"] = userId
	}

	if ids, ok := query["ids[]"]; ok && len(ids) > 0 {
		objectIDs := make([]bson.ObjectID, 0, len(ids))
		for _, idStr := range ids {
			objID, err := bson.ObjectIDFromHex(idStr)
			if err == nil {
				objectIDs = append(objectIDs, objID)
			}
		}
		if len(objectIDs) > 0 {
			filter["_id"] = bson.M{"$in": objectIDs}
		} else {
			// Se nenhum ID válido, forçar resultado vazio
			filter["_id"] = bson.M{"$in": []bson.ObjectID{}}
		}
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

	// valores padrão
	limit := 100
	page := 1

	if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
		if v <= 100 {
			limit = v
		}
	}
	if v, err := strconv.Atoi(pageStr); err == nil && v > 0 {
		page = v
	}

	skip := (page - 1) * limit

	// conexão Mongo
	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	col := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)

	filter, bad := buildFilterFromQueryParams(r)
	if bad {
		utils.SendResponse(w, http.StatusBadRequest,
			"Parâmetro 'status' inválido. Use 'closed' ou 'opened'", nil, 0)
		return
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := col.Find(ctx, filter, findOptions)
	if err != nil {
		// log do erro para diagnóstico
		log.Printf("[GetAllChats] erro ao executar Find. filter=%v err=%v", filter, err)
		utils.SendResponse(w, http.StatusInternalServerError,
			"Erro ao buscar chats: "+err.Error(), nil, utils.CANNOT_FIND_LEADS_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	var chats []schemas.SpaceDeskChat
	if err := cursor.All(ctx, &chats); err != nil {
		log.Printf("[GetAllChats] erro no cursor.All: %v", err)
		utils.SendResponse(w, http.StatusInternalServerError,
			"Erro ao ler chats: "+err.Error(), nil, utils.CANNOT_FIND_LEADS_IN_MONGODB)
		return
	}

	// filtro por numbers[] se necessário
	if len(numbers) > 0 {
		nMap := make(map[string]bool, len(numbers))
		for _, n := range numbers {
			nMap[n] = true
		}
		var out []schemas.SpaceDeskChat
		for _, c := range chats {
			if nMap[onlyDigits(c.ClientPhoneNumber)] {
				out = append(out, c)
			}
		}
		chats = out
	}

	utils.SendResponse(w, http.StatusOK, "Chats encontrados com sucesso", chats, 0)
}
