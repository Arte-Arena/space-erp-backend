package main

import (
	"api/source/entities/budgets"
	"api/source/entities/clients"
	"api/source/entities/funnels"
	"api/source/entities/leads"
	"api/source/middlewares"
	"api/source/utils"
	"fmt"
	"net/http"
	"os"
)

func main() {
	utils.LoadEnvVariables()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/funnels", funnels.GetAll)
	mux.HandleFunc("GET /v1/funnels/{id}", funnels.GetOne)
	mux.HandleFunc("POST /v1/funnels", funnels.CreateOne)
	mux.HandleFunc("PATCH /v1/funnels/{id}", funnels.UpdateOne)
	mux.HandleFunc("DELETE /v1/funnels/{id}", funnels.DeleteOne)

	mux.HandleFunc("GET /v1/leads", leads.GetAll)
	mux.HandleFunc("GET /v1/leads/{id}", leads.GetOne)
	mux.HandleFunc("POST /v1/leads", leads.CreateOne)
	mux.HandleFunc("PATCH /v1/leads/{id}", leads.UpdateOne)

	mux.HandleFunc("GET /v1/clients", clients.GetAll)
	mux.HandleFunc("GET /v1/clients/{id}", clients.GetOne)

	mux.HandleFunc("GET /v1/budgets", budgets.GetAll)
	mux.HandleFunc("GET /v1/budgets/{id}", budgets.GetOne)

	mux.HandleFunc("GET /v1/reports", leads.GetAll)

	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv(utils.PORT)), middlewares.SecurityHeaders(middlewares.Cors(mux)))
}
