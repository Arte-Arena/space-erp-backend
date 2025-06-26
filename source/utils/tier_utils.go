package utils

import (
	"api/schemas"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductInLegacyList struct {
	Preco      float64 `json:"preco"`
	Quantidade float64 `json:"quantidade"`
}

func CalculateLeadTier(relatedOrders bson.A, tiers []schemas.LeadTier) (any, error) {
	totalValue := 0.0
	maxValue := 0.0

	for i, orderDoc := range relatedOrders {
		var orderMap map[string]any
		switch v := orderDoc.(type) {
		case bson.M:
			orderMap = v
		case map[string]any:
			orderMap = v
		case bson.D:
			orderMap = make(map[string]any)
			for _, elem := range v {
				orderMap[elem.Key] = elem.Value
			}
		default:
			return nil, fmt.Errorf("orderDoc at index %d is not a map", i)
		}
		productsListLegacy, ok := orderMap["products_list_legacy"].(string)
		if !ok || productsListLegacy == "" {
			return nil, fmt.Errorf("order at index %d missing or empty products_list_legacy", i)
		}
		products := []ProductInLegacyList{}
		if err := json.Unmarshal([]byte(productsListLegacy), &products); err != nil {
			return nil, fmt.Errorf("failed to unmarshal products_list_legacy at index %d: %w", i, err)
		}
		currentOrderValue := 0.0
		for _, p := range products {
			currentOrderValue += p.Preco * p.Quantidade
		}
		totalValue += currentOrderValue
		if currentOrderValue > maxValue {
			maxValue = currentOrderValue
		}
	}

	for _, tier := range tiers {
		var valueToCompare float64
		switch tier.SumType {
		case "total":
			valueToCompare = totalValue
		case "individual":
			valueToCompare = maxValue
		default:
			return nil, fmt.Errorf("invalid SumType in tier: %s", tier.SumType)
		}
		if valueToCompare >= tier.MinValue && valueToCompare <= tier.MaxValue {
			return tier, nil
		}
	}
	return nil, fmt.Errorf("no matching tier found for the calculated values")
}
