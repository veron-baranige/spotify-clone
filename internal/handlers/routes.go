package handlers

import (
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo) {
	fileHandler := NewFileHandler()
	fileRoutes := e.Group("/api/v1/files")
	fileRoutes.POST("", fileHandler.SaveFile)
}
