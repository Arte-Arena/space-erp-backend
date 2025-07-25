package spacedesk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"time"

	"api/database"
	"api/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateOneMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Erro no parse do form: "+err.Error(), http.StatusBadRequest)
		return
	}
	chatId := r.FormValue("to")
	if chatId == "" {
		http.Error(w, "Parâmetro 'chatId' ausente", http.StatusBadRequest)
		return
	}

	userId := r.FormValue("userId")
	if userId == "" {
		http.Error(w, "Parâmetro 'userId' ausente", http.StatusBadRequest)
		return
	}

	mediaType := r.FormValue("type")
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Arquivo não encontrado: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Falha ao ler arquivo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	mediaId, err := uploadTo360(fileBytes, header)
	if err != nil {
		log.Printf("[SendMedia] upload error: %v", err)
		http.Error(w, "Falha no upload de mídia: "+err.Error(), http.StatusBadGateway)
		return
	}
	// pega o numero do telefone com base no id do chat
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	clientOpts := options.Client().ApplyURI(mongoURI)
	dbClient, err := mongo.Connect(clientOpts)
	if err != nil {
		http.Error(w, "Erro ao conectar ao MongoDB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dbClient.Disconnect(ctx)

	colChats := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
	var chatDoc struct {
		ClientePhoneNumber string `bson:"cliente_phone_number"`
		CompanyPhoneNumber string `bson:"company_phone_number"`
	}

	objID, err := bson.ObjectIDFromHex(chatId)
	if err != nil {
		http.Error(w, "ID de chat inválido", http.StatusBadRequest)
		return
	}

	err = colChats.FindOne(ctx, bson.M{"_id": objID}).Decode(&chatDoc)
	if err != nil {
		http.Error(w, "Chat não encontrado", http.StatusNotFound)
		return
	}

	to := chatDoc.ClientePhoneNumber

	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              mediaType,
		mediaType: map[string]string{
			"id": mediaId,
		},
	}
	bodyBytes, _ := json.Marshal(payload)

	var apiKey string
	switch chatDoc.CompanyPhoneNumber {
	case "5511958339942":
		apiKey = os.Getenv(utils.SPACE_DESK_API_KEY_2)
	case "551123371548":
		apiKey = os.Getenv(utils.SPACE_DESK_API_KEY)
	}

	req360, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"https://waba-v2.360dialog.io/messages",
		bytes.NewReader(bodyBytes),
	)
	req360.Header.Set("Content-Type", "application/json")
	req360.Header.Set("D360-API-KEY", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp360, err := client.Do(req360)
	if err != nil {
		log.Printf("[SendMedia] send message error: %v", err)
		http.Error(w, "Falha ao enviar mensagem", http.StatusBadGateway)
		return
	}
	defer resp360.Body.Close()

	bodyRaw, _ := io.ReadAll(resp360.Body)
	log.Printf("[SendMedia] 360dialog response: %s", string(bodyRaw))

	var respData struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(bodyRaw, &respData); err != nil {
		log.Printf("[SendMedia] Erro ao decodificar resposta do 360dialog: %v", err)
		http.Error(w, "Falha ao decodificar resposta do 360dialog", http.StatusBadGateway)
		return
	}

	if len(respData.Messages) == 0 || respData.Messages[0].ID == "" {
		log.Printf("[SendMedia] Nenhuma mensagem retornada pela API 360dialog.\n"+
			"Payload enviado: %s\n"+
			"Resp: %s", string(bodyBytes), string(bodyRaw))
		http.Error(w, "Falha ao enviar mídia: resposta vazia do 360dialog", http.StatusBadGateway) //deu erro aqui
		return
	}

	url, errUrl := fetchMediaURL(mediaId)
	if errUrl != nil {
		log.Printf("[SendMedia] Erro ao buscar URL da mídia: %v", errUrl)
		url = ""
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
		defer cancel()
		mongoURI := os.Getenv(utils.MONGODB_URI)
		clientOpts := options.Client().ApplyURI(mongoURI)
		dbClient, err := mongo.Connect(clientOpts)
		if err == nil {
			now := time.Now().UTC()
			timestampStr := fmt.Sprintf("%d", now.Unix())
			from := "551123371548"
			msgID := respData.Messages[0].ID

			messageEvent := bson.M{
				"from":      from,
				"to":        to,
				"id":        msgID,
				"timestamp": timestampStr,
				"type":      mediaType,
				"user":      userId,
			}
			switch mediaType {
			case "image":
				messageEvent["image"] = bson.M{"id": mediaId}
			case "document":
				messageEvent["document"] = bson.M{"id": mediaId}
			case "video":
				messageEvent["video"] = bson.M{"id": mediaId}
			case "audio":
				messageEvent["audio"] = bson.M{"id": mediaId}
			case "sticker":
				messageEvent["sticker"] = bson.M{"id": mediaId}
			}

			raw := bson.M{
				"entry": []any{
					bson.M{
						"id": "1343302196977353", // opcional: pode deixar vazio ou buscar de config/env
						"changes": []any{
							bson.M{
								"field": "messages",
								"value": bson.M{
									"messages": []any{messageEvent},
								},
							},
						},
					},
				},
			}

			col := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_EVENTS_WHATSAPP)
			if _, err := col.InsertOne(ctx, raw); err != nil {
				log.Printf("[SendMedia] Erro ao inserir no Mongo: %v", err)
			}

			newRaw := bson.M{
				"body":              mediaId,
				"media_id":          mediaId,
				"chat_id":           objID,
				"by":                userId,
				"from":              "company",
				"created_at":        time.Now().UTC(),
				"message_id":        msgID,
				"message_timestamp": fmt.Sprint(now.Unix()),
				"type":              mediaType,
				"status":            "",
				"updated_at":        time.Now().UTC().Format(time.RFC3339),
			}
			colMessages := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_MESSAGE)
			_, err = colMessages.InsertOne(ctx, newRaw)
			if err != nil {
				log.Println("Erro ao inserir evento no MongoDB:", err)
				utils.SendResponse(w, http.StatusInternalServerError, "Erro ao inserir evento no MongoDB: "+err.Error(), nil, utils.ERROR_TO_INSERT_IN_MONGODB)
				return
			}

			//chat
			filter := bson.M{"_id": objID}
			update := bson.M{
				"$set": bson.M{
					"last_message_id":        msgID,
					"last_message_timestamp": fmt.Sprint(now.Unix()),
					"last_message_excerpt":   mediaId,
					"last_message_type":      mediaType,
					"last_message_sender":    "company",
					"updated_at":             time.Now().UTC().Format(time.RFC3339),
				},
			}

			colChat := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
			updateOpts := options.UpdateOne().SetUpsert(false) //true caso puder criar
			_, err = colChat.UpdateOne(ctx, filter, update, updateOpts)
			if err != nil {
				log.Println("Erro ao inserir evento no MongoDB:", err)
				utils.SendResponse(w, http.StatusInternalServerError, "Erro ao inserir evento no MongoDB: "+err.Error(), nil, utils.ERROR_TO_INSERT_IN_MONGODB)
				return
			}
		}
	}()

	broadcastSpaceDeskMessage(map[string]any{
		"id":   chatId,
		"from": "company",
		"to":   to,
		"messages": []map[string]string{{
			"id":      respData.Messages[0].ID,
			"mediaId": mediaId,
			"url":     url,
			"type":    mediaType,
		}},
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":   chatId,
		"from": "company",
		"to":   to,
		"messages": []map[string]string{{
			"id":      respData.Messages[0].ID,
			"mediaId": mediaId,
			"url":     url,
			"type":    mediaType,
		}},
	})
}

