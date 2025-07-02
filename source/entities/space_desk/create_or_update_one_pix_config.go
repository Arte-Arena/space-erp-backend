package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Corrija os tipos de ID, use primitive.ObjectID
type Settings struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Type      string        `bson:"type" json:"type"`
	Pix       []PixConfig   `bson:"pix" json:"pix"`
	Phones    []PhoneConfig `bson:"phones" json:"phones"`
	Theme     string        `bson:"theme,omitempty" json:"theme,omitempty"`
	CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
}

type PixConfig struct {
	ID     bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Chave  string        `bson:"chave" json:"chave"`
	Tipo   string        `bson:"tipo" json:"tipo"` // Adicionado!
	Banco  string        `bson:"banco" json:"banco"`
	Status string        `bson:"status,omitempty" json:"status,omitempty"`
	Nome   string        `bson:"nome,omitempty" json:"nome,omitempty"`
	Ativo  bool          `bson:"ativo" json:"ativo"` // Adicionado!
}

func CreateOrUpdatePixConfig(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var payload PixConfig
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	// Validação dos campos obrigatórios
	if payload.Chave == "" || payload.Tipo == "" || payload.Banco == "" {
		utils.SendResponse(w, http.StatusBadRequest, "", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	// Valores padrão
	if payload.Status == "" {
		payload.Status = "Ativo"
	}
	if payload.Nome == "" {
		payload.Nome = "Chave Pix"
	}
	payload.Ativo = payload.Status == "Ativo"

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer client.Disconnect(ctx)

	collection := client.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CONFIG)

	// Busca config global
	var settings Settings
	_ = collection.FindOne(ctx, bson.M{"type": "global"}).Decode(&settings)

	// --- Limite de 5 chaves
	if payload.ID.IsZero() && len(settings.Pix) >= 5 {
		utils.SendResponse(w, http.StatusConflict, "", nil, utils.LIMIT_REACHED)
		return
	}

	// --- Duplicidade de chave (criação)
	if payload.ID.IsZero() {
		filterExists := bson.M{
			"type":      "global",
			"pix.chave": payload.Chave,
		}
		count, err := collection.CountDocuments(ctx, filterExists)
		if err != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_QUERY_MONGODB)
			return
		}
		if count > 0 {
			utils.SendResponse(w, http.StatusConflict, "", nil, utils.ALREADY_EXISTS)
			return
		}
	}

	if payload.ID.IsZero() {
		payload.ID = bson.NewObjectID()
		filter := bson.M{"type": "global"}
		update := bson.M{
			"$push":        bson.M{"pix": payload},
			"$setOnInsert": bson.M{"type": "global"},
			"$currentDate": bson.M{"updatedAt": true, "createdAt": true},
		}
		_, err = collection.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
		if err != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
			return
		}
		utils.SendResponse(w, http.StatusCreated, "", payload, 0)
		return
	}

	filter := bson.M{
		"type":    "global",
		"pix._id": payload.ID,
	}
	update := bson.M{
		"$set": bson.M{
			"pix.$.chave":  payload.Chave,
			"pix.$.tipo":   payload.Tipo,
			"pix.$.banco":  payload.Banco,
			"pix.$.status": payload.Status,
			"pix.$.nome":   payload.Nome,
			"pix.$.ativo":  payload.Ativo,
		},
		"$currentDate": bson.M{"updatedAt": true},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}
	if result.MatchedCount == 0 {
		utils.SendResponse(w, http.StatusNotFound, "", nil, utils.NOT_FOUND)
		return
	}
	utils.SendResponse(w, http.StatusOK, "", payload, 0)
}
