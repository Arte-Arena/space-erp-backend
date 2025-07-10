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

func UpdateCommercialGoal(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), database.MONGO_TIMEOUT)
	defer cancel()

	idStr := r.PathValue("id")

	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "ID inválido", nil, utils.INVALID_CLIENT_ID_FORMAT)
		return
	}

	var goalUpdate schemas.ReportCommercialGoals
	err = json.NewDecoder(r.Body).Decode(&goalUpdate)
	if err != nil {
		utils.SendResponse(w, http.StatusBadRequest, "Dados de requisição inválidos", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
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

	var existingGoal schemas.ReportCommercialGoals
	err = collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&existingGoal)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.SendResponse(w, http.StatusNotFound, "Meta comercial não encontrada", nil, utils.NOT_FOUND)
		} else {
			utils.SendResponse(w, http.StatusInternalServerError, "Erro ao buscar meta comercial", nil, utils.ERROR_TO_FIND_IN_MONGODB)
		}
		return
	}

	update := bson.D{}

	if goalUpdate.Name != "" {
		update = append(update, bson.E{Key: "name", Value: goalUpdate.Name})
	}

	if goalUpdate.GoalType != "" {
		validGoalType := goalUpdate.GoalType == schemas.REPORT_GOAL_TYPE_MONTHLY ||
			goalUpdate.GoalType == schemas.REPORT_GOAL_TYPE_YEARLY
		if !validGoalType {
			utils.SendResponse(w, http.StatusBadRequest, "Tipo de meta inválido", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
			return
		}
		update = append(update, bson.E{Key: "goal_type", Value: goalUpdate.GoalType})
	}

	if goalUpdate.RelatedTo != "" {
		validRelatedTo := goalUpdate.RelatedTo == schemas.REPORT_GOAL_RELATED_BUDGETS ||
			goalUpdate.RelatedTo == schemas.REPORT_GOAL_RELATED_CLIENTS ||
			goalUpdate.RelatedTo == schemas.REPORT_GOAL_RELATED_LEADS ||
			goalUpdate.RelatedTo == schemas.REPORT_GOAL_RELATED_ORDERS
		if !validRelatedTo {
			utils.SendResponse(w, http.StatusBadRequest, "Tipo relacionado inválido", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
			return
		}
		update = append(update, bson.E{Key: "related_to", Value: goalUpdate.RelatedTo})
	}

	if goalUpdate.TargetValue > 0 {
		update = append(update, bson.E{Key: "target_value", Value: goalUpdate.TargetValue})
	}

	update = append(update, bson.E{Key: "updated_at", Value: time.Now()})

	if len(update) == 1 {
		utils.SendResponse(w, http.StatusBadRequest, "Nenhum dado para atualizar", nil, utils.SPACE_DESK_INVALID_REQUEST_DATA)
		return
	}

	_, err = collection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: id}},
		bson.D{{Key: "$set", Value: update}},
	)
	if err != nil {
		utils.SendResponse(w, http.StatusInternalServerError, "Erro ao atualizar meta comercial", nil, utils.ERROR_TO_UPDATE_IN_MONGODB)
		return
	}

	utils.SendResponse(w, http.StatusOK, "Meta comercial atualizada com sucesso", nil, 0)
}
