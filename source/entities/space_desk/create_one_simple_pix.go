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

type CopyCodeButton struct {
	Title   string `json:"title"`
	Payload string `json:"payload"`
}

type ButtonAction struct {
	Type     string         `json:"type"`
	CopyCode CopyCodeButton `json:"copy_code"`
}

type InteractiveButton struct {
	Type string `json:"type"`
	Body struct {
		Text string `json:"text"`
	} `json:"body"`
	Action struct {
		Buttons []ButtonAction `json:"buttons"`
	} `json:"action"`
}

type SimplePixMessageRequest struct {
	To          string            `json:"to"`
	UserId      string            `json:"userId"`
	Interactive InteractiveButton `json:"interactive"`
}

func CreateSimplePixMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// 1) Decodifica JSON de entrada
	var req SimplePixMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "JSON inválido: "+err.Error(), nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	// Validação mínima
	if req.To == "" || req.Interactive.Body.Text == "" || len(req.Interactive.Action.Buttons) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "Campos obrigatórios ausentes", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	// 2) Conexão com MongoDB (mesmo código que na versão completa)
	ctx, cancel := context.WithTimeout(r.Context(), database.MONGO_TIMEOUT)
	defer cancel()
	client, err := mongo.Connect(options.Client().ApplyURI(os.Getenv(utils.MONGODB_URI)))
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao conectar ao MongoDB: "+err.Error(), nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	objID, err := bson.ObjectIDFromHex(req.To)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "ID de chat inválido", nil, utils.INVALID_CHAT_ID_FORMAT)
		return
	}

	var chatDoc struct {
		ClientePhoneNumber string `bson:"cliente_phone_number"`
	}
	chatCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
	if err := chatCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&chatDoc); err != nil {
		utils.SendResponse(w, http.StatusNotFound, "Chat não encontrado", nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		return
	}

	interactivePayload := map[string]any{
		"type": "button",
		"body": map[string]any{
			"text": req.Interactive.Body.Text,
		},
		"action": map[string]any{
			"buttons": []map[string]any{
				{
					"type": "copy_code",
					"copy_code": map[string]any{
						"title":   req.Interactive.Action.Buttons[0].CopyCode.Title,
						"payload": req.Interactive.Action.Buttons[0].CopyCode.Payload,
					},
				},
			},
		},
	}

	// 3) Monta payload 360dialog simplificado
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                chatDoc.ClientePhoneNumber,
		"type":              "interactive",
		"interactive":       interactivePayload,
	}
	bodyBytes, _ := json.Marshal(payload)

	apiKey := os.Getenv(utils.SPACE_DESK_API_KEY)
	req360, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://waba-v2.360dialog.io/messages",
		bytes.NewReader(bodyBytes),
	)
	req360.Header.Set("Content-Type", "application/json")
	req360.Header.Set("Accept", "application/json")
	req360.Header.Set("D360-API-KEY", apiKey)

	resp, err := (&http.Client{}).Do(req360)
	if err != nil {
		log.Printf("Erro ao enviar mensagem para 360dialog: %v", err)
		utils.SendResponse(w, http.StatusBadGateway, "Falha ao enviar mensagem", nil, utils.ERROR_TO_SEND_MESSAGE)
		return
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("Erro ao ler body da resposta 360dialog: %v", readErr)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao ler resposta externa", nil, utils.ERROR_TO_READ_MESSAGE)
		return
	}
	log.Printf("360dialog HTTP %d — body: %s", resp.StatusCode, string(respBody))

	var respMap map[string]any
	if err := json.Unmarshal(respBody, &respMap); err != nil {
		log.Printf("Erro ao fazer Unmarshal da resposta 360dialog: %v\nConteúdo recebido: %s", err, string(respBody))
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao ler resposta externa", nil, utils.ERROR_TO_READ_MESSAGE)
		return
	}

	// 5) Extrai o ID retornado (wamid)
	wamid := extractWamid(respMap)
	if wamid == "not_returned" {
		utils.SendResponse(w, http.StatusBadGateway, "API não retornou wamid", nil, utils.ERROR_TO_SEND_MESSAGE)
		return
	}

	now := time.Now().UTC()
	timestamp := fmt.Sprint(now.Unix())

	// --- 6) Grava na coleção de eventos WhatsApp (adaptado) ---
	eventRaw := bson.M{
		"entry": []any{
			bson.M{
				"changes": []any{
					bson.M{
						"field": "messages",
						"value": bson.M{
							"messages": []any{
								bson.M{
									"type":      "button",
									"from":      "space-erp-backend",
									"to":        req.To,
									"id":        wamid,
									"timestamp": timestamp,
									"button":    req.Interactive.Action.Buttons[0].CopyCode,
									"user":      req.UserId,
								},
							},
						},
					},
				},
			},
		},
	}
	eventsCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_EVENTS_WHATSAPP)
	if _, err := eventsCol.InsertOne(ctx, eventRaw); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao inserir evento WhatsApp: "+err.Error(), nil, utils.ERROR_TO_INSERT_IN_MONGODB)
		return
	}

	msgRaw := bson.M{
		"body":              req.Interactive.Body.Text,
		"chat_id":           objID,
		"by":                req.UserId,
		"from":              "company",
		"created_at":        now,
		"message_id":        wamid,
		"message_timestamp": timestamp,
		"type":              "pix_simple",
		"status":            "",
		"updated_at":        now.Format(time.RFC3339),
	}
	msgCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_MESSAGE)
	if _, err := msgCol.InsertOne(ctx, msgRaw); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao inserir mensagem interna: "+err.Error(), nil, utils.ERROR_TO_INSERT_IN_MONGODB)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"last_message_id":        wamid,
			"last_message_excerpt":   req.Interactive.Body.Text,
			"last_message_sender":    "company",
			"last_message_timestamp": timestamp,
			"updated_at":             now.Format(time.RFC3339),
		},
	}
	if _, err := chatCol.UpdateOne(ctx, bson.M{"_id": objID}, update); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao atualizar chat: "+err.Error(), nil, utils.ERROR_TO_INSERT_IN_MONGODB)
		return
	}

	respMap["from"] = "company"
	respMap["to"] = req.To
	respMap["type"] = "pix_simple"
	broadcastSpaceDeskMessage(respMap)

	utils.SendResponse(w, http.StatusCreated, "", respMap, 0)
}

func RegisterPixRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/v1/space-desk/pix-message", CreatePixMessage)              // Rota existente
	mux.HandleFunc("/v1/space-desk/simple-pix-message", CreateSimplePixMessage) // Nova rota
}
