package report

import (
	"api/utils"
	"net/http"
)

func GetByQueryV2(w http.ResponseWriter, r *http.Request) {
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

	period := [2]string{from, until}
	responseData := map[string]any{}

	var err error

	handleErr := func(e error) bool {
		if e != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_FIND_IN_MONGODB)
			return true
		}
		return false
	}

	if _, ok := params["clients_total"]; ok {
		var v int64
		v, err = GetClientsTotal(period[0], period[1])
		if handleErr(err) {
			return
		}
		responseData["clients_total"] = v
	}

	if _, ok := params["clients_new_per_day"]; ok {
		var v map[string]int64
		v, err = GetClientsNewPerDayV2(period[0], period[1])
		if handleErr(err) {
			return
		}
		responseData["clients_new_per_day"] = v
	}

	if _, ok := params["clients_by_person_type_per_day"]; ok {
		var v map[string]map[string]int64
		v, err = GetClientsByPersonTypePerDayV2(period[0], period[1])
		if handleErr(err) {
			return
		}
		responseData["clients_by_person_type_per_day"] = v
	}

	if _, ok := params["clients_by_state_per_day"]; ok {
		var v map[string]map[string]int64
		v, err = GetClientsByStatePerDayV2(period[0], period[1])
		if handleErr(err) {
			return
		}
		responseData["clients_by_state_per_day"] = v
	}

	utils.SendResponse(w, http.StatusOK, "", responseData, 0)
}
