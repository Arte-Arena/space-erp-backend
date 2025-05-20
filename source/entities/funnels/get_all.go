package funnels

import (
	"api/source/database"
	"api/source/utils"
	"context"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAll(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(ctx, opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_FUNNELS)

	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$stages"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_LEADS},
			{Key: "localField", Value: "stages.related_leads"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "stages.related_leads_data"},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_BUDGETS},
			{Key: "localField", Value: "stages.related_budgets"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "stages.related_budgets_data"},
		}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "stages.related_leads", Value: "$stages.related_leads_data"},
			{Key: "stages.related_budgets", Value: "$stages.related_budgets_data"},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "stages.related_leads_data", Value: 0},
			{Key: "stages.related_budgets_data", Value: 0},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$_id"},
			{Key: "name", Value: bson.D{{Key: "$first", Value: "$name"}}},
			{Key: "type", Value: bson.D{{Key: "$first", Value: "$type"}}},
			{Key: "stages", Value: bson.D{{Key: "$push", Value: "$stages"}}},
			{Key: "created_at", Value: bson.D{{Key: "$first", Value: "$created_at"}}},
			{Key: "updated_at", Value: bson.D{{Key: "$first", Value: "$updated_at"}}},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_FUNNELS_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	var funnels []bson.M
	if err := cursor.All(ctx, &funnels); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_FUNNELS_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", funnels, 0)
}
