package main

import (
	"bunjgames-server/common"
	"bunjgames-server/feud"
	"bunjgames-server/jeopardy"
	"bunjgames-server/weakest"
	"bunjgames-server/whirligig"
	"log"
	"net/http"
	"os"
)

func cleanMediaFolder() {
	log.Println("Cleaning media directory contents...")
	if err := os.MkdirAll("media", 0755); err != nil {
		log.Fatalf("Failed to create media directory: %v", err)
	}
	if entries, err := os.ReadDir("media"); err == nil {
		for _, entry := range entries {
			if entry.Name() == ".gitignore" {
				continue
			}
			path := "media/" + entry.Name()
			if err := os.RemoveAll(path); err != nil {
				log.Printf("Warning: Failed to remove %s: %v", path, err)
			}
		}
		log.Println("Media directory cleaned successfully")
	} else {
		log.Printf("Warning: Failed to read media directory: %v", err)
	}
}

func main() {
	cleanMediaFolder()

	hub := common.NewHub()
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/whirligig/create", whirligig.CreateHandler)
	mux.HandleFunc("POST /api/jeopardy/create", jeopardy.CreateHandler)
	mux.HandleFunc("POST /api/jeopardy/players/register", jeopardy.RegisterPlayerHandler)
	mux.HandleFunc("POST /api/weakest/create", weakest.CreateHandler)
	mux.HandleFunc("POST /api/weakest/players/register", weakest.RegisterPlayerHandler)
	mux.HandleFunc("POST /api/feud/create", feud.CreateHandler)
	mux.HandleFunc("POST /api/feud/players/register", feud.RegisterPlayerHandler)

	whirligigConsumer := whirligig.NewConsumer(hub)
	jeopardyConsumer := jeopardy.NewConsumer(hub)
	weakestConsumer := weakest.NewConsumer(hub)
	feudConsumer := feud.NewConsumer(hub)

	mux.Handle("/ws/whirligig/{token}", whirligigConsumer)
	mux.Handle("/ws/jeopardy/{token}", jeopardyConsumer)
	mux.Handle("/ws/weakest/{token}", weakestConsumer)
	mux.Handle("/ws/feud/{token}", feudConsumer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
