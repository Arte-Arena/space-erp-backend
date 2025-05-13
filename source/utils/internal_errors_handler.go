package utils

import "fmt"

const (
	LEADS_INVALID_REQUEST_DATA = iota
	CANNOT_CONNECT_TO_MONGODB
	CANNOT_INSERT_LEAD_TO_MONGODB
)

func SendInternalError(internalErrorCode int) string {
	return fmt.Sprintf("Ocorreu um erro interno no servidor. Por favor, tente novamene mais tarde (Cod: %d)", internalErrorCode)
}
