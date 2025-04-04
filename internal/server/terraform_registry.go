package server

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/config/selector"
	"go-terraform-registry/internal/controller"
	"go-terraform-registry/internal/storage"
	"log"
	"os"
)

var (
	ginLambda *ginadapter.GinLambda
)

func StartServer(version string) {
	ctx := context.Background()

	if version != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	gin.DefaultWriter = log.Writer()
	r := gin.Default()
	r.Use(gin.LoggerWithWriter(log.Writer()))

	// Get configuration and select backend
	c := config.GetRegistryConfig()
	b := selector.SelectBackend(ctx, c)

	// Configure storage
	s := selector.SelectStorage(ctx, c)
	if sae, ok := s.(storage.RegistryProviderStorageAssetEndpoint); ok {
		assetEndpoint := r.Group("/asset")
		sae.ConfigureEndpoint(ctx, assetEndpoint)
	}

	// Configure controllers
	_ = controller.NewServiceController(r)
	_ = controller.NewProviderController(r, c, b, s)
	_ = controller.NewModuleController(r, c, b)
	_ = controller.NewAuthenticationController(r, c)
	apiController := controller.NewAPIController(c, b, s)

	apiController.CreateEndpoints(r)

	lambdaFunction := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")

	if lambdaFunction == "" {
		err := r.SetTrustedProxies(nil)
		if err != nil {
			panic(err)
		}
		err = r.Run()
		if err != nil {
			panic(err)
		}
	} else {
		ginLambda = ginadapter.New(r)
		lambda.Start(Handler)
	}
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}
