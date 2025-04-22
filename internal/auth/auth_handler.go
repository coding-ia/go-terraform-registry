package auth

import (
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/response"
	"net/http"
	"strings"
)

type Authentication struct {
	Config registryconfig.RegistryConfig
}

type AuthenticationMiddleware interface {
	AuthenticationHandlerMiddleware(http.Handler) http.Handler
}

func NewAuthenticationMiddleware(config registryconfig.RegistryConfig) AuthenticationMiddleware {
	return &Authentication{
		Config: config,
	}
}

func (a *Authentication) AuthenticationHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.JsonResponse(w, http.StatusUnauthorized, response.ErrorResponse{
				Error: "Authorization header missing",
			})
			return
		}

		parts := strings.Fields(authHeader)
		token, err := GetJWTToken(parts[1], []byte(a.Config.TokenEncryptionKey))
		if err != nil {
			response.JsonResponse(w, http.StatusUnauthorized, response.ErrorResponse{
				Error: "Error parsing token",
			})
			return
		}

		if !token.Valid {
			response.JsonResponse(w, http.StatusUnauthorized, response.ErrorResponse{
				Error: "Invalid token",
			})
			return
		}
	})
}
