package spacedesk

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"api/database"
	"api/utils"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DeleteOneReadyMessage(w http.ResponseWriter, r *http.Request) {
	// Permitir apenas DELETE
	if r.Method != http.MethodDelete {
		http.Error(w, "Método não permitido. Use DELETE.", http.StatusMethodNotAllowed)
		return
	}

	// Pega o id da query (ex: /v1/space-desk/ready-messages?id=123456)
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID não fornecido.", http.StatusBadRequest)
		return
	}

	// Converter string para ObjectID
	objID, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "ID inválido.", http.StatusBadRequest)
		return
	}

	// Conexão MongoDB
	mongoUri := os.Getenv(utils.MONGODB_URI)
	if mongoUri == "" {
		http.Error(w, "MongoDB URI não configurado.", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOpts := options.Client().ApplyURI(mongoUri)
	dbClient, err := mongo.Connect(clientOpts)
	if err != nil {
		http.Error(w, "Erro ao conectar ao MongoDB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dbClient.Disconnect(ctx)

	col := dbClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_READY_MESSAGE)

	res, err := col.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		http.Error(w, "Erro ao deletar mensagem: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retorna sucesso, inclusive quantos registros deletados
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"deleted": res.DeletedCount,
		"id":      idStr,
	})
}
