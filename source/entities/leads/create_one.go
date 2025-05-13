package leads

import (
	"api/source/schemas"
	"api/source/utils"
	"encoding/json"
	"net/http"
)

func CreateOne(w http.ResponseWriter, r *http.Request) {
	leadSchema := &schemas.Lead{}
	if err := json.NewDecoder(r.Body).Decode(&leadSchema); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: utils.SendInternalError(utils.LEADS_INVALID_REQUEST_DATA),
		})
		return
	}

}
