package spacedesk

import (
	"api/database"
	"api/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
		utils.SendResponse(w, http.StatusBadRequest, "JSON inválido: "+err.Error(), nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	if reqBody.To == "" || reqBody.Body == "" {
		utils.SendResponse(w, http.StatusBadRequest, "Campos 'to' e 'body' são obrigatórios", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                reqBody.To,
		"type":              "text",
		"text": map[string]string{
			"body": reqBody.Body,
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao serializar payload: "+err.Error(), nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	apiKey := os.Getenv(utils.D360_API_KEY)
	if apiKey == "" {
		utils.SendResponse(w, http.StatusInternalServerError, "API key não configurada", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), database.MONGO_TIMEOUT)
	defer cancel()

	req360, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://waba-v2.360dialog.io/messages",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao criar requisição externa: "+err.Error(), nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	req360.Header.Set("Content-Type", "application/json")
	req360.Header.Set("Accept", "application/json")
	req360.Header.Set("D360-API-KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req360)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "Falha ao enviar mensagem", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	defer resp.Body.Close()

	respMap := make(map[string]any)
	if err := json.NewDecoder(resp.Body).Decode(&respMap); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao ler resposta externa", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	go func(reqBody CreateMessageRequest) {
		ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
		defer cancel()

		mongoURI := os.Getenv(utils.MONGODB_URI)
		clientOpts := options.Client().ApplyURI(mongoURI)
		dbClient, err := mongo.Connect(clientOpts)
		if err != nil {
			return
		}
		defer dbClient.Disconnect(ctx)

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
		_, _ = col.InsertOne(ctx, raw)
	}(reqBody)

	broadcastSpaceDeskMessage(respMap)

	utils.SendResponse(w, http.StatusCreated, "", respMap, 0)
}
