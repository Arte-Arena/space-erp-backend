package spacedesk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"api/database"
	"api/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// --- Modelos para o request ---

type PixDynamicCode struct {
	Code         string `json:"code"`
	MerchantName string `json:"merchant_name"`
	Key          string `json:"key"`
	KeyType      string `json:"key_type"`
}

type PaymentSetting struct {
	Type           string         `json:"type"`
	PixDynamicCode PixDynamicCode `json:"pix_dynamic_code"`
}

type Amount struct {
	Value  int `json:"value"`
	Offset int `json:"offset"`
}

type OrderIten struct {
	RetailerID string `json:"retailer_id"`
	Name       string `json:"name"`
	Amount     Amount `json:"amount"`
	Quantity   int    `json:"quantity"`
}

type OrderDetailsParameters struct {
	ReferenceID     string           `json:"reference_id"`
	Type            string           `json:"type"`
	PaymentType     string           `json:"payment_type"`
	PaymentSettings []PaymentSetting `json:"payment_settings"`
	Currency        string           `json:"currency"`
	TotalAmount     Amount           `json:"total_amount"`
	Order           struct {
		Status string `json:"status"`
		Tax    struct {
			Value       int    `json:"value"`
			Offset      int    `json:"offset"`
			Description string `json:"description"`
		} `json:"tax"`
		Items    []OrderIten `json:"items"`
		Subtotal Amount      `json:"subtotal"`
	} `json:"order"`
}

type InteractiveOrderDetails struct {
	Type string `json:"type"`
	Body struct {
		Text string `json:"text"`
	} `json:"body"`
	Action struct {
		Name       string                 `json:"name"`
		Parameters OrderDetailsParameters `json:"parameters"`
	} `json:"action"`
}

type CreatePixMessageRequest struct {
	To          string                  `json:"to"`
	UserId      string                  `json:"userId"`
	Interactive InteractiveOrderDetails `json:"interactive"`
}

func CreatePixMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// 1) Decodifica JSON de entrada
	var req CreatePixMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "JSON inválido: "+err.Error(), nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	if req.To == "" || req.Interactive.Body.Text == "" || req.Interactive.Action.Name == "" {
		utils.SendResponse(w, http.StatusBadRequest, "Campos obrigatórios ausentes", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	// 2) Conexão com MongoDB e busca do número do cliente
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
		CompanyPhoneNumber string `bson:"company_phone_number"`
	}
	chatCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
	if err := chatCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&chatDoc); err != nil {
		utils.SendResponse(w, http.StatusNotFound, "Chat não encontrado", nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		return
	}

	// 3) Monta payload 360dialog
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                chatDoc.ClientePhoneNumber,
		"type":              "interactive",
		"interactive":       req.Interactive,
	}
	bodyBytes, _ := json.Marshal(payload)

	// 4) Envia mensagem
	req360, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://waba-v2.360dialog.io/messages",
		bytes.NewReader(bodyBytes),
	)

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
		log.Printf("Erro ao enviar mensagem para 360dialog: %v", err)
		utils.SendResponse(w, http.StatusBadGateway, "Falha ao enviar mensagem", nil, utils.ERROR_TO_SEND_MESSAGE)
		return
	}
	defer resp.Body.Close()

	// Lê o corpo e já faz log do status HTTP e do conteúdo bruto
	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Printf("Erro ao ler body da resposta 360dialog: %v", readErr)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao ler resposta externa", nil, utils.ERROR_TO_READ_MESSAGE)
		return
	}
	log.Printf("360dialog HTTP %d — body: %s", resp.StatusCode, string(respBody))

	// Tenta desserializar e, se falhar, loga o erro e o JSON malformado
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

	// --- 6) Grava na coleção de eventos WhatsApp ---
	eventRaw := bson.M{
		"entry": []any{
			bson.M{
				"changes": []any{
					bson.M{
						"field": "messages",
						"value": bson.M{
							"messages": []any{
								bson.M{
									"type":          "order_details",
									"from":          "space-erp-backend",
									"to":            req.To,
									"id":            wamid,
									"timestamp":     timestamp,
									"order_details": req.Interactive.Action.Parameters,
									"user":          req.UserId,
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

	// --- 7) Grava na coleção de mensagens internas ---
	msgRaw := bson.M{
		"body":              req.Interactive.Body.Text,
		"chat_id":           objID,
		"pix":               req.Interactive,
		"by":                req.UserId,
		"from":              "company",
		"created_at":        now,
		"message_id":        wamid,
		"message_timestamp": timestamp,
		"type":              "pix",
		"status":            "",
		"updated_at":        now.Format(time.RFC3339),
	}
	msgCol := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_MESSAGE)
	if _, err := msgCol.InsertOne(ctx, msgRaw); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao inserir mensagem interna: "+err.Error(), nil, utils.ERROR_TO_INSERT_IN_MONGODB)
		return
	}

	// --- 8) Atualiza último estado no chat ---
	update := bson.M{
		"$set": bson.M{
			"last_message_id":        wamid,
			"last_message_excerpt":   req.Interactive.Body.Text,
			"last_message_sender":    "company",
			"last_message_type":      "pix",
			"last_message_timestamp": timestamp,
			"updated_at":             now.Format(time.RFC3339),
		},
	}
	if _, err := chatCol.UpdateOne(ctx, bson.M{"_id": objID}, update); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao atualizar chat: "+err.Error(), nil, utils.ERROR_TO_INSERT_IN_MONGODB)
		return
	}

	// --- 9) Broadcast via WebSocket ---
	respMap["from"] = "company"
	respMap["to"] = req.To
	respMap["type"] = "pix"
	respMap["pix"] = req.Interactive
	broadcastSpaceDeskMessage(respMap)

	// 10) Retorna resposta ao cliente HTTP
	utils.SendResponse(w, http.StatusCreated, "", respMap, 0)
}

// curl -X POST http://localhost:8080/v1/space-desk/pix-message \
// -H "Content-Type: application/json" \
// -H "Authorization: Bearer 71|Su7QAphr2E2sYDQ8TrY8K5xaBIR1AQrT6fL0W4iH073ef2e8" \
// -d '{
//   "to": "6868135cc561ee5c7bbcae79",
//   "userId": "5",
//   "interactive": {
//     "type": "order_details",
//     "body": {
//       "text": "Olá Gustavo, finalize seu pedido via PIX 😊"
//     },
//     "action": {
//       "name": "review_and_pay",
//       "parameters": {
//         "reference_id": "order_98765",
//         "type": "digital-goods",
//         "payment_type": "br",
//         "payment_settings": [
//           {
//             "type": "pix_dynamic_code",
//             "pix_dynamic_code": {
//               "code": "51107734835",
//               "merchant_name": "Arte Arena",
//             }
//           }
//         ],
//         "currency": "BRL",
//         "total_amount":   { "value": 2600, "offset": 100 },
//         "order": {
//           "status": "pending",
//           "tax": {
//             "value": 100,   "offset": 100,
//             "description": "Imposto (10%)"
//           },
//           "items": [
//             {
//               "retailer_id": "item01",
//               "name": "Camisa Premium",
//               "amount": { "value": 1500, "offset": 100 },
//               "quantity": 1
//             },
//             {
//               "retailer_id": "item02",
//               "name": "Caneca Personalizada",
//               "amount": { "value": 1000, "offset": 100 },
//               "quantity": 1
//             }
//           ],
//           "subtotal": { "value": 2500, "offset": 100 },
//           "shipping": { "value": 0, "offset": 100 },
//           "discount": { "value": 0, "offset": 100 }
//         }
//       }
//     }
//   }
// }'
