package main

import (
	"api/entities/budgets"
	"api/entities/clients"
	"api/entities/funnels"
	funnelshistory "api/entities/funnels_history"
	"api/entities/leads"
	"api/entities/orders"
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

	mux.Handle("GET /v1/space-desk", middlewares.LaravelAuth(http.HandlerFunc(leads.GetAll)))
	mux.Handle("POST /v1/space-desk", middlewares.LaravelAuth(http.HandlerFunc(leads.GetAll)))
	mux.HandleFunc("/v1/ws/space-desk", funnels.FunnelWebSocketHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv(utils.PORT)), middlewares.SecurityHeaders(middlewares.Cors(mux)))
}
