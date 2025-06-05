package budgets

import (
	"api/utils"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

const SHIPPING_SERVICE_CODE = "03220"
const RECIPIENT_COUNTRY = "BR"

type ShippingQuoteItem struct {
	Height   float64 `json:"Height"`
	Length   float64 `json:"Length"`
	Quantity int     `json:"Quantity"`
	Weight   float64 `json:"Weight"`
	Width    float64 `json:"Width"`
}

type ShippingQuoteRequest struct {
	SellerCEP            string              `json:"SellerCEP"`
	RecipientCEP         string              `json:"RecipientCEP"`
	ShipmentInvoiceValue float64             `json:"ShipmentInvoiceValue"`
	ShippingServiceCode  string              `json:"ShippingServiceCode"`
	ShippingItemArray    []ShippingQuoteItem `json:"ShippingItemArray"`
	RecipientCountry     string              `json:"RecipientCountry"`
}

type ShippingQuoteService struct {
	Carrier               string `json:"Carrier"`
	CarrierCode           string `json:"CarrierCode"`
	DeliveryTime          string `json:"DeliveryTime"`
	Msg                   string `json:"Msg"`
	ServiceCode           string `json:"ServiceCode"`
	ServiceDescription    string `json:"ServiceDescription"`
	ShippingPrice         string `json:"ShippingPrice"`
	OriginalDeliveryTime  string `json:"OriginalDeliveryTime"`
	OriginalShippingPrice string `json:"OriginalShippingPrice"`
	Error                 bool   `json:"Error"`
}

type ShippingQuoteResponse struct {
	ShippingSevicesArray []ShippingQuoteService `json:"ShippingSevicesArray"`
	Timeout              int                    `json:"Timeout"`
}

type ShippingQuoteInput struct {
	SellerCEP            string  `json:"seller_cep"`
	RecipientCEP         string  `json:"recipient_cep"`
	ShipmentInvoiceValue float64 `json:"shipment_invoice_value"`
	Height               float64 `json:"height"`
	Length               float64 `json:"length"`
	Weight               float64 `json:"weight"`
	Width                float64 `json:"width"`
}

type SedexData struct {
	Price              string `json:"price"`
	DeliveryTime       string `json:"delivery_time"`
	Carrier            string `json:"carrier"`
	ServiceDescription string `json:"service_description"`
}

func CreateShippingQuote(w http.ResponseWriter, r *http.Request) {
	service := r.PathValue("service")

	shippingServiceCode := ""
	switch service {
	case "sedex":
		shippingServiceCode = "03220"
	case "pac":
		shippingServiceCode = "03298"
	default:
		utils.SendResponse(w, http.StatusBadRequest, "Serviço inválido. Use 'sedex' ou 'pac'", nil, 0)
		return
	}

	input := ShippingQuoteInput{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "Dados inválidos", nil, 0)
		return
	}

	if input.SellerCEP == "" || input.RecipientCEP == "" || input.ShipmentInvoiceValue == 0 || input.Height == 0 || input.Length == 0 || input.Weight == 0 || input.Width == 0 {
		utils.SendResponse(w, http.StatusBadRequest, "Todos os campos são obrigatórios e devem ser preenchidos corretamente", nil, 0)
		return
	}

	req := ShippingQuoteRequest{
		SellerCEP:            input.SellerCEP,
		RecipientCEP:         input.RecipientCEP,
		ShipmentInvoiceValue: input.ShipmentInvoiceValue,
		ShippingServiceCode:  shippingServiceCode,
		ShippingItemArray: []ShippingQuoteItem{
			{
				Height:   input.Height,
				Length:   input.Length,
				Quantity: 1,
				Weight:   input.Weight,
				Width:    input.Width,
			},
		},
		RecipientCountry: RECIPIENT_COUNTRY,
	}

	frenetToken := os.Getenv("FRENET_API_KEY")
	if frenetToken == "" {
		utils.SendResponse(w, http.StatusInternalServerError, "Token Frenet não configurado", nil, 0)
		return
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao serializar request", nil, 0)
		return
	}

	httpReq, err := http.NewRequest("POST", "https://api.frenet.com.br/shipping/quote", io.NopCloser(bytes.NewReader(jsonBody)))
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao criar request para Frenet", nil, 0)
		return
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("token", frenetToken)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "Erro ao conectar na Frenet", nil, 0)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.SendResponse(w, http.StatusBadGateway, "Erro da Frenet", nil, 0)
		return
	}

	frenetResp := ShippingQuoteResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&frenetResp); err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "Erro ao ler resposta da Frenet", nil, 0)
		return
	}

	var sedexData *SedexData = nil
	for _, svc := range frenetResp.ShippingSevicesArray {
		if svc.ServiceCode == shippingServiceCode {
			sedexData = &SedexData{
				Price:              svc.ShippingPrice,
				DeliveryTime:       svc.DeliveryTime,
				Carrier:            svc.Carrier,
				ServiceDescription: svc.ServiceDescription,
			}
			break
		}
	}

	if sedexData == nil {
		utils.SendResponse(w, http.StatusNotFound, "Serviço não encontrado na resposta da Frenet", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", sedexData, 0)
}
