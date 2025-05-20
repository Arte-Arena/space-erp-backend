package budgets

import (
	"api/source/schemas"
	"database/sql"
	"fmt"
	"os"
	"strings"
)

func GetManyOld(oldIds []int) ([]*schemas.BudgetOld, error) {
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

	query := fmt.Sprintf("SELECT * FROM orcamentos WHERE id IN (%s)", strings.Join(placeholders, ","))

	rows, err := mysqlDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query budgets from MySQL: %w", err)
	}
	defer rows.Close()

	budgets := []*schemas.BudgetOld{}
	for rows.Next() {
		budget := &schemas.BudgetOld{}
		err := rows.Scan(
			&budget.ID,
			&budget.UserID,
			&budget.ClienteOctaNumber,
			&budget.NomeCliente,
			&budget.ListaProdutos,
			&budget.TextoOrcamento,
			&budget.EnderecoCep,
			&budget.Endereco,
			&budget.OpcaoEntrega,
			&budget.PrazoOpcaoEntrega,
			&budget.PrecoOpcaoEntrega,
			&budget.CreatedAt,
			&budget.UpdatedAt,
			&budget.Antecipado,
			&budget.DataAntecipa,
			&budget.TaxaAntecipa,
			&budget.Descontado,
			&budget.TipoDesconto,
			&budget.ValorDesconto,
			&budget.PercentualDesconto,
			&budget.TotalOrcamento,
			&budget.Brinde,
			&budget.ProdutosBrinde,
			&budget.PrazoProducao,
			&budget.PrevEntrega,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan budget row: %w", err)
		}
		budgets = append(budgets, budget)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating budget rows: %w", err)
	}

	return budgets, nil
}
