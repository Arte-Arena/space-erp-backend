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
	to := r.FormValue("to")
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

	req360, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"https://waba-v2.360dialog.io/messages",
		bytes.NewReader(bodyBytes),
	)
	req360.Header.Set("Content-Type", "application/json")
	req360.Header.Set("D360-API-KEY", os.Getenv(utils.D360_API_KEY))

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
		log.Printf("[SendMedia] Nenhuma mensagem retornada pela API 360dialog. Resp: %s", string(bodyRaw))
		http.Error(w, "Falha ao enviar mídia: resposta vazia do 360dialog", http.StatusBadGateway)
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
											"to":        to,
											"timestamp": fmt.Sprint(now.Unix()),
											mediaType:   bson.M{"id": mediaId},
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
			_ = dbClient.Disconnect(ctx)
		}
	}()

	broadcastSpaceDeskMessage(map[string]any{
		"messages": []map[string]string{{
			"id":  respData.Messages[0].ID,
			"url": url,
		}},
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": []map[string]string{{
			"id":  respData.Messages[0].ID,
			"url": url,
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
	req.Header.Set("D360-API-KEY", os.Getenv(utils.D360_API_KEY))
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
	req.Header.Set("D360-API-KEY", os.Getenv(utils.D360_API_KEY))

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
