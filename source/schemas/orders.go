package schemas

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderStatus string

const (
	StatusPendente     OrderStatus = "Pendente"
	StatusEmAndamento  OrderStatus = "Em andamento"
	StatusArteOK       OrderStatus = "Arte OK"
	StatusEmEspera     OrderStatus = "Em espera"
	StatusCorTeste     OrderStatus = "Cor teste"
	StatusProcessando  OrderStatus = "Processando"
	StatusImpresso     OrderStatus = "Impresso"
	StatusEmImpressao  OrderStatus = "Em impressão"
	StatusSeparacao    OrderStatus = "Separação"
	StatusCosturado    OrderStatus = "Costurado"
	StatusPrensa       OrderStatus = "Prensa"
	StatusCalandra     OrderStatus = "Calandra"
	StatusEmSeparacao  OrderStatus = "Em separação"
	StatusRetirada     OrderStatus = "Retirada"
	StatusEmEntrega    OrderStatus = "Em entrega"
	StatusEntregue     OrderStatus = "Entregue"
	StatusDevolucao    OrderStatus = "Devolução"
	StatusNaoCortado   OrderStatus = "Não cortado"
	StatusCortado      OrderStatus = "Cortado"
	StatusNaoConferido OrderStatus = "Não conferido"
	StatusConferido    OrderStatus = "Conferido"
)

type OrderStage string

const (
	StageDesign      OrderStage = "Design"
	StageImpressao   OrderStage = "Impressão"
	StageSublimacao  OrderStage = "Sublimação"
	StageCostura     OrderStage = "Costura"
	StageExpedicao   OrderStage = "Expedição"
	StageCorte       OrderStage = "Corte"
	StageConferencia OrderStage = "Conferência"
)

type OrderType string

const (
	TypePrazoNormal  OrderType = "Prazo normal"
	TypeAntecipacao  OrderType = "Antecipação"
	TypeFaturado     OrderType = "Faturado"
	TypeMetadeMetade OrderType = "Metade/Metade"
	TypeAmostra      OrderType = "Amostra"
	TypeReposicao    OrderType = "Reposição"
)

type TinyOrder struct {
	ID     string `json:"id,omitempty" bson:"id,omitempty"`
	Number string `json:"number,omitempty" bson:"number,omitempty"`
}

type Order struct {
	ID                 bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	OldID              uint64        `json:"old_id" bson:"old_id"`
	CreatedBy          bson.ObjectID `json:"created_by,omitempty" bson:"created_by,omitempty"`
	RelatedSeller      bson.ObjectID `json:"related_seller,omitempty" bson:"related_seller,omitempty"`
	RelatedDesigner    bson.ObjectID `json:"related_designer,omitempty" bson:"related_designer,omitempty"`
	TrackingCode       string        `json:"tracking_code,omitempty" bson:"tracking_code,omitempty"`
	Status             OrderStatus   `json:"status,omitempty" bson:"status,omitempty"`
	Stage              OrderStage    `json:"stage,omitempty" bson:"stage,omitempty"`
	Type               OrderType     `json:"type,omitempty" bson:"type,omitempty"`
	UrlTrello          string        `json:"url_trello,omitempty" bson:"url_trello,omitempty"`
	ProductsListLegacy string        `json:"products_list_legacy,omitempty" bson:"products_list_legacy,omitempty"`
	RelatedBudget      bson.ObjectID `json:"related_budget,omitempty" bson:"related_budget,omitempty"`
	ExpectedDate       time.Time     `json:"expected_date,omitempty" bson:"expected_date,omitempty"`
	CustomProperties   any           `json:"custom_properties,omitempty" bson:"custom_properties,omitempty"`
	Tiny               TinyOrder     `json:"tiny,omitempty" bson:"tiny,omitempty"`
	Notes              string        `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt          time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at" bson:"updated_at"`
}
