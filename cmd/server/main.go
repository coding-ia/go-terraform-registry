package main

import "go-terraform-registry/internal/service"

var (
	version string = "dev"
)

func main() {
	service.StartServer(version)
}
