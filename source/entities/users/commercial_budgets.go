package users

import (
	"api/database"
	"api/middlewares"
	"api/schemas"
	"api/utils"
	"context"
	"math"
	"net/http"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetCommercialBudgets(w http.ResponseWriter, r *http.Request) {
	ctxUserRaw := r.Context().Value(middlewares.UserContextKey)
	if ctxUserRaw == nil {
		utils.SendResponse(w, http.StatusUnauthorized, "Usuário não autenticado", nil, 0)
		return
	}
	laravelUser, ok := ctxUserRaw.(middlewares.LaravelUser)
	if !ok {
		utils.SendResponse(w, http.StatusUnauthorized, "Usuário inválido", nil, 0)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	var userDoc schemas.User
	if err := client.Database(database.GetDB()).Collection(database.COLLECTION_USERS).
		FindOne(ctx, bson.M{"old_id": laravelUser.ID}).Decode(&userDoc); err != nil {
		utils.SendResponse(w, http.StatusUnauthorized, "Usuário não encontrado", nil, utils.NOT_FOUND)
		return
	}

	isCommercial := false
	for _, role := range userDoc.Role {
		if role == schemas.USERS_ROLE_COMMERCIAL {
			isCommercial = true
			break
		}
	}
	if !isCommercial {
		utils.SendResponse(w, http.StatusForbidden, "Usuário não possui permissão comercial", nil, 0)
		return
	}

	params := r.URL.Query()
	pageStr := params.Get("page")
	pageSizeStr := params.Get("pageSize")
	oldIDStr := params.Get("old_id")
	var page int64 = 1
	var pageSize int64 = 25
	if p, err := strconv.ParseInt(pageStr, 10, 64); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.ParseInt(pageSizeStr, 10, 64); err == nil && ps > 0 {
		pageSize = ps
		if pageSize > 100 {
			pageSize = 100
		}
	}
	skip := (page - 1) * pageSize

	coll := client.Database(database.GetDB()).Collection(database.COLLECTION_BUDGETS)

	filter := bson.D{{Key: "seller", Value: userDoc.ID}}
	if oldIDStr != "" {
		if oldID, err := strconv.ParseInt(oldIDStr, 10, 64); err == nil {
			filter = append(filter, bson.E{Key: "old_id", Value: oldID})
		} else {
			utils.SendResponse(w, http.StatusBadRequest, "Parâmetro old_id inválido", nil, 0)
			return
		}
	}

	totalItems, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_FIND_IN_MONGODB)
		return
	}
	totalPages := int64(math.Ceil(float64(totalItems) / float64(pageSize)))

	findOpts := options.Find().SetSkip(skip).SetLimit(pageSize).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := coll.Find(ctx, filter, findOpts)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_FIND_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	var budgets []bson.M
	if err := cursor.All(ctx, &budgets); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_FIND_IN_MONGODB)
		return
	}

	response := map[string]any{
		"items": budgets,
		"pagination": map[string]any{
			"page":        page,
			"page_size":   pageSize,
			"total_items": totalItems,
			"total_pages": totalPages,
		},
	}

	utils.SendResponse(w, http.StatusOK, "", response, 0)
}
