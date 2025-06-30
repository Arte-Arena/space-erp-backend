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

	if reportType == schemas.REPORT_TYPE_CLIENTS {
		clientsTotal := params.Get("clients_total")
		clientsMonthlyAverage := params.Get("clients_monthly_average")
		clientsConversionLessThirtyDays := params.Get("clients_conversion_less_thirty_days")
		clientsTimeToClosePurchase := params.Get("clients_time_to_close_purchase")
		clientsNewPerMonth := params.Get("clients_new_per_month")

		_ = clientsTotal
		_ = clientsMonthlyAverage
		_ = clientsConversionLessThirtyDays
		_ = clientsTimeToClosePurchase
		_ = clientsNewPerMonth
	}

	if reportType == schemas.REPORT_TYPE_BUDGETS {
		budgetsConvertedSales := params.Get("budgets_converted_sales")
		budgetsTotalSalesValue := params.Get("budgets_total_sales_value")
		budgetsAverageTicket := params.Get("budgets_average_ticket")
		budgetsMonthlySalesHistory := params.Get("budgets_monthly_sales_history")
		budgetsSalesValueBySegment := params.Get("budgets_sales_value_by_segment")

		_ = budgetsConvertedSales
		_ = budgetsTotalSalesValue
		_ = budgetsAverageTicket
		_ = budgetsMonthlySalesHistory
		_ = budgetsSalesValueBySegment
	}

	if reportType == schemas.REPORT_TYPE_LEADS {
		leadsTotal := params.Get("leads_total")
		leadsMonthlyAverage := params.Get("leads_monthly_average")
		leadsConversionLessThirtyDays := params.Get("leads_conversion_less_thirty_days")
		leadsTimeToClosePurchase := params.Get("leads_time_to_close_purchase")
		leadsNewPerMonth := params.Get("leads_new_per_month")

		_ = leadsTotal
		_ = leadsMonthlyAverage
		_ = leadsConversionLessThirtyDays
		_ = leadsTimeToClosePurchase
		_ = leadsNewPerMonth
	}

	if reportType == schemas.REPORT_TYPE_ORDERS {
		ordersTotal := params.Get("orders_total")
		ordersMonthlyAverage := params.Get("orders_monthly_average")
		ordersSalesValueByStatus := params.Get("orders_sales_value_by_status")
		ordersSalesValueByType := params.Get("orders_sales_value_by_type")

		_ = ordersTotal
		_ = ordersMonthlyAverage
		_ = ordersSalesValueByStatus
		_ = ordersSalesValueByType
	}
}
