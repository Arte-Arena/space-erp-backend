package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type base64CacheItem struct {
	data      string
	timestamp time.Time
}

var (
	base64Cache      = make(map[string]base64CacheItem)
	base64CacheMutex sync.RWMutex
	base64CacheTTL   = 2 * time.Hour
)

func cleanupBase64Cache() {
	for {
		time.Sleep(30 * time.Minute)
		now := time.Now()
		base64CacheMutex.Lock()
		for k, v := range base64Cache {
			if now.Sub(v.timestamp) > base64CacheTTL {
				delete(base64Cache, k)
			}
		}
		base64CacheMutex.Unlock()
	}
}

func init() {
	go cleanupBase64Cache()
}

// fetchMeta tenta buscar a metadata e sinaliza se veio code 10
func fetchMeta(client *http.Client, url, apiKey string) (meta struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
}, code10 bool, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return meta, false, fmt.Errorf("criando request metadata: %w", err)
	}
	req.Header.Set("D360-API-KEY", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		return meta, false, fmt.Errorf("fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var fbErr struct {
			Error struct{ Code int } `json:"error"`
		}
		_ = json.Unmarshal(body, &fbErr)
		if fbErr.Error.Code == 10 {
			return meta, true, nil
		}
		return meta, false, fmt.Errorf("metadata non-OK %d: %s", resp.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, &meta); err != nil {
		return meta, false, fmt.Errorf("decode metadata: %w", err)
	}
	return meta, false, nil
}

