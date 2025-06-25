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

	Cliente             *TinyCliente   `json:"cliente,omitempty" bson:"cliente,omitempty"`
	CodigoRastreamento  string         `json:"codigo_rastreamento,omitempty" bson:"codigo_rastreamento,omitempty"`
	CondicaoPagamento   string         `json:"condicao_pagamento,omitempty" bson:"condicao_pagamento,omitempty"`
	DataEntrega         string         `json:"data_entrega,omitempty" bson:"data_entrega,omitempty"`
	DataEnvio           string         `json:"data_envio,omitempty" bson:"data_envio,omitempty"`
	DataFaturamento     string         `json:"data_faturamento,omitempty" bson:"data_faturamento,omitempty"`
	DataPedido          string         `json:"data_pedido,omitempty" bson:"data_pedido,omitempty"`
	DataPrevista        string         `json:"data_prevista,omitempty" bson:"data_prevista,omitempty"`
	Deposito            string         `json:"deposito,omitempty" bson:"deposito,omitempty"`
	DescricaoListaPreco string         `json:"descricao_lista_preco,omitempty" bson:"descricao_lista_preco,omitempty"`
	FormaEnvio          string         `json:"forma_envio,omitempty" bson:"forma_envio,omitempty"`
	FormaFrete          string         `json:"forma_frete,omitempty" bson:"forma_frete,omitempty"`
	FormaPagamento      string         `json:"forma_pagamento,omitempty" bson:"forma_pagamento,omitempty"`
	FretePorConta       string         `json:"frete_por_conta,omitempty" bson:"frete_por_conta,omitempty"`
	IdListaPreco        string         `json:"id_lista_preco,omitempty" bson:"id_lista_preco,omitempty"`
	IdNaturezaOperacao  string         `json:"id_natureza_operacao,omitempty" bson:"id_natureza_operacao,omitempty"`
	IdNotaFiscal        string         `json:"id_nota_fiscal,omitempty" bson:"id_nota_fiscal,omitempty"`
	IdVendedor          string         `json:"id_vendedor,omitempty" bson:"id_vendedor,omitempty"`
	Itens               []TinyItem     `json:"itens,omitempty" bson:"itens,omitempty"`
	Marcadores          []TinyMarcador `json:"marcadores,omitempty" bson:"marcadores,omitempty"`
	MeioPagamento       string         `json:"meio_pagamento,omitempty" bson:"meio_pagamento,omitempty"`
	NomeTransportador   string         `json:"nome_transportador,omitempty" bson:"nome_transportador,omitempty"`
	NomeVendedor        string         `json:"nome_vendedor,omitempty" bson:"nome_vendedor,omitempty"`
	NumeroEcommerce     *string        `json:"numero_ecommerce,omitempty" bson:"numero_ecommerce,omitempty"`
	NumeroOrdemCompra   string         `json:"numero_ordem_compra,omitempty" bson:"numero_ordem_compra,omitempty"`
	Obs                 string         `json:"obs,omitempty" bson:"obs,omitempty"`
	ObsInterna          string         `json:"obs_interna,omitempty" bson:"obs_interna,omitempty"`
	OutrasDespesas      string         `json:"outras_despesas,omitempty" bson:"outras_despesas,omitempty"`
	Parcelas            []TinyParcela  `json:"parcelas,omitempty" bson:"parcelas,omitempty"`
	Situacao            string         `json:"situacao,omitempty" bson:"situacao,omitempty"`
	TotalPedido         string         `json:"total_pedido,omitempty" bson:"total_pedido,omitempty"`
	TotalProdutos       string         `json:"total_produtos,omitempty" bson:"total_produtos,omitempty"`
	UrlRastreamento     string         `json:"url_rastreamento,omitempty" bson:"url_rastreamento,omitempty"`
	ValorDesconto       any            `json:"valor_desconto,omitempty" bson:"valor_desconto,omitempty"`
	ValorFrete          string         `json:"valor_frete,omitempty" bson:"valor_frete,omitempty"`
	Nome                string         `json:"nome,omitempty" bson:"nome,omitempty"`
	Valor               float64        `json:"valor,omitempty" bson:"valor,omitempty"`
}

type TinyCliente struct {
	Bairro       string `json:"bairro,omitempty" bson:"bairro,omitempty"`
	Cep          string `json:"cep,omitempty" bson:"cep,omitempty"`
	Cidade       string `json:"cidade,omitempty" bson:"cidade,omitempty"`
	Codigo       string `json:"codigo,omitempty" bson:"codigo,omitempty"`
	Complemento  string `json:"complemento,omitempty" bson:"complemento,omitempty"`
	CpfCnpj      string `json:"cpf_cnpj,omitempty" bson:"cpf_cnpj,omitempty"`
	Email        string `json:"email,omitempty" bson:"email,omitempty"`
	Endereco     string `json:"endereco,omitempty" bson:"endereco,omitempty"`
	Fone         string `json:"fone,omitempty" bson:"fone,omitempty"`
	Ie           string `json:"ie,omitempty" bson:"ie,omitempty"`
	Nome         string `json:"nome,omitempty" bson:"nome,omitempty"`
	NomeFantasia string `json:"nome_fantasia,omitempty" bson:"nome_fantasia,omitempty"`
	Numero       string `json:"numero,omitempty" bson:"numero,omitempty"`
	Rg           string `json:"rg,omitempty" bson:"rg,omitempty"`
	TipoPessoa   string `json:"tipo_pessoa,omitempty" bson:"tipo_pessoa,omitempty"`
	Uf           string `json:"uf,omitempty" bson:"uf,omitempty"`
}

type TinyItem struct {
	Item TinyItemDetail `json:"item" bson:"item"`
}

type TinyItemDetail struct {
	Codigo        string `json:"codigo,omitempty" bson:"codigo,omitempty"`
	Descricao     string `json:"descricao,omitempty" bson:"descricao,omitempty"`
	IdProduto     string `json:"id_produto,omitempty" bson:"id_produto,omitempty"`
	Quantidade    string `json:"quantidade,omitempty" bson:"quantidade,omitempty"`
	Unidade       string `json:"unidade,omitempty" bson:"unidade,omitempty"`
	ValorUnitario string `json:"valor_unitario,omitempty" bson:"valor_unitario,omitempty"`
}

type TinyMarcador struct {
	Marcador TinyMarcadorDetail `json:"marcador" bson:"marcador"`
}

type TinyMarcadorDetail struct {
	Cor       string `json:"cor,omitempty" bson:"cor,omitempty"`
	Descricao string `json:"descricao,omitempty" bson:"descricao,omitempty"`
	ID        string `json:"id,omitempty" bson:"id,omitempty"`
}

type TinyParcela struct {
	Parcela TinyParcelaDetail `json:"parcela" bson:"parcela"`
}

type TinyParcelaDetail struct {
	Data           string `json:"data,omitempty" bson:"data,omitempty"`
	Dias           string `json:"dias,omitempty" bson:"dias,omitempty"`
	FormaPagamento string `json:"forma_pagamento,omitempty" bson:"forma_pagamento,omitempty"`
	MeioPagamento  string `json:"meio_pagamento,omitempty" bson:"meio_pagamento,omitempty"`
	Obs            string `json:"obs,omitempty" bson:"obs,omitempty"`
	Valor          string `json:"valor,omitempty" bson:"valor,omitempty"`
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
