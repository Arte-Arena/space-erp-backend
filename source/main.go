package main

import (
	"fmt"
	"net/http"
)

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Fprintf(w, "Usuário ID: %s\n", id)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/leads/{id}", getUserHandler)
	mux.HandleFunc("GET /v1/funnels/{id}", getUserHandler)

	http.ListenAndServe(":8080", mux)
}
