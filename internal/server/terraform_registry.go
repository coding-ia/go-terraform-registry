package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"go-terraform-registry/internal/backend"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/config/selector"
	"go-terraform-registry/internal/controller"
	"go-terraform-registry/internal/storage"
	"log"
)

func StartServer(version string) {
	ctx := context.Background()

	cr := chi.NewRouter()

	if version != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DefaultWriter = log.Writer()
	r := gin.Default()
	r.Use(gin.LoggerWithWriter(log.Writer()))

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
		endpoint := r.Group("/asset")
		endpoint.Any("*any", gin.WrapH(cr))
	}

	// Configure controllers
	_ = controller.NewServiceController(r)
	_ = controller.NewProviderController(r, c, *b, s)
	_ = controller.NewModuleController(r, c, *b, s)
	_ = controller.NewAuthenticationController(r, c)
	apiController := controller.NewAPIController(c, *b, s, cr)

	apiController.CreateEndpoints(r, cr)

	err = r.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}
	err = r.Run()
	if err != nil {
		panic(err)
	}
}
