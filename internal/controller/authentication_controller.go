package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	registryconfig "go-terraform-registry/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
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

	ac.OauthConfig = &oauth2.Config{
		ClientID:     config.OauthClientID,
		ClientSecret: config.OauthClientSecret,
		RedirectURL:  "https://6850-2601-201-8481-34d0-e544-bbdb-ab20-fddb.ngrok-free.app/oauth/callback",
		Scopes:       []string{"user"},
		Endpoint:     github.Endpoint,
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
	c.Redirect(http.StatusFound, url)
}

func (a *AuthenticationController) Callback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	redirectURL := fmt.Sprintf("http://localhost:18523/login?code=%s&state=%s", code, state)
	c.Redirect(http.StatusFound, redirectURL)
}

func (a *AuthenticationController) AccessToken(c *gin.Context) {
	code := c.PostForm("code")
	token, err := a.OauthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": token.AccessToken})
}
