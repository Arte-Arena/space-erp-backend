package schemas

import (
	"database/sql"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Order struct {
	ID    bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	OldID uint64        `json:"old_id" bson:"old_id"`
}

type OrderOld struct {
	ID                 uint64         `db:"id"`
	UserID             sql.NullInt64  `db:"user_id"`
	NumeroPedido       sql.NullString `db:"numero_pedido"`
	PrazoArteFinal     sql.NullTime   `db:"prazo_arte_final"`
	PrazoConfeccao     sql.NullTime   `db:"prazo_confeccao"`
	ListaProdutos      sql.NullString `db:"lista_produtos"`
	Observacoes        sql.NullString `db:"observacoes"`
	Rolo               sql.NullString `db:"rolo"`
	PedidoStatusID     sql.NullInt64  `db:"pedido_status_id"`
	PedidoTipoID       sql.NullInt64  `db:"pedido_tipo_id"`
	Estagio            sql.NullString `db:"estagio"`
	URLTrello          sql.NullString `db:"url_trello"`
	Situacao           sql.NullString `db:"situacao"`
	Prioridade         sql.NullString `db:"prioridade"`
	OrcamentoID        sql.NullInt64  `db:"orcamento_id"`
	CreatedAt          sql.NullTime   `db:"created_at"`
	UpdatedAt          sql.NullTime   `db:"updated_at"`
	TinyPedidoID       sql.NullString `db:"tiny_pedido_id"`
	DataPrevista       sql.NullTime   `db:"data_prevista"`
	VendedorID         sql.NullInt64  `db:"vendedor_id"`
	DesignerID         sql.NullInt64  `db:"designer_id"`
	CodigoRastreamento sql.NullString `db:"codigo_rastreamento"`
}
