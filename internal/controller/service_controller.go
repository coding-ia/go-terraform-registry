package controller

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type ServiceController struct {
}

type RegistryServiceController interface {
	ServiceDiscovery(http.ResponseWriter, *http.Request)
}

func NewServiceController(r chi.Router) RegistryServiceController {
	sc := &ServiceController{}

	r.Get("/.well-known/terraform.json", sc.ServiceDiscovery)

	return sc
}

func (s ServiceController) ServiceDiscovery(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(serviceData))
	if err != nil {
		log.Println("ERROR: ServiceDiscovery", err)
	}
}
