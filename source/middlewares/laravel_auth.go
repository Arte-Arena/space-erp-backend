package middlewares

import (
	"api/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type contextKey string

const UserContextKey = contextKey("laravel_user")

type LaravelUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func LaravelAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			utils.SendResponse(w, http.StatusUnauthorized, "Token não informado", nil, 0)
			return
		}

		laravelURL := os.Getenv("LARAVEL_API_URL")
		if laravelURL == "" {
			laravelURL = "http://localhost:8000"
		}
		userURL := fmt.Sprintf("%s/api/user", laravelURL)

		req, err := http.NewRequest("GET", userURL, nil)
		if err != nil {
			utils.SendResponse(w, http.StatusInternalServerError, "Erro ao criar requisição de autenticação", nil, 0)
			return
		}
		req.Header.Set("Authorization", token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			utils.SendResponse(w, http.StatusBadGateway, "Erro ao conectar na API de autenticação", nil, 0)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			utils.SendResponse(w, http.StatusUnauthorized, "Token inválido ou usuário não autenticado", nil, 0)
			return
		}

		user := LaravelUser{}
		err = json.NewDecoder(resp.Body).Decode(&user)
		if err != nil || user.ID == 0 || user.Name == "" || user.Email == "" {
			utils.SendResponse(w, http.StatusUnauthorized, "Usuário inválido retornado pela autenticação", nil, 0)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
