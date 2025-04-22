package controller

import (
	"context"
	"github.com/go-chi/chi/v5"
	"go-terraform-registry/internal/api"
	"go-terraform-registry/internal/auth"
	"go-terraform-registry/internal/backend"
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/response"
	"go-terraform-registry/internal/storage"
	"net/http"
	"strings"
)

type APIController struct {
	Config  registryconfig.RegistryConfig
	Backend backend.Backend
	Storage storage.RegistryProviderStorage
	Chi     *chi.Mux
}

type RegistryAPIController interface {
	CreateEndpoints(cr *chi.Mux)
	AuthenticateRequestMiddleware(next http.Handler) http.Handler
}

func NewAPIController(config registryconfig.RegistryConfig, backend backend.Backend, storage storage.RegistryProviderStorage) RegistryAPIController {
	ac := &APIController{
		Config:  config,
		Backend: backend,
		Storage: storage,
	}

	return ac
}

func (a *APIController) CreateEndpoints(cr *chi.Mux) {
	cr.Route("/api", func(r chi.Router) {
		r.Use(a.AuthenticateRequestMiddleware)

		providerVersionsAPI := api.ProviderVersionsAPI{
			Config:  a.Config,
			Backend: a.Backend,
			Storage: a.Storage,
		}
		r.With(ValidateOrganizationMiddleware).Post("/v2/organizations/{organization}/registry-providers/{registry}/{namespace}/{name}/versions", providerVersionsAPI.CreateVersion)
		r.With(ValidateOrganizationMiddleware).Get("/v2/organizations/{organization}/registry-providers/{registry}/{namespace}/{name}/versions/", providerVersionsAPI.ListVersions)
		r.With(ValidateOrganizationMiddleware).Get("/v2/organizations/{organization}/registry-providers/{registry}/{namespace}/{name}/versions/{version}", providerVersionsAPI.GetVersion)
		r.With(ValidateOrganizationMiddleware).Delete("/v2/organizations/{organization}/registry-providers/{registry}/{namespace}/{name}/versions/{version}", providerVersionsAPI.DeleteVersion)
		r.With(ValidateOrganizationMiddleware).Post("/v2/organizations/{organization}/registry-providers/{registry}/{namespace}/{name}/versions/{version}/platforms", providerVersionsAPI.CreatePlatform)
		r.With(ValidateOrganizationMiddleware).Get("/v2/organizations/{organization}/registry-providers/{registry}/{namespace}/{name}/versions/{version}/platforms", providerVersionsAPI.ListPlatform)
		r.With(ValidateOrganizationMiddleware).Get("/v2/organizations/{organization}/registry-providers/{registry}/{namespace}/{name}/versions/{version}/platforms/{os}/{arch}", providerVersionsAPI.GetPlatform)
		r.With(ValidateOrganizationMiddleware).Delete("/v2/organizations/{organization}/registry-providers/{registry}/{namespace}/{name}/versions/{version}/platforms/{os}/{arch}", providerVersionsAPI.DeletePlatform)

		providersAPI := api.ProvidersAPI{
			Config:  a.Config,
			Backend: a.Backend,
			Storage: a.Storage,
		}
		r.With(ValidateOrganizationMiddleware).Get("/v2/organizations/{organization}/registry-providers", providersAPI.List)
		r.With(ValidateOrganizationMiddleware).Post("/v2/organizations/{organization}/registry-providers", providersAPI.Create)
		r.With(ValidateOrganizationMiddleware).Get("/v2/organizations/{organization}/registry-providers/{registry}/{namespace}/{name}", providersAPI.Get)
		r.With(ValidateOrganizationMiddleware).Delete("/v2/organizations/{organization}/registry-providers/{registry}/{namespace}/{name}", providersAPI.Delete)

		modulesAPI := api.ModulesAPI{
			Config:  a.Config,
			Backend: a.Backend,
			Storage: a.Storage,
		}
		r.With(ValidateOrganizationMiddleware).Post("/v2/organizations/{organization}/registry-modules", modulesAPI.Create)
		r.With(ValidateOrganizationMiddleware).Get("/v2/organizations/{organization}/registry-modules/{registry}/{namespace}/{name}/{provider}", modulesAPI.Get)

		moduleVersionsAPI := api.ModuleVersionsAPI{
			Config:  a.Config,
			Backend: a.Backend,
			Storage: a.Storage,
		}
		r.With(ValidateOrganizationMiddleware).Post("/v2/organizations/{organization}/registry-modules/{registry}/{namespace}/{name}/{provider}/versions", moduleVersionsAPI.Create)
		r.With(ValidateOrganizationMiddleware).Delete("/v2/organizations/{organization}/registry-modules/{registry}/{namespace}/{name}/{provider}/{version}", moduleVersionsAPI.Delete)

		gpgKeysAPI := api.GPGKeysAPI{
			Config:  a.Config,
			Backend: a.Backend,
			Storage: a.Storage,
		}
		r.Get("/registry/{registry}/v2/gpg-keys", gpgKeysAPI.List)
		r.Post("/registry/{registry}/v2/gpg-keys", gpgKeysAPI.Add)
		r.Get("/registry/{registry}/v2/gpg-keys/{namespace}/{key_id}", gpgKeysAPI.Get)
		r.Patch("/registry/{registry}/v2/gpg-keys/{namespace}/{key_id}", gpgKeysAPI.Update)
		r.Delete("/registry/{registry}/v2/gpg-keys/{namespace}/{key_id}", gpgKeysAPI.Delete)
	})
}

func (a *APIController) AuthenticateRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.JsonResponse(w, http.StatusUnauthorized, response.ErrorResponse{
				Error: "Missing Authorization header",
			})
			return
		}

		const prefix = "Bearer "
		if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
			response.JsonResponse(w, http.StatusUnauthorized, response.ErrorResponse{
				Error: "Invalid Authorization header format",
			})
			return
		}

		tokenString := authHeader[len(prefix):]
		token, err := auth.GetJWTClaimsToken(tokenString, []byte(a.Config.TokenEncryptionKey))
		if err != nil {
			response.JsonResponse(w, http.StatusUnauthorized, response.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		if token != nil && token.Valid {
			if claims, ok := token.Claims.(*auth.RegistryClaims); ok {
				ctx := context.WithValue(r.Context(), "organization", claims.Organization)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		}
	})
}

func ValidateOrganizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organizationParam := chi.URLParam(r, "organization")
		orgVal := r.Context().Value("organization")

		if orgVal == nil {
			response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
				Error: "Missing organization",
			})
			return
		}

		if !strings.EqualFold(orgVal.(string), organizationParam) {
			response.JsonResponse(w, http.StatusUnprocessableEntity, response.ErrorResponse{
				Error: "Invalid token for organization",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