func uploadTo360(fileBytes []byte, header *multipart.FileHeader) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("messaging_product", "whatsapp")

	fileContentType := header.Header.Get("Content-Type")
	if fileContentType == "" {
		fileContentType = "image/jpeg"
	}

	formFileHeader := textproto.MIMEHeader{}
	formFileHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, header.Filename))
	formFileHeader.Set("Content-Type", fileContentType)

	part, err := writer.CreatePart(formFileHeader)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(fileBytes); err != nil {
		return "", err
	}
	writer.Close()

	req, err := http.NewRequest(
		http.MethodPost,
		"https://waba-v2.360dialog.io/media",
		body,
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("D360-API-KEY", os.Getenv(utils.SPACE_DESK_API_KEY))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[SendMedia] uploadTo360 response: %s", string(respBody))

	var data struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(respBody, &data)
	if data.ID == "" {
		return "", fmt.Errorf("ID da mídia vazio: resposta inesperada de upload 360dialog")
	}
	return data.ID, nil
}

func fetchMediaURL(mediaId string) (string, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://waba-v2.360dialog.io/media/%s", mediaId),
		nil,
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("D360-API-KEY", os.Getenv(utils.SPACE_DESK_API_KEY))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var body struct {
		Url string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	return body.Url, nil
}
