package main

import (
	"api/source/entities/funnels"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/funnels", funnels.CreateOne)
	mux.HandleFunc("GET /v1/funnels/{id}", funnels.CreateOne)
	mux.HandleFunc("POST /v1/funnels", funnels.CreateOne)
	mux.HandleFunc("PATCH /v1/funnels/{id}", funnels.UpdateOne)
	mux.HandleFunc("DELETE /v1/funnels/{id}", funnels.DeleteOne)

	mux.HandleFunc("GET /v1/leads", funnels.CreateOne)
	mux.HandleFunc("GET /v1/leads/{id}", funnels.CreateOne)
	mux.HandleFunc("POST /v1/funnels", funnels.CreateOne)
	mux.HandleFunc("PATCH /v1/funnels/{id}", funnels.UpdateOne)
	mux.HandleFunc("DELETE /v1/funnels/{id}", funnels.DeleteOne)

	http.ListenAndServe(":8080", mux)
}
