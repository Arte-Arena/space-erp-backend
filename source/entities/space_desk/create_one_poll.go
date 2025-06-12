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

type CreatePollRequest struct {
	To     string   `json:"to"`
	UserId string   `json:"userId"`
	Poll   PollBody `json:"poll"`
}

type PollBody struct {
	Name                   string   `json:"name"`
	Options                []string `json:"options"`
	SelectableOptionsCount int      `json:"selectable_options_count"`
}

func CreateOnePoll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var reqBody CreatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "JSON inválido: "+err.Error(), nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	if reqBody.To == "" || reqBody.Poll.Name == "" || len(reqBody.Poll.Options) == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "Campos obrigatórios ausentes", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	clientOpts := options.Client().ApplyURI(mongoURI)
	dbClient, err := mongo.Connect(clientOpts)
	if err != nil {
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
		utils.SendResponse(w, http.StatusBadRequest, "ID do chat inválido", nil, utils.INVALID_CHAT_ID_FORMAT)
		return
	}
	err = colChats.FindOne(r.Context(), bson.M{"_id": objID}).Decode(&chatDoc)
	if err != nil {
		utils.SendResponse(w, http.StatusNotFound, "Chat não encontrado", nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		return
	}
	recipient := chatDoc.ClientePhoneNumber

	// Monta as opções para o payload do 360dialog
	pollOptions := []map[string]string{}
	for _, o := range reqBody.Poll.Options {
		pollOptions = append(pollOptions, map[string]string{"option": o})
	}
	// var btns []map[string]interface{}
	// for i, option := range reqBody.Poll.Options {
	// 	btns = append(btns, map[string]interface{}{
	// 		"type": "reply",
	// 		"reply": map[string]interface{}{
	// 			"id":    fmt.Sprintf("option_%d", i+1),
	// 			"title": option,
	// 		},
	// 	})
	// }

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                recipient,
		"type":              "interactive",
		"interactive": map[string]interface{}{
			"type": "button",
			"body": map[string]string{
				"text": reqBody.Poll.Name, // ou outra mensagem
			},
			"action": map[string]interface{}{
				"buttons": func() []map[string]interface{} {
					var btns []map[string]interface{}
					for i, option := range reqBody.Poll.Options {
						btns = append(btns, map[string]interface{}{
							"type": "reply",
							"reply": map[string]interface{}{
								"id":    fmt.Sprintf("option_%d", i+1),
								"title": option,
							},
						})
					}
					return btns
				}(),
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao serializar payload: "+err.Error(), nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	apiKey := os.Getenv(utils.SPACE_DESK_API_KEY)
	if apiKey == "" {
		utils.SendResponse(w, http.StatusInternalServerError, "API key não configurada", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}

	req360, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://waba-v2.360dialog.io/messages",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao criar requisição externa: "+err.Error(), nil, utils.ERROR_TO_CREATE_EXTERNAL_CONNECTION)
		return
	}
	req360.Header.Set("Content-Type", "application/json")
	req360.Header.Set("Accept", "application/json")
	req360.Header.Set("D360-API-KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req360)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "Falha ao enviar mensagem", nil, utils.ERROR_TO_SEND_MESSAGE)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Resposta da 360dialog: %s", string(respBody))

	respMap := make(map[string]any)
	if err := json.Unmarshal(respBody, &respMap); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao ler resposta externa", nil, utils.ERROR_TO_READ_MESSAGE)
		return
	}
	wamid := extractWamid(respMap)
	if wamid == "not_returned" {
		utils.SendResponse(w, http.StatusBadGateway, "A API não retornou o wamid, mensagem não salva.", nil, utils.ERROR_TO_SEND_MESSAGE)
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
									"type":      "poll",
									"from":      "space-erp-backend",
									"to":        reqBody.To,
									"id":        wamid,
									"timestamp": fmt.Sprint(now.Unix()),
									"poll": bson.M{
										"name":                     reqBody.Poll.Name,
										"options":                  reqBody.Poll.Options,
										"selectable_options_count": reqBody.Poll.SelectableOptionsCount,
									},
									"user": reqBody.UserId,
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
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao inserir evento no MongoDB: "+err.Error(), nil, utils.ERROR_TO_INSERT_IN_MONGODB)
		return
	}

	broadcastSpaceDeskMessage(respMap)
	utils.SendResponse(w, http.StatusCreated, "", respMap, 0)
}
