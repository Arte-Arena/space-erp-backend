package leads

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

func GetOneTier(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "ID inválido", nil, 0)
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

	collection := mongoClient.Database(database.GetDB()).Collection("lead_tiers")

	var tier schemas.LeadTier
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&tier)
	if err != nil {
		utils.SendResponse(w, http.StatusNotFound, "Tier não encontrado", nil, 0)
		return
	}

	utils.SendResponse(w, http.StatusOK, "", tier, 0)
}
