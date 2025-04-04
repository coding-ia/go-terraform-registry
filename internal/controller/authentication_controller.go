package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	registryauth "go-terraform-registry/internal/auth"
	registryconfig "go-terraform-registry/internal/config"
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
	Authorization(c *gin.Context)
	Callback(c *gin.Context)
	AccessToken(c *gin.Context)
}

func NewAuthenticationController(r *gin.Engine, config registryconfig.RegistryConfig) RegistryAuthenticationController {
	ac := &AuthenticationController{
		Config: config,
	}

	endpoint := github.Endpoint
	if config.OauthAuthURL != "" && config.OauthTokenURL != "" {
		endpoint = oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/login/oauth/authorize", config.OauthAuthURL),
			TokenURL: fmt.Sprintf("%s/login/oauth/access_token", config.OauthTokenURL),
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

	authentication := r.Group("/oauth")
	{
		authentication.GET("/authorization", ac.Authorization)
		authentication.GET("/callback", ac.Callback)
		authentication.POST("/token", ac.AccessToken)
	}

	return ac
}

func (a *AuthenticationController) Authorization(c *gin.Context) {
	state := c.Query("state")
	url := a.OauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	redirectURL := c.Query("redirect_uri")
	c.SetCookie("redirect-uri", redirectURL, 300, "/", "", true, true)

	c.Redirect(http.StatusFound, url)
}

func (a *AuthenticationController) Callback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	redirectUrl, err := c.Cookie("redirect-uri")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get redirect uri"})
		return
	}

	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectUrl, code, state)
	c.Redirect(http.StatusFound, redirectURL)
}

func (a *AuthenticationController) AccessToken(c *gin.Context) {
	code := c.PostForm("code")
	token, err := a.OauthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	userName, err := registryauth.GetGitHubUserName(c.Request.Context(), token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get GitHub user name"})
		return
	}

	accessToken, err := registryauth.CreateJWTToken(*userName, []byte(a.Config.TokenEncryptionKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}
