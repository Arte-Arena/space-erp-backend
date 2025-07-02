package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"net/http"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func GetAllPixConfig(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	idParam := r.URL.Query().Get("id")
	nomeParam := r.URL.Query().Get("nome")
	chaveParam := r.URL.Query().Get("chave")

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	collection := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CONFIG)

	var settings Settings
	err = collection.FindOne(ctx, bson.M{"type": "global"}).Decode(&settings)
	if err != nil {
		utils.SendResponse(w, http.StatusNotFound, "", nil, utils.NOT_FOUND)
		return
	}

	var filtered []PixConfig
	for _, pix := range settings.Pix {
		match := true

		if idParam != "" {
			objID, err := bson.ObjectIDFromHex(idParam)
			if err != nil || pix.ID != objID {
				match = false
			}
		}

		if nomeParam != "" && !StringContainsCI(pix.Nome, nomeParam) {
			match = false
		}

		if chaveParam != "" && !StringContainsCI(pix.Chave, chaveParam) {
			match = false
		}

		if match {
			filtered = append(filtered, pix)
		}
	}

	utils.SendResponse(w, http.StatusOK, "", map[string]interface{}{
		"keys": filtered,
	}, 0)
}

func StringContainsCI(a, b string) bool {
	return strings.Contains(strings.ToLower(a), strings.ToLower(b))
}
