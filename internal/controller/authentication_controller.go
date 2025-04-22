package controller

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	registryauth "go-terraform-registry/internal/auth"
	registryconfig "go-terraform-registry/internal/config"
	"go-terraform-registry/internal/githubclient"
	"go-terraform-registry/internal/response"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"log"
	"net/http"
)

type AuthenticationController struct {
	Config      registryconfig.RegistryConfig
	OauthConfig *oauth2.Config
}

type RegistryAuthenticationController interface {
	Authorization(http.ResponseWriter, *http.Request)
	Callback(http.ResponseWriter, *http.Request)
	AccessToken(http.ResponseWriter, *http.Request)
}

func NewAuthenticationController(router chi.Router, config registryconfig.RegistryConfig) RegistryAuthenticationController {
	ac := &AuthenticationController{
		Config: config,
	}

	endpoint := github.Endpoint
	if config.GitHubEndpoint != "" {
		endpoint = oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/login/oauth/authorize", config.GitHubEndpoint),
			TokenURL: fmt.Sprintf("%s/login/oauth/access_token", config.GitHubEndpoint),
		}
	}

	log.Printf("Authorization endpoint: %s", endpoint.AuthURL)
	log.Printf("Token endpoint: %s", endpoint.TokenURL)

	ac.OauthConfig = &oauth2.Config{
		ClientID:     config.OauthClientID,
		ClientSecret: config.OauthClientSecret,
		RedirectURL:  config.OauthClientRedirectURL,
		Scopes:       []string{"user"},
		Endpoint:     endpoint,
	}

	router.Route("/oauth", func(r chi.Router) {
		r.Get("/authorization", ac.Authorization)
		r.Get("/callback", ac.Callback)
		r.Post("/token", ac.AccessToken)
	})

	return ac
}

func (a *AuthenticationController) Authorization(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	url := a.OauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	redirectURL := r.URL.Query().Get("redirect_uri")
	http.SetCookie(w, &http.Cookie{
		Name:     "redirect-uri",
		Value:    redirectURL,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, url, http.StatusFound)
}

func (a *AuthenticationController) Callback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	redirectUrl, err := r.Cookie("redirect-uri")
	if err != nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to get redirect uri",
		})
		return
	}

	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectUrl, code, state)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (a *AuthenticationController) AccessToken(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		response.JsonResponse(w, http.StatusBadRequest, response.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	code := r.FormValue("code")
	token, err := a.OauthConfig.Exchange(r.Context(), code)
	if err != nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to exchange token",
		})
		return
	}

	client, err := githubclient.NewClient(r.Context(), token.AccessToken, a.Config.GitHubEndpoint)
	if err != nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "Unable to create GitHub client connection",
		})
		return
	}

	userName, err := registryauth.GetGitHubUserName(r.Context(), client, token.AccessToken)
	if err != nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to get GitHub user name",
		})
		return
	}

	accessToken, err := registryauth.CreateJWTToken(*userName, []byte(a.Config.TokenEncryptionKey))
	if err != nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create access token",
		})
		return
	}

	response.JsonResponse(w, http.StatusOK, response.AccessTokenResponse{
		Token: *accessToken,
	})
}
