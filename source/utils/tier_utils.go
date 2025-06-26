package utils

import (
	"api/schemas"
	"encoding/json"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductInLegacyList struct {
	Preco      float64 `json:"preco"`
	Quantidade float64 `json:"quantidade"`
}

func CalculateLeadTier(relatedOrders bson.A, tiers []schemas.LeadTier) any {
	totalValue := 0.0
	maxValue := 0.0

	for _, orderDoc := range relatedOrders {
		orderMap, ok := orderDoc.(bson.M)
		if !ok {
			continue
		}
		if productsListLegacy, ok := orderMap["products_list_legacy"].(string); ok && productsListLegacy != "" {
			products := []ProductInLegacyList{}
			if err := json.Unmarshal([]byte(productsListLegacy), &products); err == nil {
				currentOrderValue := 0.0
				for _, p := range products {
					currentOrderValue += p.Preco * p.Quantidade
				}
				totalValue += currentOrderValue
				if currentOrderValue > maxValue {
					maxValue = currentOrderValue
				}
			}
		}
	}

	for _, tier := range tiers {
		var valueToCompare float64
		if tier.SumType == "total_value" {
			valueToCompare = totalValue
		} else if tier.SumType == "max_value" {
			valueToCompare = maxValue
		} else {
			continue
		}
		if valueToCompare >= tier.MinValue && valueToCompare <= tier.MaxValue {
			return tier
		}
	}
	return nil
}
