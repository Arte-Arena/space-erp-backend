package report

import (
	"api/database"
	"api/schemas"
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

func CreateCommercialGoal(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	var goal schemas.ReportCommercialGoals
	err := json.NewDecoder(r.Body).Decode(&goal)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "Dados de requisição inválidos", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	if goal.Name == "" || goal.GoalType == "" || goal.RelatedTo == "" {
		utils.SendResponse(w, http.StatusBadRequest, "Campos obrigatórios não preenchidos", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	validGoalType := goal.GoalType == schemas.REPORT_GOAL_TYPE_DAILY ||
		goal.GoalType == schemas.REPORT_GOAL_TYPE_MONTHLY ||
		goal.GoalType == schemas.REPORT_GOAL_TYPE_YEARLY
	if !validGoalType {
		utils.SendResponse(w, http.StatusBadRequest, "Tipo de meta inválido", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	validRelatedTo := goal.RelatedTo == schemas.REPORT_GOAL_RELATED_BUDGETS ||
		goal.RelatedTo == schemas.REPORT_GOAL_RELATED_CLIENTS ||
		goal.RelatedTo == schemas.REPORT_GOAL_RELATED_LEADS ||
		goal.RelatedTo == schemas.REPORT_GOAL_RELATED_ORDERS
	if !validRelatedTo {
		utils.SendResponse(w, http.StatusBadRequest, "Tipo relacionado inválido", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
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

	now := time.Now()
	goal.CreatedAt = now
	goal.UpdatedAt = now

	result, err := collection.InsertOne(ctx, goal)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao criar meta comercial", nil, utils.ERROR_TO_INSERT_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusCreated, "Meta comercial criada com sucesso", bson.M{"_id": result.InsertedID}, 0)
}
