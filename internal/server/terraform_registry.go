package server

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/controller"
	"go-terraform-registry/internal/storage/s3_storage"
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

	r := gin.Default()

	c := config.GetRegistryConfig()
	b := config.SelectBackend(ctx, c.Backend)
	s := s3_storage.NewS3Storage(c)

	_ = controller.NewServiceController(r)
	_ = controller.NewProviderController(r, c, b)
	_ = controller.NewModuleController(r, c, b)
	_ = controller.NewAuthenticationController(r, c)
	_ = controller.NewAPIController(r, c, b, s)

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
