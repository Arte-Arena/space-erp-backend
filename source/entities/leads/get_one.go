package leads

import (
	"api/source/database"
	"api/source/utils"
	"context"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetOne(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.INVALID_LEAD_ID_FORMAT)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_LEADS)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: id}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_BUDGETS},
			{Key: "localField", Value: "related_budgets"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "related_budgets"},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_ORDERS},
			{Key: "localField", Value: "related_orders"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "related_orders"},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: database.COLLECTION_USERS},
			{Key: "localField", Value: "responsible"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "responsible_data"},
		}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$responsible_data"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$addFields", Value: bson.D{{Key: "responsible", Value: "$responsible_data"}}}},
		{{Key: "$project", Value: bson.D{{Key: "responsible_data", Value: 0}}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEAD_BY_ID_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_FIND_LEAD_BY_ID_IN_MONGODB)
		return
	}

	if len(results) == 0 {
		utils.SendResponse(w, http.StatusNotFound, "Lead não encontrado", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", results[0], 0)
}
