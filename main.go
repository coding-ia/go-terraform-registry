package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/controller"
	"os"
)

var ginLambda *ginadapter.GinLambda

func main() {
	ctx := context.Background()

	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	c := config.GetRegistryConfig()
	b := config.SelectBackend(ctx, "dynamodb")

	_ = controller.NewServiceController(r)
	_ = controller.NewProviderController(r, c, b)
	_ = controller.NewModuleController(r, c, b)
	_ = controller.NewAuthenticationController(r, c)

	lambdaFunction := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")

	if lambdaFunction == "" {
		err := r.SetTrustedProxies(nil)
		if err != nil {
			panic(err)
		}
		err = r.Run(":8080")
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
