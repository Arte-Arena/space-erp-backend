package report

import (
	"api/schemas"
	"api/utils"
	"net/http"
)

func GetByQuery(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	reportType := params.Get("type")
	reportTypeCheck := (reportType != schemas.REPORT_TYPE_CLIENTS) && (reportType != schemas.REPORT_TYPE_BUDGETS) && (reportType != schemas.REPORT_TYPE_LEADS) && (reportType != schemas.REPORT_TYPE_ORDERS)
	if reportTypeCheck {
		utils.SendResponse(w, http.StatusBadRequest, "Tipo de relatório inválido", nil, 0)
		return
	}

	period := [2]string{"", ""}
	from := params.Get("from")
	until := params.Get("until")

	fromValid := utils.IsValidDate(from)
	untilValid := utils.IsValidDate(until)

	if fromValid && untilValid {
		period = [2]string{from, until}
	} else if fromValid {
		period = [2]string{from, ""}
	} else if untilValid {
		period = [2]string{"", until}
	}

	_ = period

	responseData := map[string]any{}

	var err error

	handleErr := func(e error) bool {
		if e != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_FIND_IN_MONGODB)
			return true
		}
		return false
	}

	if reportType == schemas.REPORT_TYPE_CLIENTS {
		if _, ok := params["clients_total"]; ok {
			var v int64
			v, err = GetClientsTotal(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["clients_total"] = v
		}

		if _, ok := params["clients_monthly_average"]; ok {
			var v float64
			v, err = GetClientsMonthlyAverage(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["clients_monthly_average"] = v
		}

		if _, ok := params["clients_conversion_less_thirty_days"]; ok {
			var v int64
			v, err = GetClientsConversionLessThirtyDays(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["clients_conversion_less_thirty_days"] = v
		}

		if _, ok := params["clients_time_to_close_purchase"]; ok {
			var v float64
			v, err = GetClientsTimeToClosePurchase(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["clients_time_to_close_purchase"] = v
		}

		if _, ok := params["clients_new_per_month"]; ok {
			var v map[string]int64
			v, err = GetClientsNewPerMonth(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["clients_new_per_month"] = v
		}
	}

	if reportType == schemas.REPORT_TYPE_BUDGETS {
		if _, ok := params["budgets_converted_sales"]; ok {
			var v int64
			v, err = GetBudgetsConvertedSales(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["budgets_converted_sales"] = v
		}

		if _, ok := params["budgets_total_sales_value"]; ok {
			var v float64
			v, err = GetBudgetsTotalSalesValue(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["budgets_total_sales_value"] = v
		}

		if _, ok := params["budgets_average_ticket"]; ok {
			var v float64
			v, err = GetBudgetsAverageTicket(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["budgets_average_ticket"] = v
		}

		if _, ok := params["budgets_monthly_sales_history"]; ok {
			var v map[string]float64
			v, err = GetBudgetsMonthlySalesHistory(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["budgets_monthly_sales_history"] = v
		}

		if _, ok := params["budgets_sales_value_by_segment"]; ok {
			var v map[string]float64
			v, err = GetBudgetsSalesValueBySegment(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["budgets_sales_value_by_segment"] = v
		}
	}

	if reportType == schemas.REPORT_TYPE_LEADS {
		if _, ok := params["leads_total"]; ok {
			var v int64
			v, err = GetLeadsTotal(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["leads_total"] = v
		}

		if _, ok := params["leads_monthly_average"]; ok {
			var v float64
			v, err = GetLeadsMonthlyAverage(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["leads_monthly_average"] = v
		}

		if _, ok := params["leads_conversion_less_thirty_days"]; ok {
			var v int64
			v, err = GetLeadsConversionLessThirtyDays(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["leads_conversion_less_thirty_days"] = v
		}

		if _, ok := params["leads_time_to_close_purchase"]; ok {
			var v float64
			v, err = GetLeadsTimeToClosePurchase(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["leads_time_to_close_purchase"] = v
		}

		if _, ok := params["leads_new_per_month"]; ok {
			var v map[string]int64
			v, err = GetLeadsNewPerMonth(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["leads_new_per_month"] = v
		}
	}

	if reportType == schemas.REPORT_TYPE_ORDERS {
		if _, ok := params["orders_total"]; ok {
			var v int64
			v, err = GetOrdersTotal(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["orders_total"] = v
		}

		if _, ok := params["orders_monthly_average"]; ok {
			var v float64
			v, err = GetOrdersMonthlyAverage(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["orders_monthly_average"] = v
		}

		if _, ok := params["orders_sales_value_by_status"]; ok {
			var v map[string]float64
			v, err = GetOrdersSalesValueByStatus(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["orders_sales_value_by_status"] = v
		}

		if _, ok := params["orders_sales_value_by_type"]; ok {
			var v map[string]float64
			v, err = GetOrdersSalesValueByType(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["orders_sales_value_by_type"] = v
		}

		if _, ok := params["orders_total_sales_value"]; ok {
			var v float64
			v, err = GetOrdersTotalSalesValue(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["orders_total_sales_value"] = v
		}

		if _, ok := params["orders_monthly_sales_history"]; ok {
			var v map[string]float64
			v, err = GetOrdersMonthlySalesHistory(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["orders_monthly_sales_history"] = v
		}

		if _, ok := params["orders_daily_sales"]; ok {
			var v map[string]int64
			v, err = GetOrdersDailySales(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["orders_daily_sales"] = v
		}

		if _, ok := params["orders_daily_sales_value"]; ok {
			var v map[string]float64
			v, err = GetOrdersDailySalesValue(period[0], period[1])
			if handleErr(err) {
				return
			}
			responseData["orders_daily_sales_value"] = v
		}
	}

	if len(responseData) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "Nenhum relatório selecionado", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", responseData, 0)
}
