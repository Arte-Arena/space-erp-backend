package report

import (
	"api/database"
	"context"
	"math"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetCommercialBudgetsTotalSalesValue(seller bson.ObjectID, from, until string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return 0, err
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
			{Key: "_id", Value: nil},
			{Key: "total_value", Value: bson.D{{Key: "$sum", Value: "$billing.installments.value"}}},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return 0, err
	}

	if len(result) > 0 {
		if total, ok := result[0]["total_value"].(float64); ok {
			total = math.Round(total*100) / 100
			return total, nil
		}
	}

	return 0, nil
}
