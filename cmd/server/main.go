package main

import "go-terraform-registry/internal/server"

var (
	version string = "dev"
)

func main() {
	server.StartServer(version)
}
