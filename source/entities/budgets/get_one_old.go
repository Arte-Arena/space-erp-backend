package budgets

import (
	"api/schemas"
	"database/sql"
	"fmt"
	"os"
)

func GetOneOld(oldId int) (*schemas.BudgetOld, error) {
	mysqlURI := os.Getenv("MYSQL_URI")

	mysqlDB, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	defer mysqlDB.Close()

	budget := schemas.BudgetOld{}
	err = mysqlDB.QueryRow("SELECT * FROM orcamentos WHERE id = ?", oldId).Scan(
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

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch budget from MySQL: %w", err)
	}

	return &budget, nil
}
