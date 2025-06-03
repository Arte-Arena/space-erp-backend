// extchat/media_base64.go
package spacedesk

import (
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

func HandlerMediaBase64(w http.ResponseWriter, r *http.Request) {
	mediaID := r.URL.Query().Get("media_id")
	if mediaID == "" {
		http.Error(w, "media_id é obrigatório", http.StatusBadRequest)
		return
	}

	// BUSCA DO CACHE PRIMEIRO!
	base64CacheMutex.RLock()
	cache, ok := base64Cache[mediaID]
	base64CacheMutex.RUnlock()

	if ok && time.Since(cache.timestamp) < base64CacheTTL {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"base64": cache.data})
		return
	}

	apiKey := os.Getenv("SPACE_DESK_API_KEY")
	client := &http.Client{Timeout: 15 * time.Second}

	// 1) GET metadata no endpoint correto (sem /v1/media/)
	metaURL := fmt.Sprintf("https://waba-v2.360dialog.io/%s", mediaID)
	reqMeta, _ := http.NewRequest(http.MethodGet, metaURL, nil)
	reqMeta.Header.Set("D360-API-KEY", apiKey)

	respMeta, err := client.Do(reqMeta)
	if err != nil {
		log.Printf("[Base64] erro fetch metadata: %v", err)
		http.Error(w, "falha metadata", http.StatusBadGateway)
		return
	}
	defer respMeta.Body.Close()

	if respMeta.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(respMeta.Body)
		log.Printf("[Base64] metadata retornou status %d: %s", respMeta.StatusCode, string(body))
		http.Error(w, "falha metadata", http.StatusBadGateway)
		return
	}

	var meta struct {
		URL      string `json:"url"`
		MimeType string `json:"mime_type"`
	}
	if err := json.NewDecoder(respMeta.Body).Decode(&meta); err != nil {
		log.Printf("[Base64] decode metadata: %v", err)
		http.Error(w, "metadata inválida", http.StatusBadGateway)
		return
	}

	// 2) Usa a URL assinada que veio em meta.URL
	signed := meta.URL
	// Se ainda for lookaside, faz a troca de host:
	signed = strings.Replace(
		signed,
		"https://lookaside.fbsbx.com",
		"https://waba-v2.360dialog.io",
		1,
	)

	reqImg, _ := http.NewRequest(http.MethodGet, signed, nil)
	reqImg.Header.Set("D360-API-KEY", apiKey)
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

	// 3) lê tudo e converte em base64
	data, err := io.ReadAll(respImg.Body)
	if err != nil {
		log.Printf("[Base64] erro read img: %v", err)
		http.Error(w, "falha read img", http.StatusBadGateway)
		return
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", meta.MimeType, b64)

	// ------> SALVA NO CACHE
	base64CacheMutex.Lock()
	base64Cache[mediaID] = base64CacheItem{
		data:      dataURI,
		timestamp: time.Now(),
	}
	base64CacheMutex.Unlock()

	// 4) devolve JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"base64": dataURI, "mimeType": meta.MimeType})
}
