package spacedesk

import (
	"api/database"
	"api/utils"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ListTemplatesResp struct {
	Templates []D360Template `json:"waba_templates"`
}

type D360Template struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Language   string      `json:"language"`
	Category   string      `json:"category"`
	Status     string      `json:"status"`
	Components []Component `json:"components"`
}

func ListAndSyncD360Templates(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("SPACE_DESK_API_KEY")
	if apiKey == "" {
		http.Error(w, "API Key da 360dialog não configurada.", http.StatusInternalServerError)
		return
	}

	// 1. Chamar API da D360
	req, err := http.NewRequest("GET", "https://waba-v2.360dialog.io/v1/configs/templates", nil)
	if err != nil {
		http.Error(w, "Erro ao criar request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("D360-API-KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Erro ao chamar API da D360: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Logar resposta bruta da API D360
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var tplResp ListTemplatesResp
	if err := json.NewDecoder(resp.Body).Decode(&tplResp); err != nil {
		http.Error(w, "Erro ao ler resposta da D360: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Conectar no MongoDB
	mongoUri := os.Getenv(utils.MONGODB_URI)
	if mongoUri == "" {
		http.Error(w, "MongoDB URI não configurado.", http.StatusInternalServerError)
		return
	}
	ctx := context.Background()
	clientOpts := options.Client().ApplyURI(mongoUri)
	dbClient, err := mongo.Connect(clientOpts)
	if err != nil {
		http.Error(w, "Erro ao conectar no MongoDB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dbClient.Disconnect(ctx)
	col := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_READY_MESSAGE)

	// 3. Para cada template aprovado/pending, tenta inserir se não existir
	for _, tpl := range tplResp.Templates {
		if tpl.Status == "rejected" {
			continue
		}
		// Verifica se já existe pelo nome
		count, err := col.CountDocuments(ctx, bson.M{"titulo": tpl.Name})
		if err != nil {
			continue // opcional: logar erro e seguir
		}
		if count == 0 {
			// Extrai o texto do body (se existir)
			newTitle := "R$ " + tpl.Name
			countWithPrefix, err := col.CountDocuments(ctx, bson.M{"titulo": newTitle})
			if err != nil || countWithPrefix > 0 {
				continue
			}

			body := ""
			for _, c := range tpl.Components {
				if strings.ToLower(c.Type) == "body" && c.Text != "" {
					body = c.Text
					break
				}
			}
			if body == "" {
				body = "(sem mensagem de corpo definida)"
			}
			_, err = col.InsertOne(ctx, bson.M{
				"titulo":     newTitle,
				"menssagens": []string{body},
				"createdAt":  time.Now().UTC(),
			})
			if err != nil {
				// opcional: logar erro
				continue
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tplResp)
}
