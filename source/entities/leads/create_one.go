package leads

import (
	"api/database"
	"api/schemas"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateOne(w http.ResponseWriter, r *http.Request) {
	lead := &schemas.Lead{}
	if err := json.NewDecoder(r.Body).Decode(&lead); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.LEADS_INVALID_REQUEST_DATA)
		return
	}

	lead.CreatedAt = time.Now()
	lead.UpdatedAt = time.Now()

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

	_, err = collection.InsertOne(ctx, lead)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_INSERT_LEAD_TO_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusCreated, "", nil, 0)
}
