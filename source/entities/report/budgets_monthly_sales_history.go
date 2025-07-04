package report

import (
	"api/database"
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetBudgetsMonthlySalesHistory(from, until string) (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_BUDGETS)

	filter := bson.D{{Key: "approved", Value: true}}
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
				{Key: "year", Value: bson.D{{Key: "$year", Value: "$created_at"}}},
				{Key: "month", Value: bson.D{{Key: "$month", Value: "$created_at"}}},
			}},
			{Key: "total_value", Value: bson.D{{Key: "$sum", Value: "$billing.installments.value"}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id.year", Value: 1}, {Key: "_id.month", Value: 1}}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]float64)
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		idVal := doc["_id"]

		castToInt32 := func(v any) int32 {
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

		var year32, month32 int32
		switch id := idVal.(type) {
		case bson.M:
			year32 = castToInt32(id["year"])
			month32 = castToInt32(id["month"])
		case bson.D:
			for _, elem := range id {
				if elem.Key == "year" {
					year32 = castToInt32(elem.Value)
				}
				if elem.Key == "month" {
					month32 = castToInt32(elem.Value)
				}
			}
		}

		var totalFloat float64
		switch tv := doc["total_value"].(type) {
		case int32:
			totalFloat = float64(tv)
		case int64:
			totalFloat = float64(tv)
		case float64:
			totalFloat = tv
		}

		key := fmt.Sprintf("%04d-%02d", year32, month32)
		result[key] = totalFloat
	}
	return result, nil
}
