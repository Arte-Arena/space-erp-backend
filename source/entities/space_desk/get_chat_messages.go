package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetAllMessagesByChatId(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	chatId := r.URL.Query().Get("chat_id")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := int64(1)
	pageSize := int64(25)

	if pageStr != "" {
		if parsedPage, err := strconv.ParseInt(pageStr, 10, 64); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if pageSizeStr != "" {
		if parsedPageSize, err := strconv.ParseInt(pageSizeStr, 10, 64); err == nil && parsedPageSize > 0 {
			pageSize = parsedPageSize
			if pageSize > 100 {
				pageSize = 100
			}
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

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_MESSAGE)

	objectID, err := bson.ObjectIDFromHex(chatId)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_CHAT_ID)
		return
	}

	filter := bson.D{{Key: "chat_id", Value: objectID}}

	// Conta total e total de páginas
	totalItems, _ := collection.CountDocuments(ctx, filter)
	totalPages := int64(math.Ceil(float64(totalItems) / float64(pageSize)))

	// Paginação padrão: skip e limit
	skip := max((page-1)*pageSize, 0)

	findOptions := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}). // do mais novo para o mais antigo
		SetSkip(skip).
		SetLimit(pageSize)

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEADS_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	var allMessages []bson.M
	if err := cursor.All(ctx, &allMessages); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEADS_IN_MONGODB)
		return
	}

	sort.Slice(allMessages, func(i, j int) bool {
		id1 := allMessages[i]["_id"].(bson.ObjectID)
		id2 := allMessages[j]["_id"].(bson.ObjectID)
		return id1.Timestamp().Before(id2.Timestamp())
	})

	response := map[string]any{
		"items": allMessages,
		"pagination": map[string]any{
			"page":        page,
			"page_size":   pageSize,
			"total_items": totalItems,
			"total_pages": totalPages,
		},
	}

	utils.SendResponse(w, http.StatusOK, "", response, 0)
}
