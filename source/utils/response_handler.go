package utils

import (
	"api/source/schemas"
	"encoding/json"
	"net/http"
)

func SendResponse(w http.ResponseWriter, statusCode int, message string, data any, internalErrorCode int) {
	if internalErrorCode != 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: SendInternalError(internalErrorCode),
		})
		return
	}

	if (message == "") && (data == nil) {
		w.WriteHeader(statusCode)
		return
	}

	if (message != "") && (data == nil) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Message: message,
		})
		return
	}

	if (message == "") && (data != nil) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(schemas.ApiResponse{
			Data: data,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(schemas.ApiResponse{
		Data:    data,
		Message: message,
	})
}
