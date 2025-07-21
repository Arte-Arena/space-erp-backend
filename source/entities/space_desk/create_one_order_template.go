package spacedesk

import (
	"api/database"
	"api/utils"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateOrderDetailsTemplate(r *http.Request, w http.ResponseWriter) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		To string `json:"chat_id"`
	}

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

	colChats := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
	var chatDoc struct {
		ClientePhoneNumber string `bson:"cliente_phone_number"`
		LastMessage        any    `bson:"last_message_from_client_timestamp"`
		CompanyPhoneNumber string `bson:"company_phone_number"`
	}

	objID, err := bson.ObjectIDFromHex(reqBody.To)
	if err != nil {
		log.Println("Erro ao converter ID do chat para ObjectID:", err, "ID recebido:", reqBody.To)
		utils.SendResponse(w, http.StatusBadRequest, "ID do chat inválido", nil, utils.INVALID_CHAT_ID_FORMAT)
		return
	}
	err = colChats.FindOne(r.Context(), bson.M{"_id": objID}).Decode(&chatDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.SendResponse(w, http.StatusNotFound, "Chat não encontrado", nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
			return
		}
		log.Println("Erro ao buscar chat:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao buscar chat: "+err.Error(), nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		return
	}

	var apiKey string
	switch chatDoc.CompanyPhoneNumber {
	case "5511958339942":
		apiKey = os.Getenv(utils.SPACE_DESK_API_KEY_2)
	case "551123371548":
		apiKey = os.Getenv(utils.SPACE_DESK_API_KEY)
	}

	payload := map[string]interface{}{
		"name":           "order_details_pix_2",
		"language":       "pt_BR",
		"category":       "UTILITY",
		"display_format": "ORDER_DETAILS",
		"components": []interface{}{
			map[string]interface{}{
				"type":   "HEADER",
				"format": "TEXT",
				"text":   "Teste pagamento pix",
			},
			map[string]interface{}{
				"type": "BODY",
				"text": "Obrigado pela sua compra. Segue abaixo o codigo",
			},
			map[string]interface{}{
				"type": "BUTTONS",
				"buttons": []interface{}{
					map[string]interface{}{
						"type": "ORDER_DETAILS",
						"text": "Copy Pix code",
					},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(context.Background(), "POST",
		"https://waba-v2.360dialog.io/message_templates",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("D360-API-KEY", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao enviar requisição: %v", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao enviar requisição", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("Falha ao criar template: %s", resp.Status)
		utils.SendResponse(w, http.StatusInternalServerError, "Falha ao criar template", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	utils.SendResponse(w, http.StatusOK, "Template criado com sucesso", nil, 0)
}
