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

func GetCommercialBudgetsStatusPercentages(seller bson.ObjectID, from, until string) (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)

	coll := client.Database(database.GetDB()).Collection(database.COLLECTION_BUDGETS)

	baseFilter := bson.D{{Key: "seller", Value: seller}}
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
			baseFilter = append(baseFilter, bson.E{Key: "created_at", Value: dateFilter})
		}
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: baseFilter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$approved"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var total int64
	var approvedCount int64
	var notApprovedCount int64

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		count, _ := doc["count"].(int32)
		approved, _ := doc["_id"].(bool)
		if approved {
			approvedCount += int64(count)
		} else {
			notApprovedCount += int64(count)
		}
		total += int64(count)
	}

	if total == 0 {
		return map[string]float64{"approved": 0, "not_approved": 0}, nil
	}

	return map[string]float64{
		"approved":     (float64(approvedCount) / float64(total)) * 100.0,
		"not_approved": (float64(notApprovedCount) / float64(total)) * 100.0,
	}, nil
}
