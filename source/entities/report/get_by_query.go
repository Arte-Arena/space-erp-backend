package report

import "net/http"

func GetByQuery(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	params.Get("type")
	params.Get("from")
	params.Get("until")

	params.Get("clients_total")
	params.Get("clients_monthly_average")
	params.Get("clients_monthly_average")
	params.Get("clients_conversion_less_thirty_days")
	params.Get("clients_time_to_close_purchase")
	params.Get("clients_new_per_month")

	params.Get("budgets_converted_sales")
	params.Get("budgets_total_sales_value")
	params.Get("budgets_average_ticket")
	params.Get("budgets_monthly_sales_history")
	params.Get("budgets_sales_value_by_segment")
}
