package report

import (
	"api/schemas"
	"api/utils"
	"net/http"
)

func GetByQuery(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	reportType := params.Get("type")
	reportTypeCheck := (reportType != schemas.REPORT_TYPE_CLIENTS) && (reportType != schemas.REPORT_TYPE_BUDGETS)
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
		params.Get("clients_total")
		params.Get("clients_monthly_average")
		params.Get("clients_monthly_average")
		params.Get("clients_conversion_less_thirty_days")
		params.Get("clients_time_to_close_purchase")
		params.Get("clients_new_per_month")
	}

	if reportType == schemas.REPORT_TYPE_BUDGETS {
		params.Get("budgets_converted_sales")
		params.Get("budgets_total_sales_value")
		params.Get("budgets_average_ticket")
		params.Get("budgets_monthly_sales_history")
		params.Get("budgets_sales_value_by_segment")
	}
}
