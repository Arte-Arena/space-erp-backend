package spacedesk

import (
	"api/database"
	"api/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CreateMessageRequest struct {
	To   string `json:"to"`
	Body string `json:"body"`
}

func CreateOneMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var reqBody CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Println("Erro ao decodificar JSON do corpo da requisição:", err)
		utils.SendResponse(w, http.StatusBadRequest, "JSON inválido: "+err.Error(), nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	if reqBody.To == "" || reqBody.Body == "" {
		log.Println("Campos obrigatórios 'to' ou 'body' ausentes no corpo da requisição")
		utils.SendResponse(w, http.StatusBadRequest, "Campos 'to' e 'body' são obrigatórios", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
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

	colChats := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT_METADATA)
	var chatDoc struct {
		ClientePhoneNumber string `bson:"cliente_phone_number"`
	}

	objID, err := bson.ObjectIDFromHex(reqBody.To)
	if err != nil {
		log.Println("Erro ao converter ID do chat para ObjectID:", err, "ID recebido:", reqBody.To)
		// É uma boa prática ter um código de erro específico para ID inválido
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

	// Usar o número do cliente como destinatário
	recipient := chatDoc.ClientePhoneNumber

	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                recipient,
		"type":              "text",
		"text": map[string]string{
			"body": reqBody.Body,
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("Erro ao serializar payload:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao serializar payload: "+err.Error(), nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	apiKey := os.Getenv(utils.SPACE_DESK_API_KEY)
	if apiKey == "" {
		log.Println("API key não configurada (variável de ambiente não encontrada)")
		utils.SendResponse(w, http.StatusInternalServerError, "API key não configurada", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}

	req360, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://waba-v2.360dialog.io/messages",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		log.Println("Erro ao criar requisição externa:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao criar requisição externa: "+err.Error(), nil, utils.ERROR_TO_CREATE_EXTERNAL_CONNECTION)
		return
	}
	req360.Header.Set("Content-Type", "application/json")
	req360.Header.Set("Accept", "application/json")
	req360.Header.Set("D360-API-KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req360)
	if err != nil {
		log.Println("Falha ao enviar mensagem:", err)
		utils.SendResponse(w, http.StatusBadGateway, "Falha ao enviar mensagem", nil, utils.ERROR_TO_SEND_MESSAGE)
		return
	}
	defer resp.Body.Close()

	respMap := make(map[string]any)
	if err := json.NewDecoder(resp.Body).Decode(&respMap); err != nil {
		log.Println("Erro ao ler resposta externa:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao ler resposta externa", nil, utils.ERROR_TO_READ_MESSAGE)
		return
	}

	now := time.Now().UTC()
	raw := bson.M{
		"entry": []any{
			bson.M{
				"changes": []any{
					bson.M{
						"field": "messages",
						"value": bson.M{
							"messages": []any{
								bson.M{
									"from":      "space-erp-backend",
									"to":        reqBody.To,
									"timestamp": fmt.Sprint(now.Unix()),
									"text":      bson.M{"body": reqBody.Body},
								},
							},
						},
					},
				},
			},
		},
	}
	col := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_EVENTS_WHATSAPP)
	_, err = col.InsertOne(ctx, raw)
	if err != nil {
		log.Println("Erro ao inserir evento no MongoDB:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao inserir evento no MongoDB: "+err.Error(), nil, utils.ERROR_TO_INSERT_IN_MONGODB)
		return
	}

	broadcastSpaceDeskMessage(respMap)

	utils.SendResponse(w, http.StatusCreated, "", respMap, 0)
}
