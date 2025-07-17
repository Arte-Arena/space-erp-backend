package report

import (
	"api/database"
	"context"
	"encoding/json"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetBudgetsSalesValueBySegment(from, until string, notApproved bool) (map[string]float64, error) {
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

	var filter bson.D
	if notApproved {
		filter = bson.D{{Key: "approved", Value: false}}
	} else {
		filter = bson.D{{Key: "approved", Value: true}}
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

	if notApproved {
		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		type Product struct {
			Preco      float64 `json:"preco"`
			Quantidade int     `json:"quantidade"`
		}

		result := make(map[string]float64)
		for cursor.Next(ctx) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				continue
			}

			segment := "Sem segmento"
			if lead, ok := doc["lead"].(bson.M); ok {
				if seg, ok := lead["segment"].(string); ok && seg != "" {
					segment = seg
				}
			}

			oldProductsList, _ := doc["old_products_list"].(string)
			var products []Product
			budgetTotal := 0.0
			if oldProductsList != "" {
				_ = json.Unmarshal([]byte(oldProductsList), &products)
				for _, p := range products {
					budgetTotal += p.Preco * float64(p.Quantidade)
				}
			}

			delivery, _ := doc["delivery"].(bson.M)
			if delivery != nil {
				if price, ok := delivery["price"].(float64); ok {
					budgetTotal += price
				}
			}
			result[segment] += budgetTotal
		}
		if err := cursor.Err(); err != nil {
			return nil, err
		}
		return result, nil
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_LEADS},
			{Key: "localField", Value: "related_lead"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "lead"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$lead"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		bson.D{{Key: "$unwind", Value: "$billing.installments"}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$lead.segment"},
			{Key: "total_value", Value: bson.D{{Key: "$sum", Value: "$billing.installments.value"}}},
		}}},
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
		segment, _ := doc["_id"].(string)
		total, _ := doc["total_value"].(float64)
		if segment == "" {
			segment = "Sem segmento"
		}
		result[segment] = total
	}

	return result, nil
}
