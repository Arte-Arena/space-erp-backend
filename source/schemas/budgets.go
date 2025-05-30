package schemas

import (
	"database/sql"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Budget struct {
	ID    bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	OldID uint64        `json:"old_id" bson:"old_id"`
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
