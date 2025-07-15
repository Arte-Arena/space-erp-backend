package spacedesk

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func CreateOrderDetailsTemplate(apiKey string) error {
	payload := map[string]interface{}{
		"name":           "order_details_pix_2",
		"language":       "pt_BR",
		"category":       "UTILITY",
		"display_format": "ORDER_DETAILS",
		"components": []interface{}{
			map[string]interface{}{
				"type":   "HEADER",
				"format": "TEXT",
				"text":   "Teste pagamento pix",
			},
			map[string]interface{}{
				"type": "BODY",
				"text": "Obrigado pela sua compra. Segue abaixo o codigo",
			},
			map[string]interface{}{
				"type": "BUTTONS",
				"buttons": []interface{}{
					map[string]interface{}{
						"type": "ORDER_DETAILS",
						"text": "Copy Pix code",
					},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(context.Background(), "POST",
		"https://waba-v2.360dialog.io/message_templates",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("D360-API-KEY", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("Falha ao criar template: %s", resp.Status)
	}
	return nil
}