func HandlerMediaBase64(w http.ResponseWriter, r *http.Request) {
	mediaID := r.URL.Query().Get("media_id")
	chatID := r.URL.Query().Get("chat_id")

	if mediaID == "" || chatID == "" {
		http.Error(w, "media_id e chat_id são obrigatórios", http.StatusBadRequest)
		return
	}

	// obter chave do numero pelo qual esta midia foi enviada
	// na collection space_desk_chat, temos o campo company_phone_number
	// e o campo company_phone_number_2 que é o numero do whatsapp do suporte
	// temos o midia id e podemos dar find no campo body da collection space_desk_messages

	ctx, cancel := context.WithTimeout(r.Context(), database.MONGO_TIMEOUT)
	defer cancel()
	mongoURI := os.Getenv(utils.MONGODB_URI)
	clientOpts := options.Client().ApplyURI(mongoURI)
	dbClient, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Printf("[MediaDownload][%s] mongo connect error: %v", mediaID, err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao conectar ao MongoDB: "+err.Error(), nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer dbClient.Disconnect(ctx)

	// Buscar informações do chat
	var chatDoc struct {
		CompanyPhoneNumber string `bson:"company_phone_number"`
	}
	objID, err := bson.ObjectIDFromHex(chatID)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "Chat ID inválido", nil, utils.CANNOT_FIND_SPACE_DESK_CHAT_ID)
		return
	}

	col := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
	if err := col.FindOne(ctx, bson.M{"_id": objID}).Decode(&chatDoc); err != nil {
		if err == mongo.ErrNoDocuments {
			utils.SendResponse(w, http.StatusNotFound, "Chat não encontrado", nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		} else {
			utils.SendResponse(w, http.StatusInternalServerError, "Erro ao buscar chat: "+err.Error(), nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		}
		return
	}

	// Buscar a mensagem específica para verificar se foi enviada pela empresa ou pelo cliente
	var messageDoc struct {
		From string `bson:"from"`
	}
	colMessages := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_MESSAGE)

	// Tentar encontrar a mensagem pelo media_id primeiro (mensagens do cliente)
	err = colMessages.FindOne(ctx, bson.M{"media_id": mediaID}).Decode(&messageDoc)
	if err != nil {
		// Se não encontrar pelo media_id, tentar pelo body (mensagens da empresa)
		err = colMessages.FindOne(ctx, bson.M{"body": mediaID}).Decode(&messageDoc)
		if err != nil {
			log.Printf("[MediaDownload][%s] message not found by media_id or body: %v", mediaID, err)
			// Se não encontrar a mensagem, usar a lógica baseada no company_phone_number
		} else {
			log.Printf("[MediaDownload][%s] message found by body (company message)", mediaID)
		}
	} else {
		log.Printf("[MediaDownload][%s] message found by media_id (client message)", mediaID)
	}

	primaryKey := os.Getenv(utils.SPACE_DESK_API_KEY)
	altKey := os.Getenv(utils.SPACE_DESK_API_KEY_2)

	// Determinar qual chave usar baseado na lógica do sistema
	var usedKey string
	var keyReason string

	if chatDoc.CompanyPhoneNumber == "5511958339942" {
		usedKey = altKey
		keyReason = "client message (last_message_sender)"

		if messageDoc.From == "company" {
			usedKey = primaryKey
			keyReason = "company message (from field)"
		}

	} else {
		usedKey = primaryKey
		keyReason = "default phone (551123371548)"
	}

	// e verificar se a mensagem é enviada pelo cliente ou pelo company
	// caso company env 1
	// caso client, verificamos se o campo company_phone_number é o 1 ou o 2
	// se for o 1, usamos a chave primária
	// se for o 2, usamos a chave secundária

	// Log adicional para debug
	log.Printf("[MediaDownload][%s] Debug - chatDoc.CompanyPhoneNumber: %s, messageDoc.From: %s, keyReason: %s",
		mediaID, chatDoc.CompanyPhoneNumber, messageDoc.From, keyReason)

	base64CacheMutex.RLock()
	if item, ok := base64Cache[mediaID]; ok && time.Since(item.timestamp) < base64CacheTTL {
		base64CacheMutex.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"base64": item.data})
		return
	}
	base64CacheMutex.RUnlock()

	client := &http.Client{Timeout: 15 * time.Second}
	metaURL := fmt.Sprintf("https://waba-v2.360dialog.io/%s", mediaID)

	// Tentar com a chave determinada
	meta, code10, ferr := fetchMeta(client, metaURL, usedKey)
	if ferr != nil && !code10 {
		log.Printf("[Base64] metadata error: %v", ferr)
		http.Error(w, "falha metadata", http.StatusBadGateway)
		return
	}

	// Se der erro de permissão (code 10), tentar com a chave alternativa
	if code10 {
		log.Printf("[Base64] Permission denied com chave selecionada para %s", usedKey)
		utils.SendResponse(w, http.StatusForbidden, "Permission denied na API", nil, 3)
		return
	}

	signed := strings.Replace(meta.URL,
		"https://lookaside.fbsbx.com",
		"https://waba-v2.360dialog.io",
		1,
	)
	reqImg, _ := http.NewRequest(http.MethodGet, signed, nil)
	reqImg.Header.Set("D360-API-KEY", usedKey)
	respImg, err := client.Do(reqImg)
	if err != nil {
		log.Printf("[Base64] erro fetch imagem: %v", err)
		http.Error(w, "falha fetch img", http.StatusBadGateway)
		return
	}
	defer respImg.Body.Close()

	if respImg.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(respImg.Body)
		log.Printf("[Base64] imagem retornou status %d: %s", respImg.StatusCode, string(body))
		http.Error(w, "falha fetch img", http.StatusBadGateway)
		return
	}

	// 7) Converte para base64
	data, err := io.ReadAll(respImg.Body)
	if err != nil {
		log.Printf("[Base64] erro read img: %v", err)
		http.Error(w, "falha read img", http.StatusBadGateway)
		return
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", meta.MimeType, b64)

	// 8) Salva no cache
	base64CacheMutex.Lock()
	base64Cache[mediaID] = base64CacheItem{data: dataURI, timestamp: time.Now()}
	base64CacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"base64": dataURI})
}
