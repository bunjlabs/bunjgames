package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"bunjgames-server/common"
	"bunjgames-server/feud"
	"bunjgames-server/jeopardy"
	"bunjgames-server/weakest"
	"bunjgames-server/whirligig"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func availableGamesHandler(w http.ResponseWriter, r *http.Request) {
	token := strings.ToUpper(strings.TrimSpace(r.PathValue("token")))
	var result []map[string]string
	if feud.GameStore.Exists(token) {
		result = append(result, map[string]string{"type": "feud"})
	}
	if jeopardy.GameStore.Exists(token) {
		result = append(result, map[string]string{"type": "jeopardy"})
	}
	if weakest.GameStore.Exists(token) {
		result = append(result, map[string]string{"type": "weakest"})
	}
	if result == nil {
		result = []map[string]string{}
	}
	common.JSONResponse(w, result)
}

func main() {
	hub := common.NewHub()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/common/games/{token}/available", availableGamesHandler)

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

	os.MkdirAll("media", 0755)
	mux.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("media"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}
