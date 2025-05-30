package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"net/http"
)

func GetAll(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	typeParam := r.URL.Query().Get("type")
	if typeParam == "messages" {
		getMessages(w, r, ctx)
		return
	}

	utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
}

func getMessages(w http.ResponseWriter, r *http.Request, ctx context.Context) {

}
