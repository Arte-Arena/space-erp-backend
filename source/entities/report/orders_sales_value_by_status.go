package report

import (
	"api/database"
	"context"
	"encoding/json"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetOrdersSalesValueByStatus(from, until string) (map[string]float64, error) {
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

	findOpts := options.Find().SetProjection(bson.M{
		"status":               1,
		"products_list_legacy": 1,
	})

	cursor, err := collection.Find(ctx, filter, findOpts)
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

		status, _ := doc["status"].(string)
		if status == "" {
			status = "Sem status"
		}

		listStr, _ := doc["products_list_legacy"].(string)
		if listStr == "" {
			continue
		}

		var products []legacyProduct
		if err := json.Unmarshal([]byte(listStr), &products); err != nil {
			continue
		}

		var orderTotal float64
		for _, p := range products {
			qty := p.Quantidade
			if qty == 0 {
				qty = 1
			}
			orderTotal += p.Preco * qty
		}

		result[status] += orderTotal
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
