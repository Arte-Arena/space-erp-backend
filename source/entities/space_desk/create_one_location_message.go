package spacedesk

import (
	"api/database"
	"api/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Struct para request local
type CreateLocationRequest struct {
	To     string `json:"to"`
	UserId string `json:"userId"`
	Body   string `json:"body"`
}

func CreateLocationRequestMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req CreateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "JSON inválido: "+err.Error(), nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	if req.To == "" || req.Body == "" {
		utils.SendResponse(w, http.StatusBadRequest, "Campos obrigatórios ausentes", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), database.MONGO_TIMEOUT)
	defer cancel()
	client, err := mongo.Connect(options.Client().ApplyURI(os.Getenv(utils.MONGODB_URI)))
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao conectar ao MongoDB: "+err.Error(), nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	// Busca telefone do destinatário no chat metadata
	var chatDoc struct {
		ClientePhoneNumber string `bson:"cliente_phone_number"`
		CompanyPhoneNumber string `bson:"company_phone_number"`
	}
	col := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
	objID, err := bson.ObjectIDFromHex(req.To)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "ID do chat inválido", nil, utils.INVALID_CHAT_ID_FORMAT)
		return
	}
	if err = col.FindOne(ctx, bson.M{"_id": objID}).Decode(&chatDoc); err != nil {
		utils.SendResponse(w, http.StatusNotFound, "Chat não encontrado", nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		return
	}
	recipient := chatDoc.ClientePhoneNumber

	// Monta o payload para a 360dialog
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                recipient,
		"type":              "interactive",
		"interactive": map[string]any{
			"type": "location_request_message",
			"body": map[string]string{
				"text": req.Body,
			},
			"action": map[string]string{
				"name": "send_location",
			},
		},
	}

	body, _ := json.Marshal(payload)
	req360, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://waba-v2.360dialog.io/messages",
		bytes.NewReader(body),
	)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao criar requisição externa: "+err.Error(), nil, utils.ERROR_TO_CREATE_EXTERNAL_CONNECTION)
		return
	}

	var apiKey string
	switch chatDoc.CompanyPhoneNumber {
	case "5511958339942":
		apiKey = os.Getenv(utils.SPACE_DESK_API_KEY)
	case "551123371548":
		apiKey = os.Getenv(utils.SPACE_DESK_API_KEY_2)
	}

	req360.Header.Set("Content-Type", "application/json")
	req360.Header.Set("Accept", "application/json")
	req360.Header.Set("D360-API-KEY", apiKey)

	resp, err := (&http.Client{}).Do(req360)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "Falha ao enviar mensagem", nil, utils.ERROR_TO_SEND_MESSAGE)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Resposta da 360dialog: %s", respBody)

	var respMap map[string]any
	if err := json.Unmarshal(respBody, &respMap); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao ler resposta externa", nil, utils.ERROR_TO_READ_MESSAGE)
		return
	}
	wamid := extractWamid(respMap)
	if wamid == "not_returned" {
		utils.SendResponse(w, http.StatusBadGateway, "API não retornou wamid, mensagem não salva", nil, utils.ERROR_TO_SEND_MESSAGE)
		return
	}

	now := time.Now().UTC()
	event := bson.M{
		"entry": []any{bson.M{
			"changes": []any{bson.M{
				"field": "messages", "value": bson.M{
					"messages": []any{bson.M{
						"type":      "location_request_message",
						"from":      "space-erp-backend",
						"to":        req.To,
						"id":        wamid,
						"timestamp": fmt.Sprint(now.Unix()),
						"body":      req.Body,
						"user":      req.UserId,
					}}}}}}},
	}
	colEvents := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_EVENTS_WHATSAPP)
	if _, err := colEvents.InsertOne(ctx, event); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao inserir evento: "+err.Error(), nil, utils.ERROR_TO_INSERT_IN_MONGODB)
		return
	}

	broadcastSpaceDeskMessage(respMap)
	utils.SendResponse(w, http.StatusCreated, "", respMap, 0)
}
