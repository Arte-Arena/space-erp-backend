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

func GetLeadsTimeToClosePurchase(from, until string) (float64, error) {
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
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_BUDGETS},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "related_lead"},
			{Key: "as", Value: "budgets"},
		}}},
		bson.D{{Key: "$unwind", Value: "$budgets"}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_ORDERS},
			{Key: "localField", Value: "budgets._id"},
			{Key: "foreignField", Value: "related_budget"},
			{Key: "as", Value: "orders"},
		}}},
		bson.D{{Key: "$unwind", Value: "$orders"}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "orders.created_at", Value: 1}}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$_id"},
			{Key: "lead_created_at", Value: bson.D{{Key: "$first", Value: "$created_at"}}},
			{Key: "first_purchase_at", Value: bson.D{{Key: "$first", Value: "$orders.created_at"}}},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "days_to_purchase", Value: bson.D{
				{Key: "$divide", Value: bson.A{
					bson.D{{Key: "$subtract", Value: bson.A{"$first_purchase_at", "$lead_created_at"}}},
					1000 * 60 * 60 * 24,
				}},
			}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "avg_days", Value: bson.D{{Key: "$avg", Value: "$days_to_purchase"}}},
		}}},
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
		if avg, ok := result[0]["avg_days"].(float64); ok {
			return avg, nil
		}
	}
	return 0, nil
}
