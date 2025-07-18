package spacedesk

import (
	"api/database"
	"api/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UpdateChatDescriptionBody struct {
	ID          string `json:"id"`
	Description string `json:"description" bson:"description"`
	UserId      string `json:"user_id" bson:"user_id"`
}

func UpdateChatDescription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "Erro ao ler o body", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	var body UpdateChatDescriptionBody
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "JSON inválido", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	if body.ID == "" {
		utils.SendResponse(w, http.StatusBadRequest, "Campo 'id' obrigatório ausente", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}
	var raw map[string]any
	_ = json.Unmarshal(bodyBytes, &raw)

	ctx, cancel := context.WithTimeout(r.Context(), database.MONGO_TIMEOUT)
	defer cancel()

	mongoURI := os.Getenv(utils.MONGODB_URI)
	opts := options.Client().ApplyURI(mongoURI)
	mongoClient, err := mongo.Connect(opts)
	if err != nil {
		utils.SendResponse(w, http.StatusBadGateway, "", nil, utils.CANNOT_CONNECT_TO_MONGODB)
		return
	}
	defer mongoClient.Disconnect(ctx)

	colChats := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_CHAT)
	var chatDoc struct {
		ClientePhoneNumber string `bson:"cliente_phone_number"`
		LastMessage        any    `bson:"last_message_timestamp"`
	}

	objID, err := bson.ObjectIDFromHex(body.ID)
	if err != nil {
		log.Println("Erro ao converter ID do chat para ObjectID:", err, "ID recebido:", body.ID)
		utils.SendResponse(w, http.StatusBadRequest, "ID do chat inválido", nil, utils.INVALID_CHAT_ID_FORMAT)
		return
	}
	err = colChats.FindOne(r.Context(), bson.M{"_id": objID}).Decode(&chatDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.SendResponse(w, http.StatusNotFound, "Chat não encontrado", nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
			return
		}
		log.Println("Erro ao buscar chat:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao buscar chat: "+err.Error(), nil, utils.CANNOT_FIND_SPACE_DESK_GROUP_ID_FORMAT)
		return
	}

	now := time.Now().UTC()
	internalID := bson.NewObjectID().Hex()
	colMessages := mongoClient.Database(database.GetDB()).Collection(database.COLLECTION_SPACE_DESK_MESSAGE)

	message := bson.M{
		"body":              body.Description,
		"chat_id":           objID,
		"by":                body.UserId,
		"from":              "company",
		"created_at":        now.Format(time.RFC3339),
		"message_id":        internalID,
		"message_timestamp": fmt.Sprint(now.Unix()),
		"type":              "annotation",
		"status":            "",
		"updated_at":        now.Format(time.RFC3339),
	}

	_, err = colMessages.InsertOne(ctx, message)
	if err != nil {
		log.Println("Erro ao inserir anotação interna no MongoDB:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao salvar anotação interna", nil, utils.ERROR_TO_INSERT_IN_MONGODB)
		return
	}

	// Atualiza o campo de descrição no chat (opcional)
	_, err = colChats.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{
			"description": body.Description,
			"updated_at":  now.Format(time.RFC3339),
		},
	})
	if err != nil {
		log.Println("Erro ao atualizar descrição do chat:", err)
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao atualizar descrição do chat", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Anotação interna registrada com sucesso", nil, 0)
}
