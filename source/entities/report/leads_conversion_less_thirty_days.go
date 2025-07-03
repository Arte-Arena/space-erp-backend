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

func GetLeadsConversionLessThirtyDays(from, until string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		return 0, err
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_LEADS)

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

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "related_client", Value: bson.D{{Key: "$ne", Value: nil}}}}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_CLIENTS},
			{Key: "localField", Value: "related_client"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "client"},
		}}},
		bson.D{{Key: "$unwind", Value: "$client"}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "conversion_days", Value: bson.D{
				{Key: "$divide", Value: bson.A{
					bson.D{{Key: "$subtract", Value: bson.A{"$client.created_at", "$created_at"}}},
					1000 * 60 * 60 * 24,
				}},
			}},
		}}},
		bson.D{{Key: "$match", Value: bson.D{{Key: "conversion_days", Value: bson.D{{Key: "$lte", Value: 30}}}}}},
		bson.D{{Key: "$count", Value: "count"}},
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
		if count, ok := result[0]["count"].(int32); ok {
			return int64(count), nil
		}
	}
	return 0, nil
}
