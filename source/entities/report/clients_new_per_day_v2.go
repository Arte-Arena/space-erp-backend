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

func GetClientsNewPerDayV2(from, until string) (map[string]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_CLIENTS)

	filter := bson.D{}
	if from != "" || until != "" {
		dateFilter := bson.D{}
		if from != "" {
			if fromTime, err := time.Parse("2006-01-02", from); err == nil {
				dateFilter = append(dateFilter, bson.E{Key: "$gte", Value: fromTime})
			}
		}
		if until != "" {
			if untilTime, err := time.Parse("2006-01-02", until); err == nil {
				endOfDay := untilTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
				dateFilter = append(dateFilter, bson.E{Key: "$lte", Value: endOfDay})
			}
		}
		if len(dateFilter) > 0 {
			filter = append(filter, bson.E{Key: "created_at", Value: dateFilter})
		}
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "year", Value: bson.D{{Key: "$year", Value: "$created_at"}}},
				{Key: "month", Value: bson.D{{Key: "$month", Value: "$created_at"}}},
				{Key: "day", Value: bson.D{{Key: "$dayOfMonth", Value: "$created_at"}}},
			}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id.year", Value: 1}, {Key: "_id.month", Value: 1}, {Key: "_id.day", Value: 1}}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]int64)
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

		var year32, month32, day32 int32
		switch id := idVal.(type) {
		case bson.M:
			year32 = castToInt32(id["year"])
			month32 = castToInt32(id["month"])
			day32 = castToInt32(id["day"])
		case bson.D:
			for _, elem := range id {
				if elem.Key == "year" {
					year32 = castToInt32(elem.Value)
				}
				if elem.Key == "month" {
					month32 = castToInt32(elem.Value)
				}
				if elem.Key == "day" {
					day32 = castToInt32(elem.Value)
				}
			}
		}

		countAny := doc["count"]
		var count64 int64
		switch c := countAny.(type) {
		case int32:
			count64 = int64(c)
		case int64:
			count64 = c
		case float64:
			count64 = int64(c)
		}

		key := fmt.Sprintf("%04d-%02d-%02d", year32, month32, day32)
		result[key] = count64
	}
	return result, nil
}
