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

	// period := [2]string{from, until}
	responseData := map[string]any{}

	// var err error

	// handleErr := func(e error) bool {
	// 	if e != nil {
	// 		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_FIND_IN_MONGODB)
	// 		return true
	// 	}
	// 	return false
	// }

	// Relatórios de orçamentos aprovados
	if _, ok := params["budgets_total"]; ok {
		// TODO: Implementar função que retorna total de orçamentos aprovados no período
		var v int64
		// v, err = GetBudgetsTotal(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_total"] = v
	}

	if _, ok := params["budgets_total_sales_value"]; ok {
		// TODO: Implementar função que retorna valor total de vendas de orçamentos aprovados no período
		var v float64
		// v, err = GetBudgetsTotalSalesValue(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_total_sales_value"] = v
	}

	if _, ok := params["budgets_average_ticket"]; ok {
		// TODO: Implementar função que retorna ticket médio de orçamentos aprovados no período
		var v float64
		// v, err = GetBudgetsAverageTicket(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_average_ticket"] = v
	}

	if _, ok := params["budgets_converted_sales"]; ok {
		// TODO: Implementar função que retorna orçamentos aprovados convertidos em vendas
		var v int64
		// v, err = GetBudgetsConvertedSales(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_converted_sales"] = v
	}

	if _, ok := params["budgets_daily_sales_history"]; ok {
		// TODO: Implementar função que retorna histórico diário de vendas de orçamentos aprovados (dados granulares)
		var v map[string]float64
		// v, err = GetBudgetsDailySalesHistory(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_daily_sales_history"] = v
	}

	if _, ok := params["budgets_daily_count"]; ok {
		// TODO: Implementar função que retorna contagem diária de orçamentos aprovados
		var v map[string]int64
		// v, err = GetBudgetsDailyCount(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_daily_count"] = v
	}

	if _, ok := params["budgets_sales_value_by_segment"]; ok {
		// TODO: Implementar função que retorna valor de vendas por segmento (orçamentos aprovados)
		var v map[string]float64
		// v, err = GetBudgetsSalesValueBySegment(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_sales_value_by_segment"] = v
	}

	if _, ok := params["budgets_conversion_rate"]; ok {
		// TODO: Implementar função que retorna taxa de conversão de orçamentos aprovados
		var v float64
		// v, err = GetBudgetsConversionRate(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_conversion_rate"] = v
	}

	if _, ok := params["budgets_by_payment_method"]; ok {
		// TODO: Implementar função que retorna orçamentos aprovados agrupados por método de pagamento
		var v map[string]int64
		// v, err = GetBudgetsByPaymentMethod(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_by_payment_method"] = v
	}

	// Relatórios de orçamentos NÃO aprovados
	if _, ok := params["budgets_not_approved_total"]; ok {
		// TODO: Implementar função que retorna total de orçamentos NÃO aprovados no período
		var v int64
		// v, err = GetBudgetsNotApprovedTotal(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_not_approved_total"] = v
	}

	if _, ok := params["budgets_not_approved_total_value"]; ok {
		// TODO: Implementar função que retorna valor total de orçamentos NÃO aprovados no período
		var v float64
		// v, err = GetBudgetsNotApprovedTotalValue(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_not_approved_total_value"] = v
	}

	if _, ok := params["budgets_not_approved_daily_count"]; ok {
		// TODO: Implementar função que retorna contagem diária de orçamentos NÃO aprovados
		var v map[string]int64
		// v, err = GetBudgetsNotApprovedDailyCount(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_not_approved_daily_count"] = v
	}

	if _, ok := params["budgets_not_approved_by_payment_method"]; ok {
		// TODO: Implementar função que retorna orçamentos NÃO aprovados agrupados por método de pagamento
		var v map[string]int64
		// v, err = GetBudgetsNotApprovedByPaymentMethod(period[0], period[1])
		// if handleErr(err) {
		// 	return
		// }
		responseData["budgets_not_approved_by_payment_method"] = v
	}

	utils.SendResponse(w, http.StatusOK, "Relatórios de orçamentos obtidos com sucesso", responseData, 0)
}
