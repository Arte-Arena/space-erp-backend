package schemas

import (
	"database/sql"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Delivery struct {
	Option   string  `json:"option" bson:"option"`
	Deadline uint    `json:"deadline" bson:"deadline"`
	Price    float64 `json:"price" bson:"price"`
}

type EarlyMode struct {
	Date time.Time `json:"date" bson:"date"`
	Tax  float64   `json:"tax" bson:"tax"`
}

type Discount struct {
	Type       string  `json:"type" bson:"type"`
	Value      float64 `json:"value" bson:"value"`
	Percentage float64 `json:"percentage" bson:"percentage"`
}

type Installments struct {
	Date  time.Time `json:"date" bson:"date"`
	Value float64   `json:"value" bson:"value"`
}

type Billing struct {
	Type         string         `json:"type" bson:"type"`
	Installments []Installments `json:"installments" bson:"installments"`
}

type Address struct {
	CEP     string `json:"cep" bson:"cep"`
	Details string `json:"details" bson:"details"`
}

type Budget struct {
	ID                 bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	OldID              uint64        `json:"old_id" bson:"old_id"`
	CreatedBy          bson.ObjectID `json:"created_by" bson:"created_by"`
	Seller             bson.ObjectID `json:"seller" bson:"seller"`
	RelatedLead        bson.ObjectID `json:"related_lead" bson:"related_lead"`
	RelatedClient      bson.ObjectID `json:"related_client" bson:"related_client"`
	OldProductsList    string        `json:"old_products_list" bson:"old_products_list"`
	Address            Address       `json:"address" bson:"address"`
	Delivery           Delivery      `json:"delivery" bson:"delivery"`
	EarlyMode          EarlyMode     `json:"early_mode" bson:"early_mode"`
	Discount           Discount      `json:"discount" bson:"discount"`
	OldGifts           string        `json:"old_gifts" bson:"old_gifts"`
	ProductionDeadline uint          `json:"production_deadline" bson:"production_deadline"`
	Status             string        `json:"status" bson:"status"`
	PaymentMethod      string        `json:"payment_method" bson:"payment_method"`
	Billing            Billing       `json:"billing" bson:"billing"`
	Trello_uri         string        `json:"trello_uri" bson:"trello_uri"`
	Notes              string        `json:"notes" bson:"notes"`
	DeliveryForecast   time.Time     `json:"delivery_forecast" bson:"delivery_forecast"`
	CreatedAt          time.Time     `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt          time.Time     `json:"updated_at" bson:"updated_at,omitempty"`
	Approved           bool          `json:"approved" bson:"approved"`
}

type BudgetOld struct {
	ID                 uint64          `db:"id"`
	UserID             uint64          `db:"user_id"`
	ClienteOctaNumber  sql.NullString  `db:"cliente_octa_number"`
	NomeCliente        string          `db:"nome_cliente"`
	ListaProdutos      string          `db:"lista_produtos"`
	TextoOrcamento     string          `db:"texto_orcamento"`
	EnderecoCep        sql.NullString  `db:"endereco_cep"`
	Endereco           sql.NullString  `db:"endereco"`
	OpcaoEntrega       sql.NullString  `db:"opcao_entrega"`
	PrazoOpcaoEntrega  sql.NullInt32   `db:"prazo_opcao_entrega"`
	PrecoOpcaoEntrega  sql.NullFloat64 `db:"preco_opcao_entrega"`
	CreatedAt          sql.NullString  `db:"created_at"`
	UpdatedAt          sql.NullString  `db:"updated_at"`
	Antecipado         sql.NullInt32   `db:"antecipado"`
	DataAntecipa       sql.NullTime    `db:"data_antecipa"`
	TaxaAntecipa       sql.NullFloat64 `db:"taxa_antecipa"`
	Descontado         sql.NullInt32   `db:"descontado"`
	TipoDesconto       sql.NullString  `db:"tipo_desconto"`
	ValorDesconto      sql.NullFloat64 `db:"valor_desconto"`
	PercentualDesconto sql.NullFloat64 `db:"percentual_desconto"`
	TotalOrcamento     sql.NullFloat64 `db:"total_orcamento"`
	Brinde             sql.NullInt32   `db:"brinde"`
	ProdutosBrinde     sql.NullString  `db:"produtos_brinde"`
	PrazoProducao      int             `db:"prazo_producao"`
	PrevEntrega        sql.NullString  `db:"prev_entrega"`
}
