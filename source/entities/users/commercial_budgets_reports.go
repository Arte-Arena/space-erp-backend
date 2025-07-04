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

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetCommercialBudgetsReport(w http.ResponseWriter, r *http.Request) {
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

	var totalBudgets int64
	var approvedBudgets int64
	totalComputed := false
	approvedComputed := false

	if _, ok := params["budgets_total"]; ok {
		var err error
		totalBudgets, err = report.GetCommercialBudgetsTotal(userDoc.ID, period[0], period[1])
		if handleErr(err) {
			return
		}
		responseData["budgets_total"] = totalBudgets
		totalComputed = true
	}

	if _, ok := params["budgets_total_sales_value"]; ok {
		v, err := report.GetCommercialBudgetsTotalSalesValue(userDoc.ID, period[0], period[1])
		if handleErr(err) {
			return
		}
		responseData["budgets_total_sales_value"] = v
	}

	if _, ok := params["budgets_approved"]; ok {
		var err error
		approvedBudgets, err = report.GetCommercialBudgetsApproved(userDoc.ID, period[0], period[1])
		if handleErr(err) {
			return
		}
		responseData["budgets_approved"] = approvedBudgets
		approvedComputed = true
	}

	if _, ok := params["budgets_monthly_performance"]; ok {
		v, err := report.GetCommercialBudgetsMonthlyPerformance(userDoc.ID, period[0], period[1])
		if handleErr(err) {
			return
		}
		responseData["budgets_monthly_performance"] = v
	}

	if _, ok := params["budgets_status_percentages"]; ok {
		v, err := report.GetCommercialBudgetsStatusPercentages(userDoc.ID, period[0], period[1])
		if handleErr(err) {
			return
		}
		responseData["budgets_status_percentages"] = v
	}

	if _, ok := params["budgets_conversion_rate"]; ok {
		var err error
		if !totalComputed {
			totalBudgets, err = report.GetCommercialBudgetsTotal(userDoc.ID, period[0], period[1])
			if handleErr(err) {
				return
			}
		}
		if !approvedComputed {
			approvedBudgets, err = report.GetCommercialBudgetsApproved(userDoc.ID, period[0], period[1])
			if handleErr(err) {
				return
			}
		}
		var rate float64
		if totalBudgets > 0 {
			rate = (float64(approvedBudgets) / float64(totalBudgets)) * 100.0
		} else {
			rate = 0
		}
		responseData["budgets_conversion_rate"] = rate
	}

	if len(responseData) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "Nenhum relatório selecionado", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", responseData, 0)
}
