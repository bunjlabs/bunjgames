package main

import (
	"bunjgames/storage"
	"log"
	"os"
)

func main() {
	storage.CleanFolder("media")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	hostStaticFiles := os.Getenv("HOST_STATIC_FILES") == "true"
	clientProxyUrl := os.Getenv("CLIENT_PROXY_URL")

	log.Printf("Server starting on :%s", port)
	StartServer(port, hostStaticFiles, clientProxyUrl)
}
