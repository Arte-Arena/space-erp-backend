package spacedesk

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func DeleteD360Template(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método não permitido. Use DELETE.", http.StatusMethodNotAllowed)
		return
	}

	// 1. Extrair o nome do template da URL
	templateName := strings.TrimPrefix(r.URL.Path, "/v1/space-desk/template-messages/")
	if templateName == "" {
		http.Error(w, "Nome do template não fornecido na URL.", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("SPACE_DESK_API_KEY")
	if apiKey == "" {
		http.Error(w, "API Key da 360dialog não configurada.", http.StatusInternalServerError)
		return
	}

	dialogURL := fmt.Sprintf("https://waba-v2.360dialog.io/v1/configs/templates/%s", templateName)
	req, err := http.NewRequest("DELETE", dialogURL, nil)
	if err != nil {
		http.Error(w, "Erro ao criar request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("D360-API-KEY", apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Erro ao chamar API da 360: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Template '" + templateName + "' excluído com sucesso."})
	} else {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Erro ao excluir template '%s': Status %d, Resposta: %s", templateName, resp.StatusCode, string(bodyBytes))

		errorMsg := fmt.Sprintf("Falha ao excluir template. Status: %d. Resposta da 360dialog: %s", resp.StatusCode, string(bodyBytes))
		http.Error(w, errorMsg, resp.StatusCode) // Repassa o status de erro da 360dialog
	}
}
