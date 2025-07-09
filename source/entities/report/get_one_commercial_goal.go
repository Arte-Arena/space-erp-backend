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

func GetOneCommercialGoal(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	idStr := r.PathValue("id")

	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "ID inválido", nil, utils.INVALID_CLIENT_ID_FORMAT)
		return
	}

	mongoURI := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao conectar com o banco de dados", nil, utils.ERROR_TO_CREATE_EXTERNAL_CONNECTION)
		return
	}
	defer client.Disconnect(ctx)

	collection := client.Database(database.GetDB()).Collection(database.COLLECTION_COMMERCIAL_GOALS)

	var goal schemas.ReportCommercialGoals
	err = collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&goal)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.SendResponse(w, http.StatusNotFound, "Meta comercial não encontrada", nil, utils.NOT_FOUND)
		} else {
			utils.SendResponse(w, http.StatusInternalServerError, "Erro ao buscar meta comercial", nil, utils.ERROR_TO_FIND_IN_MONGODB)
		}
		return
	}

	utils.SendResponse(w, http.StatusOK, "Meta comercial encontrada com sucesso", goal, 0)
}
