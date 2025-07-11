package report

import (
	"context"
	"fmt"
	"math"
	"time"

	"api/database"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetSuperadminSellersMonthlySales(client *mongo.Client, sellerIDs []bson.ObjectID, from, until string) (map[bson.ObjectID]struct {
	Name  string
	Sales map[string]float64
}, error) {
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
		bson.D{{Key: "$unwind", Value: "$billing.installments"}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "seller", Value: "$seller"},
				{Key: "year", Value: bson.D{{Key: "$year", Value: "$created_at"}}},
				{Key: "month", Value: bson.D{{Key: "$month", Value: "$created_at"}}},
			}},
			{Key: "total_value", Value: bson.D{{Key: "$sum", Value: "$billing.installments.value"}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "_id.year", Value: 1}, {Key: "_id.month", Value: 1}}}},
	}
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	result := map[bson.ObjectID]map[string]float64{}
	for cursor.Next(ctx) {
		var doc struct {
			ID struct {
				Seller bson.ObjectID `bson:"seller"`
				Year   int32         `bson:"year"`
				Month  int32         `bson:"month"`
			} `bson:"_id"`
			TotalValue float64 `bson:"total_value"`
		}
		if err := cursor.Decode(&doc); err == nil {
			key := fmt.Sprintf("%04d-%02d", doc.ID.Year, doc.ID.Month)
			if _, ok := result[doc.ID.Seller]; !ok {
				result[doc.ID.Seller] = map[string]float64{}
			}
			result[doc.ID.Seller][key] = doc.TotalValue
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
		Name  string
		Sales map[string]float64
	}{}
	for id, sales := range result {
		if id.IsZero() {
			continue
		}
		roundedSales := map[string]float64{}
		for k, v := range sales {
			roundedSales[k] = math.Round(v*100) / 100
		}
		final[id] = struct {
			Name  string
			Sales map[string]float64
		}{Name: nameMap[id], Sales: roundedSales}
	}
	return final, nil
}
