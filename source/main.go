package main

import (
	"api/entities/budgets"
	"api/entities/clients"
	"api/entities/funnels"
	funnelshistory "api/entities/funnels_history"
	"api/entities/leads"
	"api/entities/orders"
	"api/entities/report"
	spacedesk "api/entities/space_desk"
	users "api/entities/users"
	"api/middlewares"
	"api/utils"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	utils.LoadEnvVariables()

	env := os.Getenv(utils.ENV)
	if env == utils.ENV_RELEASE {
		fmt.Printf("\033[1;31;47m[ATENÇÃO] Rodando em ambiente de PRODUÇÃO!\033[0m\n")
	} else {
		fmt.Printf("[INFO] Ambiente atual: %s\n", env)
	}

	mux := http.NewServeMux()

	mux.Handle("GET /v1/user/{id}", middlewares.LaravelAuth(http.HandlerFunc(users.GetOneUser)))
	mux.Handle("GET /v1/user", middlewares.LaravelAuth(http.HandlerFunc(users.GetAllUsers)))

	mux.Handle("GET /v1/funnels", middlewares.LaravelAuth(http.HandlerFunc(funnels.GetAll)))
	mux.Handle("GET /v1/funnels/{id}", middlewares.LaravelAuth(http.HandlerFunc(funnels.GetOne)))
	mux.Handle("POST /v1/funnels", middlewares.LaravelAuth(http.HandlerFunc(funnels.CreateOne)))
	mux.Handle("PATCH /v1/funnels/{id}", middlewares.LaravelAuth(http.HandlerFunc(funnels.UpdateOne)))
	mux.Handle("DELETE /v1/funnels/{id}", middlewares.LaravelAuth(http.HandlerFunc(funnels.DeleteOne)))
	mux.HandleFunc("/v1/ws/funnels", funnels.FunnelWebSocketHandler)

	mux.Handle("GET /v1/leads", middlewares.LaravelAuth(http.HandlerFunc(leads.GetAll)))
	mux.Handle("GET /v1/leads/{id}", middlewares.LaravelAuth(http.HandlerFunc(leads.GetOne)))
	mux.Handle("POST /v1/leads", middlewares.LaravelAuth(http.HandlerFunc(leads.CreateOne)))
	mux.Handle("PATCH /v1/leads/{id}", middlewares.LaravelAuth(http.HandlerFunc(leads.UpdateOne)))
	mux.Handle("GET /v1/leads/tiers", middlewares.LaravelAuth(http.HandlerFunc(leads.GetAllTiers)))
	mux.Handle("GET /v1/leads/tiers/{id}", middlewares.LaravelAuth(http.HandlerFunc(leads.GetOneTier)))
	mux.Handle("POST /v1/leads/tiers", middlewares.LaravelAuth(http.HandlerFunc(leads.CreateOneTier)))
	mux.Handle("PATCH /v1/leads/tiers/{id}", middlewares.LaravelAuth(http.HandlerFunc(leads.UpdateOneTier)))

	mux.Handle("GET /v1/clients", middlewares.LaravelAuth(http.HandlerFunc(clients.GetAll)))
	mux.Handle("GET /v1/clients/{id}", middlewares.LaravelAuth(http.HandlerFunc(clients.GetOne)))

	mux.Handle("GET /v1/users", middlewares.LaravelAuth(http.HandlerFunc(users.GetAll)))
	mux.Handle("GET /v1/users/{id}", middlewares.LaravelAuth(http.HandlerFunc(users.GetOne)))
	mux.Handle("PATCH /v1/users/{id}", middlewares.LaravelAuth(http.HandlerFunc(users.UpdateOne)))
	mux.Handle("GET /v1/users/commercial/budgets", middlewares.LaravelAuth(http.HandlerFunc(users.GetCommercialBudgets)))
	mux.Handle("GET /v1/users/commercial/reports/budgets", middlewares.LaravelAuth(http.HandlerFunc(users.GetCommercialBudgetsReport)))
	mux.Handle("GET /v1/users/superadmin/reports/commercial", middlewares.LaravelAuth(http.HandlerFunc(users.GetSuperadminSellersPerformanceReport)))

	mux.Handle("GET /v1/budgets", middlewares.LaravelAuth(http.HandlerFunc(budgets.GetAll)))
	mux.Handle("POST /v1/budgets/shipping/{service}", middlewares.LaravelAuth(http.HandlerFunc(budgets.CreateShippingQuote)))
	mux.Handle("GET /v1/budgets/{id}", middlewares.LaravelAuth(http.HandlerFunc(budgets.GetOne)))

	mux.Handle("GET /v1/orders", middlewares.LaravelAuth(http.HandlerFunc(orders.GetAll)))
	mux.Handle("GET /v1/orders/{id}", middlewares.LaravelAuth(http.HandlerFunc(orders.GetOne)))

	mux.Handle("GET /v1/reports", middlewares.LaravelAuth(http.HandlerFunc(report.GetByQuery)))
	mux.Handle("GET /v1/reports/commercial/goals", middlewares.LaravelAuth(http.HandlerFunc(report.GetAllCommercialGoals)))
	mux.Handle("POST /v1/reports/commercial/goals", middlewares.LaravelAuth(http.HandlerFunc(report.CreateCommercialGoal)))
	mux.Handle("GET /v1/reports/commercial/goals/{id}", middlewares.LaravelAuth(http.HandlerFunc(report.GetOneCommercialGoal)))
	mux.Handle("PATCH /v1/reports/commercial/goals/{id}", middlewares.LaravelAuth(http.HandlerFunc(report.UpdateCommercialGoal)))
	mux.Handle("DELETE /v1/reports/commercial/goals/{id}", middlewares.LaravelAuth(http.HandlerFunc(report.DeleteCommercialGoal)))

	mux.Handle("POST /v1/funnels_history", middlewares.LaravelAuth(http.HandlerFunc(funnelshistory.CreateOne)))
	mux.Handle("GET /v1/funnels_history/{id}", middlewares.LaravelAuth(http.HandlerFunc(funnelshistory.GetAll)))

	mux.Handle("GET /v1/space-desk/chats", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetAllChats)))

	mux.Handle("POST /v1/space-desk/webhook-whatsapp", http.HandlerFunc(spacedesk.CreateOneWebhookWhatsapp))

	mux.Handle("GET /v1/space-desk/messages", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetAllMessages)))
	mux.Handle("GET /v1/space-desk/chat-messages", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetAllMessagesByChatId)))

	mux.Handle("GET /v1/space-desk/status", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetAllStatuses)))
	mux.Handle("GET /v1/space-desk/service-queue", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetServiceQueue)))
	mux.Handle("GET /v1/desk/queue", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetServiceQueueV2)))

	mux.Handle("POST /v1/space-desk/message", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateOneMessage)))
	mux.Handle("POST /v1/space-desk/media", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateOneMedia)))
	mux.Handle("POST /v1/space-desk/order-details", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateOrderDetails)))
	mux.Handle("POST /v1/space-desk/order-template", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateOrderDetails)))
	mux.Handle("POST /v1/space-desk/pix-message", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreatePixMessage)))

	mux.Handle("GET /v1/space-desk/media-base64", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.HandlerMediaBase64)))
	mux.Handle("GET /v1/space-desk/media-download", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.HandlerMediaDownload)))

	mux.Handle("POST /v1/space-desk/group", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateOneGroup)))
	mux.Handle("PATCH /v1/space-desk/group", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.UpdateOneGroup)))
	mux.Handle("PATCH /v1/space-desk/group-users", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.UpdateGroupUsers)))
	mux.Handle("POST /v1/space-desk/group-users", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.AddUsersToGroup)))
	mux.Handle("GET /v1/space-desk/group", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetAllGroups)))
	mux.Handle("DELETE /v1/space-desk/group", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.DeleteGroup)))
	mux.Handle("DELETE /v1/space-desk/group-users", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.DeleteUserFromGroup)))
	mux.Handle("GET /v1/space-desk/group-chats/{groupId}", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetChatsFromGroup)))
	
	mux.Handle("PATCH /v1/space-desk/chats/status", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.UpdateChatStatus)))
	mux.Handle("PATCH /v1/space-desk/chats/user", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.UpdateChatUser)))
	mux.Handle("PATCH /v1/space-desk/chat-description", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.UpdateChatDescription)))

	mux.Handle("POST /v1/space-desk/group-chat", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.AddChatToGroup)))
	mux.Handle("DELETE /v1/space-desk/group-chat", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.DeleteChatFromGroup)))
	mux.Handle("GET /v1/space-desk/chats-group", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetChatsByGroup)))

	mux.Handle("POST /v1/space-desk/ready-messages", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateOneReadyMessage)))
	mux.Handle("PUT /v1/space-desk/ready-messages", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.UpdateOneReadyMessage)))
	mux.Handle("DELETE /v1/space-desk/ready-messages", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.DeleteOneReadyMessage)))
	mux.Handle("GET /v1/space-desk/ready-messages", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetAllReadyMessages)))

	mux.Handle("POST /v1/space-desk/template-messages", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateOneTemplate)))
	mux.Handle("GET /v1/space-desk/template-messages", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.ListAndSyncD360Templates)))
	mux.Handle("DELETE /v1/space-desk/template-messages/", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.DeleteD360Template)))

	mux.Handle("POST /v1/space-desk/poll", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateOnePoll)))
	mux.Handle("POST /v1/space-desk/list", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateListMessage)))
	mux.Handle("POST /v1/space-desk/location", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateLocationRequestMessage)))

	mux.Handle("POST /v1/space-desk/phone-config", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreatePhoneConfig)))
	mux.Handle("PATCH /v1/space-desk/phone-config", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.UpdatePhoneConfig)))
	mux.Handle("GET /v1/space-desk/phone-config", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetAllPhoneConfig)))
	mux.Handle("DELETE /v1/space-desk/phone-config", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.DeletePhoneConfig)))

	mux.Handle("PUT /v1/space-desk/pix-config", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateOrUpdatePixConfig)))
	mux.Handle("GET /v1/space-desk/pix-config", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetAllPixConfig)))
	mux.Handle("DELETE /v1/space-desk/pix-config", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.DeletePixConfig)))

	mux.HandleFunc("/v1/ws/space-desk", spacedesk.SpaceDeskWebSocketHandler)

	fmt.Printf("Servidor iniciado na porta %s às %s\n", os.Getenv(utils.PORT), time.Now().Format("2006-01-02 15:04:05"))
	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv(utils.PORT)), middlewares.SecurityHeaders(middlewares.Cors(mux)))
}
