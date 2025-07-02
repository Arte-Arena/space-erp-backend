package report

import (
	"api/database"
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetOrdersMonthlyAverage(from, until string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		return 0, err
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

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	var months float64 = 1
	if from != "" && until != "" {
		fromTime, err1 := time.Parse(time.RFC3339, from)
		untilTime, err2 := time.Parse(time.RFC3339, until)
		if err1 == nil && err2 == nil && untilTime.After(fromTime) {
			years := untilTime.Year() - fromTime.Year()
			months = float64(years*12 + int(untilTime.Month()) - int(fromTime.Month()) + 1)
		}
	}

	if months <= 0 {
		months = 1
	}

	return float64(total) / months, nil
}
