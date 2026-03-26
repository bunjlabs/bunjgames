package main

import (
	"bunjgames/hub"
	"bunjgames/server"
	"bunjgames/storage"
	"log"
	"os"
)

func main() {
	storage.CleanFolder("media")

	port := os.Getenv("PORT")
	hostStaticFiles := os.Getenv("HOST_STATIC_FILES") == "true"
	clientProxyUrl := os.Getenv("CLIENT_PROXY_URL")

	store := storage.NewGameStore()
	wsHub := hub.GetHub()

	server.StartGameLoop(store, wsHub)

	log.Printf("Server starting on :%s", port)
	server.StartServer(store, wsHub, port, hostStaticFiles, clientProxyUrl)
}
