package spacedesk

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Estrutura do cache para arquivos binários
type fileCacheItem struct {
	data        []byte
	mimeType    string
	disposition string // content-disposition header
	timestamp   time.Time
}

var (
	fileCache      = make(map[string]fileCacheItem)
	fileCacheMutex sync.RWMutex
	fileCacheTTL   = 2 * time.Hour
)

// Limpa arquivos expirados do cache a cada 30 minutos
func cleanupFileCache() {
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
}

func init() {
	go cleanupFileCache()
}

func HandlerMediaDownload(w http.ResponseWriter, r *http.Request) {
	mediaId := r.URL.Query().Get("media_id")
	if mediaId == "" {
		http.Error(w, "media_id obrigatório", http.StatusBadRequest)
		return
	}

	// Tenta buscar do cache antes de chamar a API
	fileCacheMutex.RLock()
	cache, ok := fileCache[mediaId]
	fileCacheMutex.RUnlock()

	if ok && time.Since(cache.timestamp) < fileCacheTTL {
		w.Header().Set("Content-Type", cache.mimeType)
		if cache.disposition != "" {
			w.Header().Set("Content-Disposition", cache.disposition)
		} else {
			w.Header().Set("Content-Disposition", "attachment; filename=\"arquivo\"")
		}
		w.WriteHeader(http.StatusOK)
		w.Write(cache.data)
		return
	}

	apiKey := os.Getenv("D360_API_KEY")
	client := &http.Client{Timeout: 15 * time.Second}

	// 1. Busca metadata da mídia
	metaUrl := "https://waba-v2.360dialog.io/media/" + mediaId
	reqMeta, _ := http.NewRequest(http.MethodGet, metaUrl, nil)
	reqMeta.Header.Set("D360-API-KEY", apiKey)
	respMeta, err := client.Do(reqMeta)
	if err != nil || respMeta.StatusCode != http.StatusOK {
		log.Printf("[MediaDownload] Erro metadata: %v", err)
		http.Error(w, "Erro ao buscar metadata", http.StatusBadGateway)
		return
	}
	defer respMeta.Body.Close()

	var meta struct {
		URL      string `json:"url"`
		MimeType string `json:"mime_type"`
	}
	if err := json.NewDecoder(respMeta.Body).Decode(&meta); err != nil {
		log.Printf("[MediaDownload] Erro decode: %v", err)
		http.Error(w, "Erro na metadata", http.StatusBadGateway)
		return
	}

	// 2. Troca o host para waba-v2.360dialog.io, conforme doc
	downloadURL := strings.Replace(meta.URL, "https://lookaside.fbsbx.com", "https://waba-v2.360dialog.io", 1)

	// 3. Faz GET na URL trocada, COM o header D360-API-KEY
	reqFile, _ := http.NewRequest(http.MethodGet, downloadURL, nil)
	reqFile.Header.Set("D360-API-KEY", apiKey)
	respFile, err := client.Do(reqFile)
	if err != nil || respFile.StatusCode != http.StatusOK {
		log.Printf("[MediaDownload] Erro arquivo: %v", err)
		http.Error(w, "Erro ao baixar arquivo", http.StatusBadGateway)
		return
	}
	defer respFile.Body.Close()

	data, err := io.ReadAll(respFile.Body)
	if err != nil {
		log.Printf("[MediaDownload] Erro lendo arquivo: %v", err)
		http.Error(w, "Erro ao ler arquivo", http.StatusBadGateway)
		return
	}

	mimeType := meta.MimeType
	disposition := respFile.Header.Get("Content-Disposition")
	if disposition == "" {
		disposition = "attachment; filename=\"arquivo\""
	}

	// 6. Salva no cache
	fileCacheMutex.Lock()
	fileCache[mediaId] = fileCacheItem{
		data:        data,
		mimeType:    mimeType,
		disposition: disposition,
		timestamp:   time.Now(),
	}
	fileCacheMutex.Unlock()

	// 4. Copia headers relevantes e faz streaming do arquivo
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", disposition)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
