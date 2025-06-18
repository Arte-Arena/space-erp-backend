package spacedesk

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(responseBody)
		return
	}

	// 4. Repassa a resposta da 360dialog ao cliente
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(responseBody)
}
