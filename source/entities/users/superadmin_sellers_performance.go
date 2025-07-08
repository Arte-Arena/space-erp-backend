package users

import (
	"api/database"
	"api/entities/report"
	"api/middlewares"
	"api/schemas"
	"api/utils"
	"context"
	"net/http"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetSuperadminSellersPerformanceReport(w http.ResponseWriter, r *http.Request) {
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
	err = client.Database(database.GetDB()).Collection(database.COLLECTION_USERS).
		FindOne(ctx, bson.M{"old_id": laravelUser.ID}).Decode(&userDoc)
	if err != nil {
		utils.SendResponse(w, http.StatusUnauthorized, "Usuário não encontrado", nil, utils.NOT_FOUND)
		return
	}

	isSuperAdmin := false
	for _, role := range userDoc.Role {
		if role == schemas.USERS_ROLE_SUPER_ADMIN {
			isSuperAdmin = true
			break
		}
	}
	if !isSuperAdmin {
		utils.SendResponse(w, http.StatusForbidden, "Usuário não possui permissão super_admin", nil, 0)
		return
	}

	params := r.URL.Query()
	from := params.Get("from")
	until := params.Get("until")
	sellerIDsParam := params.Get("seller_ids")

	var sellerIDs []bson.ObjectID
	if sellerIDsParam != "" {
		for _, idStr := range strings.Split(sellerIDsParam, ",") {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}
			id, err := bson.ObjectIDFromHex(idStr)
			if err == nil {
				sellerIDs = append(sellerIDs, id)
			}
		}
	}

	responseData := map[string]any{}
	handleErr := func(e error) bool {
		if e != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_FIND_IN_MONGODB)
			return true
		}
		return false
	}

	if _, ok := params["total_sales"]; ok {
		v, err := report.GetSuperadminSellersTotalSales(client, sellerIDs, from, until)
		if handleErr(err) {
			return
		}
		responseData["total_sales"] = v
	}
	if _, ok := params["sales_count"]; ok {
		v, err := report.GetSuperadminSellersSalesCount(client, sellerIDs, from, until)
		if handleErr(err) {
			return
		}
		responseData["sales_count"] = v
	}
	if _, ok := params["conversion_rate"]; ok {
		v, err := report.GetSuperadminSellersConversionRate(client, sellerIDs, from, until)
		if handleErr(err) {
			return
		}
		responseData["conversion_rate"] = v
	}
	if _, ok := params["monthly_sales"]; ok {
		v, err := report.GetSuperadminSellersMonthlySales(client, sellerIDs, from, until)
		if handleErr(err) {
			return
		}
		responseData["monthly_sales"] = v
	}
	if _, ok := params["monthly_conversion"]; ok {
		v, err := report.GetSuperadminSellersMonthlyConversion(client, sellerIDs, from, until)
		if handleErr(err) {
			return
		}
		responseData["monthly_conversion"] = v
	}
	if _, ok := params["ranking"]; ok {
		v, err := report.GetSuperadminSellersRanking(client, from, until)
		if handleErr(err) {
			return
		}
		responseData["ranking"] = v
	}

	if len(responseData) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "Nenhum relatório selecionado", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", responseData, 0)
}
