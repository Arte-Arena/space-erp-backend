package orders

import (
	"api/source/schemas"
	"database/sql"
	"fmt"
	"os"
)

func GetOneOld(oldId int) (*schemas.OrderOld, error) {
	mysqlURI := os.Getenv("MYSQL_URI")

	mysqlDB, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	defer mysqlDB.Close()

	order := schemas.OrderOld{}
	err = mysqlDB.QueryRow("SELECT * FROM pedidos_arte_final WHERE id = ?", oldId).Scan(
		&order.ID,
		&order.UserID,
		&order.NumeroPedido,
		&order.PrazoArteFinal,
		&order.PrazoConfeccao,
		&order.ListaProdutos,
		&order.Observacoes,
		&order.Rolo,
		&order.PedidoStatusID,
		&order.PedidoTipoID,
		&order.Estagio,
		&order.URLTrello,
		&order.Situacao,
		&order.Prioridade,
		&order.OrcamentoID,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.TinyPedidoID,
		&order.DataPrevista,
		&order.VendedorID,
		&order.DesignerID,
		&order.CodigoRastreamento,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order from MySQL: %w", err)
	}

	return &order, nil
}
