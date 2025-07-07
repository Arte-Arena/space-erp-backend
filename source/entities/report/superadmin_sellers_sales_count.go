package report

import (
	"context"
	"time"

	"api/database"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetSuperadminSellersSalesCount(client *mongo.Client, sellerIDs []bson.ObjectID, from, until string) (map[bson.ObjectID]int64, error) {
	ctx := context.Background()
	filter := bson.D{{Key: "approved", Value: true}}
	if len(sellerIDs) > 0 {
		filter = append(filter, bson.E{Key: "seller", Value: bson.D{{Key: "$in", Value: sellerIDs}}})
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
	coll := client.Database(database.GetDB()).Collection(database.COLLECTION_BUDGETS)
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$seller"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	result := map[bson.ObjectID]int64{}
	for cursor.Next(ctx) {
		var doc struct {
			ID    bson.ObjectID `bson:"_id"`
			Count int64         `bson:"count"`
		}
		if err := cursor.Decode(&doc); err == nil {
			result[doc.ID] = doc.Count
		}
	}
	return result, nil
}
