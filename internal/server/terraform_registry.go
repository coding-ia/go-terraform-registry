package server

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/config/selector"
	"go-terraform-registry/internal/controller"
	"go-terraform-registry/internal/storage"
	"log"
	"net/http"
	"os"
)

func StartServer(version string) {
	ctx := context.Background()

	log.SetOutput(os.Stdout)
	log.Println(fmt.Sprintf("Version: %s", version))

	cr := chi.NewRouter()
	cr.Use(middleware.Logger)

	// Get configuration and select backend
	c := config.GetRegistryConfig()
	b := selector.SelectBackend(ctx, c)

	err := b.Configure(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func(b *backend.Backend, ctx context.Context) {
		err := b.Close(ctx)
		if err != nil {
			log.Println(err)
		}
	}(b, ctx)

	// Configure storage
	s := selector.SelectStorage(ctx, c)
	if sae, ok := s.(storage.RegistryProviderStorageAssetEndpoint); ok {
		sae.ConfigureEndpoint(ctx, cr)
	}

	// Configure controllers
	_ = controller.NewServiceController(cr)
	_ = controller.NewProviderController(cr, c, *b, s)
	_ = controller.NewModuleController(cr, c, *b, s)
	_ = controller.NewAuthenticationController(cr, c)

	apiController := controller.NewAPIController(c, *b, s)
	apiController.CreateEndpoints(cr)

	err = http.ListenAndServe(":8080", cr)
	if err != nil {
		panic(err)
	}
}
