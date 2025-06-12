package spacedesk

import (
	"api/database"
	"api/utils"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type TemplateRequest struct {
	Name                string      `json:"name"`
	Language            string      `json:"language"`
	Category            string      `json:"category"`
	AllowCategoryChange bool        `json:"allow_category_change,omitempty"`
	Components          []Component `json:"components"`
}

type Component struct {
	Type    string   `json:"type"`
	Format  string   `json:"format,omitempty"`
	Text    string   `json:"text,omitempty"`
	Buttons []Button `json:"buttons,omitempty"`
}

type Button struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func CreateOneTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse do request
	var templateReq TemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&templateReq); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Enviar para 360dialog
	apiKey := os.Getenv("SPACE_DESK_API_KEY")
	if apiKey == "" {
		http.Error(w, "API Key da 360dialog não configurada.", http.StatusInternalServerError)
		return
	}

	requestBody, err := json.Marshal(templateReq)
	if err != nil {
		http.Error(w, "Erro ao preparar a requisição: "+err.Error(), http.StatusInternalServerError)
		return
	}

	const threeSixtyDialogURL = "https://waba-v2.360dialog.io/v1/configs/templates"
	req360, err := http.NewRequest(http.MethodPost, threeSixtyDialogURL, bytes.NewBuffer(requestBody))
	if err != nil {
		http.Error(w, "Erro ao criar a requisição para a API externa: "+err.Error(), http.StatusInternalServerError)
		return
	}

	req360.Header.Set("Content-Type", "application/json")
	req360.Header.Set("D360-API-KEY", apiKey)
	client := &http.Client{}
	resp, err := client.Do(req360)
	if err != nil {
		http.Error(w, "Erro ao comunicar com a API da 360dialog: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Erro ao ler a resposta da API externa.", http.StatusInternalServerError)
		return
	}

	// 3. Verificar status
	var tplResp struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(responseBody, &tplResp)

	if tplResp.Status == "rejected" {
		// Só retorna o erro, não salva!
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(responseBody)
		return
	}

	// 4. Só agora salva no banco local
	readyMessage := ReadyMessage{
		Title:    templateReq.Name,
		Messages: []string{},
	}
	for _, comp := range templateReq.Components {
		if strings.ToLower(comp.Type) == "body" && comp.Text != "" {
			readyMessage.Messages = append(readyMessage.Messages, comp.Text)
		}
	}
	if len(readyMessage.Messages) == 0 {
		readyMessage.Messages = []string{"(sem mensagem de corpo definida)"}
	}

	mongoUri := os.Getenv(utils.MONGODB_URI)
	if mongoUri == "" {
		http.Error(w, "MongoDB URI não configurado.", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	clientOpts := options.Client().ApplyURI(mongoUri)
	dbClient, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Println("Erro ao conectar ao MongoDB:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao conectar ao MongoDB: "+err.Error(), nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer dbClient.Disconnect(ctx)

	col := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_READY_MESSAGE)
	_, err = col.InsertOne(ctx, bson.M{
		"titulo":     readyMessage.Title,
		"menssagens": readyMessage.Messages,
		"createdAt":  time.Now().UTC(),
	})

	if err != nil {
		http.Error(w, "Erro ao salvar template no banco: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Repassa a resposta da 360dialog ao cliente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(responseBody)
}
