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
	"slices"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetCommercialOrdersReport(w http.ResponseWriter, r *http.Request) {
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

	isCommercial := slices.Contains(userDoc.Role, schemas.USERS_ROLE_COMMERCIAL)
	if !isCommercial {
		utils.SendResponse(w, http.StatusForbidden, "Usuário não possui permissão comercial", nil, 0)
		return
	}

	params := r.URL.Query()
	from := params.Get("from")
	until := params.Get("until")

	period := [2]string{"", ""}
	if utils.IsValidDate(from) {
		period[0] = from
	}
	if utils.IsValidDate(until) {
		period[1] = until
	}

	responseData := map[string]any{}

	handleErr := func(e error) bool {
		if e != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_FIND_IN_MONGODB)
			return true
		}
		return false
	}

	if _, ok := params["orders_monthly_performance"]; ok {
		v, err := report.GetCommercialOrdersMonthlyPerformance(userDoc.ID, period[0], period[1])
		if handleErr(err) {
			return
		}
		responseData["orders_monthly_performance"] = v
	}

	if len(responseData) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "Nenhum relatório selecionado", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", responseData, 0)
}
