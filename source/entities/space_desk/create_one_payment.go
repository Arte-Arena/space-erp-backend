package spacedesk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

type OrderItem struct {
	RetailerID string `json:"retailer_id"`
	Name       string `json:"name"`
	Amount     struct {
		Value  int `json:"value"`
		Offset int `json:"offset"`
	} `json:"amount"`
	Quantity int `json:"quantity"`
}

type OrderDetailsRequest struct {
	ChatID      string      `json:"chatId"`
	ReferenceID string      `json:"referenceId"`
	PixCode     string      `json:"pixCode"`
	Items       []OrderItem `json:"items"`
	TotalValue  int         `json:"totalValue"`
	UserID      string      `json:"userId"`
}

func CreateOrderDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req OrderDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("JSON inválido:", err)
		utils.SendResponse(w, http.StatusBadRequest, "JSON inválido", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	// Conectar ao MongoDB
	ctx, cancel := context.WithTimeout(r.Context(), database.MONGO_TIMEOUT)
	defer cancel()
	mongoURI := os.Getenv(utils.MONGODB_URI)
	clientOpts := options.Client().ApplyURI(mongoURI)
	dbClient, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Println("Erro ao conectar ao MongoDB:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao conectar ao MongoDB: "+err.Error(), nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer dbClient.Disconnect(ctx)

	// Buscar telefone do usuário
	colChats := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
	var chat struct {
		ClientePhoneNumber string `bson:"cliente_phone_number"`
	}
	objID, _ := bson.ObjectIDFromHex(req.ChatID)
	if err := colChats.FindOne(ctx, bson.M{"_id": objID}).Decode(&chat); err != nil {
		log.Println("Chat não encontrado:", err)
		utils.SendResponse(w, http.StatusNotFound, "Chat não encontrado", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
		return
	}

	// Montar payload conforme Interactive Order Details (PIX)
	payload := map[string]any{
    "name":           "order_details_pix_2",
    "language":       "pt_BR",
    "category":       "UTILITY",
    "display_format": "ORDER_DETAILS",
    "components": []any{
      map[string]any{
        "type":   "HEADER",
        "format": "TEXT",
        "text":   "Teste pagamento pix",
      },
      map[string]any{
        "type": "BODY",
        "text": "Obrigado pela sua compra. Segue abaixo o codigo",
      },
      map[string]any{
        "type": "BUTTONS",
        "buttons": []any{
          map[string]any{
            "type": "ORDER_DETAILS",
            "text": "Copy Pix code",
          },
        },
      },
    },
  } // :contentReference[oaicite:3]{index=3}

	// Enviar requisição à 360dialog
	bodyBytes, _ := json.Marshal(payload)
	apiKey := os.Getenv(utils.SPACE_DESK_API_KEY)
	req360, _ := http.NewRequestWithContext(ctx, "POST", "https://waba-v2.360dialog.io/messages", bytes.NewReader(bodyBytes))
	req360.Header.Set("Content-Type", "application/json")
	req360.Header.Set("D360-API-KEY", apiKey)

	clientHTTP := &http.Client{Timeout: 10 * time.Second}
	resp360, err := clientHTTP.Do(req360)
	if err != nil || resp360.StatusCode >= 300 {
		log.Println("Erro ao enviar PIX:", err, resp360.Status)
		utils.SendResponse(w, http.StatusBadGateway, "Falha ao enviar mensagem", nil, utils.ERROR_TO_SEND_MESSAGE)
		return
	}
	defer resp360.Body.Close()

	var respStruct struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(resp360.Body).Decode(&respStruct); err != nil {
		log.Println("Erro ao ler resposta externa:", err)
	} else if len(respStruct.Messages) == 0 || respStruct.Messages[0].ID == "" {
		log.Println("Nenhuma mensagem retornada pela API 360dialog")
	} else {
		messageID := respStruct.Messages[0].ID
		now := time.Now().UTC()

		// 2) Inserir evento no SPACE_DESK_EVENTS_WHATSAPP
		colEvents := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_EVENTS_WHATSAPP)
		rawEvent := bson.M{
			"entry": []any{
				bson.M{
					"changes": []any{
						bson.M{
							"field": "messages",
							"value": bson.M{
								"messages": []any{
									bson.M{
										"from":      "space-erp-backend",
										"to":        req.ChatID,
										"id":        messageID,
										"timestamp": fmt.Sprint(now.Unix()),
										"type":      "interactive",
									},
								},
							},
						},
					},
				},
			},
		}
		if _, err := colEvents.InsertOne(ctx, rawEvent); err != nil {
			log.Println("Erro ao inserir evento no MongoDB:", err)
		}

		colMessages := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_MESSAGE)
		messageDoc := bson.M{
			"body":              "Detalhes do seu pedido", // ou outro texto de resumo
			"chat_id":           objID,
			"by":                req.UserID, // ou outro identificador
			"from":              "company",
			"created_at":        now,
			"message_id":        messageID,
			"message_timestamp": fmt.Sprint(now.Unix()),
			"type":              "interactive",
			"status":            "",
			"updated_at":        now.Format(time.RFC3339),
		}
		if _, err := colMessages.InsertOne(ctx, messageDoc); err != nil {
			log.Println("Erro ao inserir mensagem no MongoDB:", err)
		}

		colChat := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
		update := bson.M{"$set": bson.M{
			"last_message_id":        messageID,
			"last_message_timestamp": fmt.Sprint(now.Unix()),
			"last_message_excerpt":   req.ReferenceID,
			"last_message_type":      "interactive",
			"last_message_sender":    "company",
			"updated_at":             now.Format(time.RFC3339),
		}}
		if _, err := colChat.UpdateOne(ctx, bson.M{"_id": objID}, update); err != nil {
			log.Println("Erro ao atualizar chat no MongoDB:", err)
		}
	}

	// Prepare a map for broadcasting, including the decoded messages and extra fields
	broadcastData := map[string]any{
		"messages": respStruct.Messages,
		"from":     "company",
		"to":       objID,
	}
	broadcastSpaceDeskMessage(broadcastData)

	utils.SendResponse(w, http.StatusCreated, "", payload, 0)
}
