package report

import (
	"api/utils"
	"net/http"
)

func GetByQueryBudgetsV2(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	from := params.Get("from")
	until := params.Get("until")

	if from == "" || until == "" {
		utils.SendResponse(w, http.StatusBadRequest, "Os parâmetros 'from' e 'until' são obrigatórios", nil, 0)
		return
	}

	fromValid := utils.IsValidDate(from)
	untilValid := utils.IsValidDate(until)

	if !fromValid || !untilValid {
		utils.SendResponse(w, http.StatusBadRequest, "Formato de data inválido. Use YYYY-MM-DD", nil, 0)
		return
	}

	responseData := map[string]any{}

	var err error

	handleErr := func(e error) bool {
		if e != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_FIND_IN_MONGODB)
			return true
		}
		return false
	}

	if _, ok := params["budgets_total"]; ok {
		var v int64
		v, err = GetBudgetsTotalV2(from, until)
		if handleErr(err) {
			return
		}
		responseData["budgets_total"] = v
	}

	if _, ok := params["budgets_total_sales_value"]; ok {
		var v float64
		v, err = GetBudgetsTotalSalesValueV2(from, until)
		if handleErr(err) {
			return
		}
		responseData["budgets_total_sales_value"] = v
	}

	if _, ok := params["budgets_average_ticket"]; ok {
		var v float64
		v, err = GetBudgetsAverageTicketV2(from, until)
		if handleErr(err) {
			return
		}
		responseData["budgets_average_ticket"] = v
	}

	if _, ok := params["budgets_daily_sales_history"]; ok {
		var v map[string]float64
		v, err = GetBudgetsDailySalesHistoryV2(from, until)
		if handleErr(err) {
			return
		}
		responseData["budgets_daily_sales_history"] = v
	}

	if _, ok := params["budgets_daily_count"]; ok {
		var v map[string]int64
		v, err = GetBudgetsDailyCountV2(from, until)
		if handleErr(err) {
			return
		}
		responseData["budgets_daily_count"] = v
	}

	if _, ok := params["budgets_sales_value_by_segment"]; ok {
		var v map[string]float64
		v, err = GetBudgetsSalesValueBySegmentV2(from, until)
		if handleErr(err) {
			return
		}
		responseData["budgets_sales_value_by_segment"] = v
	}

	if _, ok := params["budgets_by_payment_method"]; ok {
		var v map[string]int64
		v, err = GetBudgetsByPaymentMethodV2(from, until)
		if handleErr(err) {
			return
		}
		responseData["budgets_by_payment_method"] = v
	}

	if _, ok := params["budgets_not_approved_total"]; ok {
		var v int64
		v, err = GetBudgetsNotApprovedTotalV2(from, until)
		if handleErr(err) {
			return
		}
		responseData["budgets_not_approved_total"] = v
	}

	if _, ok := params["budgets_not_approved_total_value"]; ok {
		var v float64
		v, err = GetBudgetsNotApprovedTotalValueV2(from, until)
		if handleErr(err) {
			return
		}
		responseData["budgets_not_approved_total_value"] = v
	}

	if _, ok := params["budgets_not_approved_daily_count"]; ok {
		var v map[string]int64
		v, err = GetBudgetsNotApprovedDailyCountV2(from, until)
		if handleErr(err) {
			return
		}
		responseData["budgets_not_approved_daily_count"] = v
	}

	if _, ok := params["budgets_not_approved_by_payment_method"]; ok {
		var v map[string]int64
		v, err = GetBudgetsNotApprovedByPaymentMethodV2(from, until)
		if handleErr(err) {
			return
		}
		responseData["budgets_not_approved_by_payment_method"] = v
	}

	utils.SendResponse(w, http.StatusOK, "Relatórios de orçamentos obtidos com sucesso", responseData, 0)
}
