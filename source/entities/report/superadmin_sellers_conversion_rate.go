package report

import (
	"context"
	"time"

	"api/database"

	"math"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetSuperadminSellersConversionRate(client *mongo.Client, sellerIDs []bson.ObjectID, from, until string) (map[bson.ObjectID]struct {
	Name string
	Rate float64
}, error) {
	ctx := context.Background()
	filter := bson.D{}
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
			{Key: "_id", Value: bson.D{{Key: "seller", Value: "$seller"}, {Key: "approved", Value: "$approved"}}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	total := map[bson.ObjectID]int64{}
	approved := map[bson.ObjectID]int64{}
	for cursor.Next(ctx) {
		var doc struct {
			ID struct {
				Seller   bson.ObjectID `bson:"seller"`
				Approved bool          `bson:"approved"`
			} `bson:"_id"`
			Count int64 `bson:"count"`
		}
		if err := cursor.Decode(&doc); err == nil {
			total[doc.ID.Seller] += doc.Count
			if doc.ID.Approved {
				approved[doc.ID.Seller] += doc.Count
			}
		}
	}
	result := map[bson.ObjectID]float64{}
	for id, t := range total {
		if t > 0 {
			result[id] = (float64(approved[id]) / float64(t)) * 100.0
		} else {
			result[id] = 0
		}
	}

	userColl := client.Database(database.GetDB()).Collection(database.COLLECTION_USERS)
	var ids []bson.ObjectID
	for id := range result {
		ids = append(ids, id)
	}
	userCursor, err := userColl.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer userCursor.Close(ctx)
	nameMap := map[bson.ObjectID]string{}
	for userCursor.Next(ctx) {
		var user struct {
			ID   bson.ObjectID `bson:"_id"`
			Name string        `bson:"name"`
		}
		if err := userCursor.Decode(&user); err == nil {
			nameMap[user.ID] = user.Name
		}
	}
	final := map[bson.ObjectID]struct {
		Name string
		Rate float64
	}{}
	for id, value := range result {
		if id.IsZero() {
			continue
		}
		final[id] = struct {
			Name string
			Rate float64
		}{Name: nameMap[id], Rate: math.Round(value*100) / 100}
	}
	return final, nil
}
