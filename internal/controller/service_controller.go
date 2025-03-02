package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ServiceController struct {
}

type RegistryServiceController interface {
	ServiceDiscovery(*gin.Context)
}

func NewServiceController(r *gin.Engine) RegistryServiceController {
	sc := &ServiceController{}

	r.GET(".well-known/terraform.json", sc.ServiceDiscovery)

	return sc
}

func (s ServiceController) ServiceDiscovery(c *gin.Context) {
	serviceData := `
{
	"providers.v1": "/terraform/providers/v1/",
	"modules.v1": "/terraform/modules/v1/",
	"login.v1": {
		"client": "terraform-cli",
		"grant_types": [
			"authz_code"
		],
		"authz": "/oauth/authorization",
		"token": "/oauth/token"
	}
}
`

	c.Data(http.StatusOK, "application/json", []byte(serviceData))
}
