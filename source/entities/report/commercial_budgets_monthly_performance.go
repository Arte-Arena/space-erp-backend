package report

import (
	"api/database"
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetCommercialBudgetsMonthlyPerformance(seller bson.ObjectID, from, until string) (map[string]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)

	coll := client.Database(database.GetDB()).Collection(database.COLLECTION_BUDGETS)

	filter := bson.D{{Key: "seller", Value: seller}, {Key: "approved", Value: true}}
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

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$unwind", Value: "$billing.installments"}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "budget_id", Value: "$_id"},
				{Key: "year", Value: bson.D{{Key: "$year", Value: "$created_at"}}},
				{Key: "month", Value: bson.D{{Key: "$month", Value: "$created_at"}}},
			}},
			{Key: "budget_total", Value: bson.D{{Key: "$sum", Value: "$billing.installments.value"}}},
		}}},

		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "year", Value: "$_id.year"},
				{Key: "month", Value: "$_id.month"},
			}},
			{Key: "total_budgets", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "total_value", Value: bson.D{{Key: "$sum", Value: "$budget_total"}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id.year", Value: 1}, {Key: "_id.month", Value: 1}}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]map[string]any)

	castInt32 := func(v any) int32 {
		switch t := v.(type) {
		case int32:
			return t
		case int64:
			return int32(t)
		case float64:
			return int32(t)
		default:
			return 0
		}
	}

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		idVal := doc["_id"]
		var year32, month32 int32
		switch id := idVal.(type) {
		case bson.M:
			year32 = castInt32(id["year"])
			month32 = castInt32(id["month"])
		case bson.D:
			for _, elem := range id {
				if elem.Key == "year" {
					year32 = castInt32(elem.Value)
				}
				if elem.Key == "month" {
					month32 = castInt32(elem.Value)
				}
			}
		}

		key := fmt.Sprintf("%04d-%02d", year32, month32)

		totalBudgets, _ := doc["total_budgets"].(int32)
		var totalValue float64
		switch tv := doc["total_value"].(type) {
		case int32:
			totalValue = float64(tv)
		case int64:
			totalValue = float64(tv)
		case float64:
			totalValue = tv
		}

		result[key] = map[string]any{
			"total_budgets": int64(totalBudgets),
			"total_value":   math.Round(totalValue*100) / 100,
		}
	}

	return result, nil
}
