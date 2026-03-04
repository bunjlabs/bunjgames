package main

import (
	"bunjgames/game"
	"bunjgames/storage"
	"cmp"
	"context"
	"net/url"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func StartServer(
	port string,
	hostStaticFiles bool,
	clientProxyUrl string,
) {
	store := storage.NewGameStore()

	server := echo.New()
	server.Use(middleware.RequestLogger())
	server.Use(middleware.Recover())

	if hostStaticFiles {
		server.Static("/media", "media")
	}

	if clientProxyUrl != "" {
		parsedClientProxyUrl, err := url.Parse(clientProxyUrl)
		if err != nil {
			server.Logger.Error("Invalid client proxy URL", "url", clientProxyUrl)
			return
		}
		server.Logger.Info("Using client proxy", "url", clientProxyUrl)
		server.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
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

	server.POST("/api/create", func(context *echo.Context) error {
		return game.HandleGameCreation(context, store)
	})
	server.GET("/ws/{token}", func(context *echo.Context) error {
		return game.HandleGameConnection(
			context, store.Get(context.Param("token")),
		)
	})

	sc := echo.StartConfig{Address: ":" + cmp.Or(port, "8000")}
	if err := sc.Start(context.Background(), server); err != nil {
		server.Logger.Error("failed to start server", "error", err)
	}
}
