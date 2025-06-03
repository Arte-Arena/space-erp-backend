package main

import (
	"api/entities/budgets"
	"api/entities/clients"
	"api/entities/funnels"
	funnelshistory "api/entities/funnels_history"
	"api/entities/leads"
	"api/entities/orders"
	spacedesk "api/entities/space_desk"
	"api/middlewares"
	"api/utils"
	"fmt"
	"net/http"
	"os"
)

func main() {
	utils.LoadEnvVariables()

	mux := http.NewServeMux()

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

	mux.Handle("GET /v1/clients", middlewares.LaravelAuth(http.HandlerFunc(clients.GetAll)))
	mux.Handle("GET /v1/clients/{id}", middlewares.LaravelAuth(http.HandlerFunc(clients.GetOne)))

	mux.Handle("GET /v1/budgets", middlewares.LaravelAuth(http.HandlerFunc(budgets.GetAll)))
	mux.Handle("GET /v1/budgets/{id}", middlewares.LaravelAuth(http.HandlerFunc(budgets.GetOne)))

	mux.Handle("GET /v1/orders", middlewares.LaravelAuth(http.HandlerFunc(orders.GetAll)))
	mux.Handle("GET /v1/orders/{id}", middlewares.LaravelAuth(http.HandlerFunc(orders.GetOne)))

	mux.Handle("GET /v1/reports", middlewares.LaravelAuth(http.HandlerFunc(leads.GetAll)))

	mux.Handle("POST /v1/funnels_history", middlewares.LaravelAuth(http.HandlerFunc(funnelshistory.CreateOne)))
	mux.Handle("GET /v1/funnels_history/{id}", middlewares.LaravelAuth(http.HandlerFunc(funnelshistory.GetAll)))

	mux.Handle("GET /v1/space-desk/chats", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetAllChats)))
	mux.Handle("GET /v1/space-desk/messages", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetAllMessages)))
	mux.Handle("GET /v1/space-desk/status", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetAllStatuses)))
	
	mux.Handle("POST /v1/space-desk/webhook-whatsapp", http.HandlerFunc(spacedesk.CreateOneWebhookWhatsapp))
	
	mux.Handle("POST /v1/space-desk/message", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateOneMessage)))
	mux.Handle("POST /v1/space-desk/media", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateOneMedia)))
	
	mux.Handle("GET /v1/space-desk/media-base64", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.HandlerMediaBase64)))
	mux.Handle("GET /v1/space-desk/media-download", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.HandlerMediaDownload)))
	
	mux.Handle("POST /v1/space-desk/group", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.CreateOneGroup)))
	mux.Handle("PATCH /v1/space-desk/group", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.UpdateOneGroup)))
	mux.Handle("DELETE /v1/space-desk/group", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.DeleteGroup)))
	mux.Handle("POST /v1/space-desk/group-chat", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.AddChatToGroup)))
	mux.Handle("DELETE /v1/space-desk/group-chat", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.DeleteChatFromGroup)))
	mux.Handle("GET /v1/space-desk/chats-group", middlewares.LaravelAuth(http.HandlerFunc(spacedesk.GetChatsByGroup)))
	
	mux.HandleFunc("/v1/ws/space-desk", spacedesk.SpaceDeskWebSocketHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv(utils.PORT)), middlewares.SecurityHeaders(middlewares.Cors(mux)))
}
