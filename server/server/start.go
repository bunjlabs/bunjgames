package server

import (
	"bunjgames/hub"
	"bunjgames/storage"
	"cmp"
	"context"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func StartServer(
	store *storage.GameStore,
	hub *hub.Hub,
	port string,
	hostStaticFiles bool,
	clientProxyUrl string,
) {
	echoServer := echo.New()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	echoServer.Logger = logger
	echoServer.Use(middleware.RequestLogger())
	echoServer.Use(middleware.Recover())

	if hostStaticFiles {
		echoServer.Logger.Info("Hosting media files server")
		echoServer.Static("/media/", "media")
	}

	if clientProxyUrl != "" {
		parsedClientProxyUrl, err := url.Parse(clientProxyUrl)
		if err != nil {
			echoServer.Logger.Error("Invalid client proxy URL", "url", clientProxyUrl)
			return
		}
		echoServer.Logger.Info("Using client proxy", "url", clientProxyUrl)
		echoServer.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
			Skipper: func(context *echo.Context) bool {
				return strings.HasPrefix(context.Path(), "/api/") ||
					strings.HasPrefix(context.Path(), "/ws/") ||
					strings.HasPrefix(context.Path(), "/media/")
			},
			Balancer: middleware.NewRandomBalancer(
				[]*middleware.ProxyTarget{
					{URL: parsedClientProxyUrl},
				},
			),
		}))
	}

	echoServer.POST("/api/create", func(context *echo.Context) error {
		return HandleGameCreation(context, store)
	})
	echoServer.POST("/api/register", func(context *echo.Context) error {
		return HandlePlayerRegistration(context, store, hub)
	})
	echoServer.GET("/ws/:token", func(context *echo.Context) error {
		return HandleGameConnection(
			context, store.Get(context.Param("token")), hub,
		)
	})

	sc := echo.StartConfig{Address: ":" + cmp.Or(port, "8000")}
	if err := sc.Start(context.Background(), echoServer); err != nil {
		echoServer.Logger.Error("failed to start server", "error", err)
	}
}
