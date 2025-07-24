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

	"github.com/redis/go-redis/v9"
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

	// Configura Redis
	var rdb *redis.Client
	redisURI := os.Getenv("REDIS_URI")
	opts, err := redis.ParseURL(redisURI)
	if err == nil {
		rdb = redis.NewClient(opts)
		defer rdb.Close()
	}

	redisKey := "spacedesk:media:env:" + mediaID
	cachedEnv := ""
	if rdb != nil {
		val, err := rdb.Get(ctx, redisKey).Result()
		if err == nil {
			cachedEnv = val
		}
	}

	primaryKey := os.Getenv(utils.SPACE_DESK_API_KEY)
	altKey := os.Getenv(utils.SPACE_DESK_API_KEY_2)

	// 1) Cache lookup
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

	usedKey := primaryKey
	if cachedEnv == "2" {
		usedKey = altKey
	}

	meta, code10, ferr := fetchMeta(client, metaURL, usedKey)
	if ferr != nil && !code10 {
		log.Printf("[Base64] metadata error: %v", ferr)
		http.Error(w, "falha metadata", http.StatusBadGateway)
		return
	}

	if code10 && cachedEnv == "" {
		log.Printf("[Base64] Permission denied com primary, retrying with fallback key")
		meta, code10, ferr = fetchMeta(client, metaURL, altKey)
		if ferr != nil || code10 {
			log.Printf("[Base64] fallback also failed: %v code10=%v", ferr, code10)
			http.Error(w, "Permission denied na API", http.StatusForbidden)
			return
		}
		usedKey = altKey
		if rdb != nil {
			rdb.Set(ctx, redisKey, "2", 90*24*time.Hour)
		}
	} else if cachedEnv == "" && rdb != nil {
		rdb.Set(ctx, redisKey, "1", 90*24*time.Hour)
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
