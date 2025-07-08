package report

import (
	"context"
	"math"
	"time"

	"api/database"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SellerRanking struct {
	SellerID      bson.ObjectID `bson:"_id" json:"seller_id"`
	SellerName    string        `json:"seller_name"`
	SalesCount    int64         `json:"sales_count"`
	TotalValue    float64       `json:"total_value"`
	AverageTicket float64       `json:"average_ticket"`
	Conversion    float64       `json:"conversion"`
}

func GetSuperadminSellersRanking(client *mongo.Client, from, until string) ([]SellerRanking, error) {
	ctx := context.Background()

	filterApproved := bson.D{{Key: "approved", Value: true}}
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
			filterApproved = append(filterApproved, bson.E{Key: "created_at", Value: dateFilter})
		}
	}
	coll := client.Database(database.GetDB()).Collection(database.COLLECTION_BUDGETS)
	pipelineApproved := bson.A{
		bson.D{{Key: "$match", Value: filterApproved}},
		bson.D{{Key: "$unwind", Value: "$billing.installments"}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$seller"},
			{Key: "total_value", Value: bson.D{{Key: "$sum", Value: "$billing.installments.value"}}},
			{Key: "sales_count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}
	cursor, err := coll.Aggregate(ctx, pipelineApproved)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	totalValue := map[bson.ObjectID]float64{}
	salesCount := map[bson.ObjectID]int64{}
	for cursor.Next(ctx) {
		var doc struct {
			SellerID   bson.ObjectID `bson:"_id"`
			TotalValue float64       `bson:"total_value"`
			SalesCount int64         `bson:"sales_count"`
		}
		if err := cursor.Decode(&doc); err == nil {
			totalValue[doc.SellerID] = doc.TotalValue
			salesCount[doc.SellerID] = doc.SalesCount
		}
	}

	filterTotal := bson.D{}
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
			filterTotal = append(filterTotal, bson.E{Key: "created_at", Value: dateFilter})
		}
	}
	pipelineTotal := bson.A{
		bson.D{{Key: "$match", Value: filterTotal}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$seller"},
			{Key: "total_budgets", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}
	cursor2, err := coll.Aggregate(ctx, pipelineTotal)
	if err != nil {
		return nil, err
	}
	defer cursor2.Close(ctx)
	totalBudgets := map[bson.ObjectID]int64{}
	for cursor2.Next(ctx) {
		var doc struct {
			SellerID     bson.ObjectID `bson:"_id"`
			TotalBudgets int64         `bson:"total_budgets"`
		}
		if err := cursor2.Decode(&doc); err == nil {
			totalBudgets[doc.SellerID] = doc.TotalBudgets
		}
	}

	userColl := client.Database(database.GetDB()).Collection(database.COLLECTION_USERS)
	var sellerIDs []bson.ObjectID
	for seller := range salesCount {
		sellerIDs = append(sellerIDs, seller)
	}
	userCursor, err := userColl.Find(ctx, bson.M{"_id": bson.M{"$in": sellerIDs}})
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

	rankings := []SellerRanking{}
	for seller, sales := range salesCount {
		total := totalBudgets[seller]
		conversion := 0.0
		if total > 0 {
			conversion = (float64(sales) / float64(total)) * 100.0
		}
		avgTicket := 0.0
		if sales > 0 {
			avgTicket = math.Round((totalValue[seller]/float64(sales))*100) / 100
		}
		rankings = append(rankings, SellerRanking{
			SellerID:      seller,
			SellerName:    nameMap[seller],
			SalesCount:    sales,
			TotalValue:    math.Round(totalValue[seller]*100) / 100,
			AverageTicket: math.Round(avgTicket*100) / 100,
			Conversion:    math.Round(conversion*100) / 100,
		})
	}
	return rankings, nil
}
