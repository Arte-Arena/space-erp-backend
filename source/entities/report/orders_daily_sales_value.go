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

func GetOrdersDailySalesValue(from, until string) (map[string]float64, error) {
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
				// Unsupported date format; skip this document
				continue
			}
		}

		listStr, _ := doc["products_list_legacy"].(string)
		var orderTotal float64

		if listStr != "" {
			var products []legacyProduct
			if err := json.Unmarshal([]byte(listStr), &products); err != nil {
				continue
			}
			for _, p := range products {
				qty := p.Quantidade
				if qty == 0 {
					qty = 1
				}
				orderTotal += p.Preco * qty
			}
		} else {
			if tinyMap, ok := doc["tiny"].(bson.M); ok {
				if valAny, ok2 := tinyMap["total_produtos"]; ok2 {
					switch v := valAny.(type) {
					case float64:
						orderTotal = v
					case int:
						orderTotal = float64(v)
					case int32:
						orderTotal = float64(v)
					case int64:
						orderTotal = float64(v)
					case string:
						if f, err := strconv.ParseFloat(v, 64); err == nil {
							orderTotal = f
						}
					}
				}
			}
		}

		if orderTotal == 0 {
			continue
		}

		key := fmt.Sprintf("%04d-%02d-%02d", createdAt.Year(), createdAt.Month(), createdAt.Day())
		result[key] += orderTotal
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
