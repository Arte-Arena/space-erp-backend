package spacedesk

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tplResp)
}
