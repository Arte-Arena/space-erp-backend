package report

import (
	"api/database"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetOrdersMonthlySalesHistory(from, until string) (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_ORDERS)

	filter := bson.D{}
	if from != "" || until != "" {
		dateFilter := bson.D{}
		if from != "" {
			if fromTime, err := time.Parse(time.RFC3339, from); err == nil {
				dateFilter = append(dateFilter, bson.E{Key: "$gte", Value: fromTime})
			}
		}
		if until != "" {
			if untilTime, err := time.Parse(time.RFC3339, until); err == nil {
				dateFilter = append(dateFilter, bson.E{Key: "$lte", Value: untilTime})
			}
		}
		if len(dateFilter) > 0 {
			filter = append(filter, bson.E{Key: "created_at", Value: dateFilter})
		}
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type legacyProduct struct {
		Preco      float64 `json:"preco"`
		Quantidade float64 `json:"quantidade"`
	}

	result := make(map[string]float64)

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		var createdAt time.Time
		switch v := doc["created_at"].(type) {
		case time.Time:
			createdAt = v
		default:
			if getter, ok := v.(interface{ Time() time.Time }); ok {
				createdAt = getter.Time()
			} else if ms, ok := v.(int64); ok {
				createdAt = time.UnixMilli(ms)
			} else {
				continue
			}
		}

		var orderTotal float64

		if listStr, _ := doc["products_list_legacy"].(string); listStr != "" {
			var products []legacyProduct
			if err := json.Unmarshal([]byte(listStr), &products); err == nil {
				for _, p := range products {
					qty := p.Quantidade
					if qty == 0 {
						qty = 1
					}
					orderTotal += p.Preco * qty
				}
			}
		} else {
			getTotalProdutos := func(tiny interface{}) float64 {
				var valAny interface{}
				var found bool

				if tinyMap, ok := tiny.(bson.M); ok {
					valAny, found = tinyMap["total_produtos"]
				} else if tinyDoc, ok := tiny.(bson.D); ok {
					for _, elem := range tinyDoc {
						if elem.Key == "total_produtos" {
							valAny = elem.Value
							found = true
							break
						}
					}
				}

				if !found {
					return 0
				}

				switch v := valAny.(type) {
				case float64:
					return v
				case int:
					return float64(v)
				case int32:
					return float64(v)
				case int64:
					return float64(v)
				case string:
					if f, err := strconv.ParseFloat(v, 64); err == nil {
						return f
					}
				}
				return 0
			}

			if tiny, ok := doc["tiny"]; ok {
				orderTotal = getTotalProdutos(tiny)
			}
		}

		if orderTotal == 0 {
			continue
		}

		key := fmt.Sprintf("%04d-%02d", createdAt.Year(), createdAt.Month())
		result[key] += orderTotal
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
