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

func GetBudgetsAverageTicket(from, until string, notApproved bool) (float64, error) {
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

	var filter bson.D
	if notApproved {
		filter = bson.D{{Key: "approved", Value: false}}
	} else {
		filter = bson.D{{Key: "approved", Value: true}}
	}

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

	if notApproved {
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
		count := 0
		for cursor.Next(ctx) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				continue
			}

			oldProductsList, _ := doc["old_products_list"].(string)
			var products []Product
			budgetTotal := 0.0
			if oldProductsList != "" {
				_ = json.Unmarshal([]byte(oldProductsList), &products)
				for _, p := range products {
					budgetTotal += p.Preco * float64(p.Quantidade)
				}
			}

			delivery, _ := doc["delivery"].(bson.M)
			if delivery != nil {
				if price, ok := delivery["price"].(float64); ok {
					budgetTotal += price
				}
			}
			if budgetTotal > 0 {
				total += budgetTotal
				count++
			}
		}
		if err := cursor.Err(); err != nil {
			return 0, err
		}
		if count == 0 {
			return 0, nil
		}
		return total / float64(count), nil
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$unwind", Value: "$billing.installments"}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$_id"},
			{Key: "budget_total", Value: bson.D{{Key: "$sum", Value: "$billing.installments.value"}}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "avg_value", Value: bson.D{{Key: "$avg", Value: "$budget_total"}}},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return 0, err
	}

	if len(result) > 0 {
		if avg, ok := result[0]["avg_value"].(float64); ok {
			return avg, nil
		}
	}

	return 0, nil
}
