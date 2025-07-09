package report

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetAllCommercialGoals(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao conectar com o banco de dados", nil, utils.ERROR_TO_CREATE_EXTERNAL_CONNECTION)
		return
	}
	defer client.Disconnect(ctx)

	collection := client.Database(database.GetDB()).Collection(database.COLLECTION_COMMERCIAL_GOALS)

	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao buscar metas comerciais", nil, utils.ERROR_TO_FIND_IN_MONGODB)
		return
	}
	defer cursor.Close(ctx)

	var goals []schemas.ReportCommercialGoals
	if err := cursor.All(ctx, &goals); err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao decodificar metas comerciais", nil, utils.ERROR_TO_FIND_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", goals, 0)
}
