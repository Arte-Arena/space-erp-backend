package orders

import (
	"api/source/schemas"
	"database/sql"
	"fmt"
	"os"
	"strings"
)

func GetManyOld(oldIds []int) ([]*schemas.OrderOld, error) {
	if len(oldIds) == 0 {
		return nil, nil
	}

	mysqlURI := os.Getenv("MYSQL_URI")

	mysqlDB, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	defer mysqlDB.Close()

	placeholders := make([]string, len(oldIds))
	args := make([]any, len(oldIds))
	for i, id := range oldIds {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT * FROM pedidos_arte_final WHERE id IN (%s)", strings.Join(placeholders, ","))

	rows, err := mysqlDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders from MySQL: %w", err)
	}
	defer rows.Close()

	orders := []*schemas.OrderOld{}
	for rows.Next() {
		order := &schemas.OrderOld{}
		err := rows.Scan(
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
		if err != nil {
			return nil, fmt.Errorf("failed to scan order row: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order rows: %w", err)
	}

	return orders, nil
}
