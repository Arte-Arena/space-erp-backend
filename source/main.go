package main

import (
	"api/source/entities/funnels"
	"api/source/entities/leads"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/funnels", funnels.GetOne)
	mux.HandleFunc("GET /v1/funnels/{id}", funnels.GetAll)
	mux.HandleFunc("POST /v1/funnels", funnels.CreateOne)
	mux.HandleFunc("PATCH /v1/funnels/{id}", funnels.UpdateOne)
	mux.HandleFunc("DELETE /v1/funnels/{id}", funnels.DeleteOne)

	mux.HandleFunc("GET /v1/leads", leads.GetAll)
	mux.HandleFunc("GET /v1/leads/{id}", leads.GetOne)
	mux.HandleFunc("POST /v1/leads", leads.CreateOne)
	mux.HandleFunc("PATCH /v1/leads/{id}", leads.UpdateOne)

	mux.HandleFunc("GET /v1/reports", leads.GetAll)

	http.ListenAndServe(":8080", mux)
}
