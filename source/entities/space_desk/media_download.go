package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
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

// Estrutura do cache para arquivos binários
type fileCacheItem struct {
	data        []byte
	mimeType    string
	disposition string
	timestamp   time.Time
}

var (
	fileCache      = make(map[string]fileCacheItem)
	fileCacheMutex sync.RWMutex
	fileCacheTTL   = 2 * time.Hour
)

func init() {
	go func() {
		for {
			time.Sleep(30 * time.Minute)
			now := time.Now()
			fileCacheMutex.Lock()
			for k, v := range fileCache {
				if now.Sub(v.timestamp) > fileCacheTTL {
					delete(fileCache, k)
				}
			}
			fileCacheMutex.Unlock()
		}
	}()
}

// fetchMetadata tenta buscar metadata e devolve code10=true se vier FacebookApiException code 10
func fetchMetadata(client *http.Client, url, apiKey string) (
	meta struct {
		URL      string `json:"url"`
		MimeType string `json:"mime_type"`
	},
	code10 bool,
	err error,
) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return meta, false, fmt.Errorf("criação requisição metadata: %w", err)
	}
	req.Header.Set("D360-API-KEY", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return meta, false, fmt.Errorf("fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		// tenta ler code 10 do JSON de erro
		var fbErr struct {
			Error struct {
				Code int `json:"code"`
			} `json:"error"`
		}
		_ = json.Unmarshal(body, &fbErr)
		if fbErr.Error.Code == 10 {
			return meta, true, nil
		}
		return meta, false, fmt.Errorf("metadata non-OK %d: %s", resp.StatusCode, string(body))
	}

	// decodifica metadata válida
	if err := json.Unmarshal(body, &meta); err != nil {
		return meta, false, fmt.Errorf("decode metadata: %w", err)
	}
	return meta, false, nil
}

func clearFileCache(mediaID string) {
	fileCacheMutex.Lock()
	delete(fileCache, mediaID)
	fileCacheMutex.Unlock()
}

func HandlerMediaDownload(w http.ResponseWriter, r *http.Request) {
	mediaID := r.URL.Query().Get("media_id")
	chatID := r.URL.Query().Get("chat_id")
	if mediaID == "" || chatID == "" {
		http.Error(w, "media_id e chat_id são obrigatórios", http.StatusBadRequest)
		return
	}

	// --- busca no MongoDB a chave primária/secundária ---
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

	var chatDoc struct {
		CompanyPhoneNumber string `bson:"company_phone_number"`
	}
	objID, _ := bson.ObjectIDFromHex(chatID)
	col := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
	if err := col.FindOne(ctx, bson.M{"_id": objID}).Decode(&chatDoc); err != nil {
		clearFileCache(mediaID)
		if err == mongo.ErrNoDocuments {
			utils.SendResponse(w, http.StatusNotFound, "Chat não encontrado", nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		} else {
			utils.SendResponse(w, http.StatusInternalServerError, "Erro ao buscar chat: "+err.Error(), nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		}
		return
	}

	primaryKey := os.Getenv(utils.SPACE_DESK_API_KEY)
	altKey := os.Getenv(utils.SPACE_DESK_API_KEY_2)
	if chatDoc.CompanyPhoneNumber == "5511958339942" {
		primaryKey, altKey = altKey, primaryKey
	}

	// --- check cache ---
	fileCacheMutex.RLock()
	if item, ok := fileCache[mediaID]; ok && time.Since(item.timestamp) < fileCacheTTL {
		fileCacheMutex.RUnlock()
		w.Header().Set("Content-Type", item.mimeType)
		w.Header().Set("Content-Disposition", item.disposition)
		w.WriteHeader(http.StatusOK)
		w.Write(item.data)
		return
	}
	fileCacheMutex.RUnlock()

	client := &http.Client{Timeout: 15 * time.Second}
	metaURL := "https://waba-v2.360dialog.io/" + mediaID

	// 1ª tentativa de metadata
	meta, code10, ferr := fetchMetadata(client, metaURL, primaryKey)
	if ferr != nil && !code10 {
		log.Printf("[MediaDownload][%s] metadata error: %v", mediaID, ferr)
		clearFileCache(mediaID)
		http.Error(w, "Erro ao buscar metadata", http.StatusBadGateway)
		return
	}
	// se permission denied (code10), tenta com a chave secundária
	usedKey := primaryKey
	if code10 {
		log.Printf("[MediaDownload][%s] permission denied, retrying with alternate key", mediaID)
		meta, code10, ferr = fetchMetadata(client, metaURL, altKey)
		if ferr != nil || code10 {
			log.Printf("[MediaDownload][%s] retry failed: err=%v code10=%v", mediaID, ferr, code10)
			clearFileCache(mediaID)
			http.Error(w, "Permission denied na API", http.StatusForbidden)
			return
		}
		usedKey = altKey
	}

	// --- faz download do binário ---
	downloadURL := strings.Replace(meta.URL, "https://lookaside.fbsbx.com", "https://waba-v2.360dialog.io", 1)
	reqFile, _ := http.NewRequest(http.MethodGet, downloadURL, nil)
	reqFile.Header.Set("D360-API-KEY", usedKey)
	respFile, err := client.Do(reqFile)
	if err != nil || respFile.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(respFile.Body)
		log.Printf("[MediaDownload][%s] file fetch error status=%d err=%v body=%s",
			mediaID, respFile.StatusCode, err, string(body))
		clearFileCache(mediaID)
		http.Error(w, "Erro ao baixar arquivo", http.StatusBadGateway)
		return
	}
	defer respFile.Body.Close()

	data, err := io.ReadAll(respFile.Body)
	if err != nil {
		log.Printf("[MediaDownload][%s] file read error: %v", mediaID, err)
		clearFileCache(mediaID)
		http.Error(w, "Erro ao ler arquivo", http.StatusBadGateway)
		return
	}

	disp := respFile.Header.Get("Content-Disposition")
	if disp == "" {
		disp = "attachment; filename=\"arquivo\""
	}

	// --- salva no cache e devolve ---
	fileCacheMutex.Lock()
	fileCache[mediaID] = fileCacheItem{data: data, mimeType: meta.MimeType, disposition: disp, timestamp: time.Now()}
	fileCacheMutex.Unlock()

	w.Header().Set("Content-Type", meta.MimeType)
	w.Header().Set("Content-Disposition", disp)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
