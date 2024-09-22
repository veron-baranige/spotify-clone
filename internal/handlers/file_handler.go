package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/veron-baranige/spotify-clone/internal/services"
)

type FileHandler struct {
	fileService *services.FileService
}

func NewFileHandler() *FileHandler {
	return &FileHandler{services.NewFileService()}
}

func (fh *FileHandler) SaveFile(c echo.Context) error {
	err := fh.fileService.SaveFile(c)

	if err != nil {
		log.Fatal(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusCreated)
}
