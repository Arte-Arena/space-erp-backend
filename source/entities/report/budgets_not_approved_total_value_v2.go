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

func GetBudgetsNotApprovedTotalValueV2(from, until string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		return 0, err
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_BUDGETS)

	filter := bson.D{{Key: "approved", Value: false}}

	if from != "" || until != "" {
		dateFilter := bson.D{}
		if from != "" {
			if fromTime, err := time.Parse("2006-01-02", from); err == nil {
				fromTime = time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(), 0, 0, 0, 0, fromTime.Location())
				dateFilter = append(dateFilter, bson.E{Key: "$gte", Value: fromTime})
			}
		}
		if until != "" {
			if untilTime, err := time.Parse("2006-01-02", until); err == nil {
				untilTime = time.Date(untilTime.Year(), untilTime.Month(), untilTime.Day(), 23, 59, 59, 999999999, untilTime.Location())
				dateFilter = append(dateFilter, bson.E{Key: "$lte", Value: untilTime})
			}
		}
		if len(dateFilter) > 0 {
			filter = append(filter, bson.E{Key: "created_at", Value: dateFilter})
		}
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	type Product struct {
		Preco      float64 `json:"preco"`
		Quantidade int     `json:"quantidade"`
	}

	total := 0.0
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		oldProductsList, _ := doc["old_products_list"].(string)
		var products []Product
		if oldProductsList != "" {
			_ = json.Unmarshal([]byte(oldProductsList), &products)
			for _, p := range products {
				total += p.Preco * float64(p.Quantidade)
			}
		}

		delivery, _ := doc["delivery"].(bson.M)
		if delivery != nil {
			if price, ok := delivery["price"].(float64); ok {
				total += price
			}
		}
	}

	if err := cursor.Err(); err != nil {
		return 0, err
	}

	return total, nil
}
