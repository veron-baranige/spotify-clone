package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/veron-baranige/spotify-clone/internal/handlers"
)

func main() {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())

	handlers.SetupRoutes(e)

	log.Fatal(e.Start(":8080"))
}
