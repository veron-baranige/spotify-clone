package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/labstack/echo/v4"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	"github.com/u2takey/go-utils/uuid"
)

type FileService struct{}

func NewFileService() *FileService {
	return &FileService{}
}

const (
	outputDirPath    = "./output"
	rawFileName      = "raw"
	audioSegmentTime = 10
)

func (fs *FileService) SaveFile(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	// TODO: validate file type (audio/mp3) and size
	// contentType := http.DetectContentType(buffer)

	fileId := uuid.NewUUID()
	if err := os.MkdirAll(outputDirPath+"/"+fileId, 0775); err != nil {
		return err
	}

	fileExt := filepath.Ext(file.Filename)
	filePath := fmt.Sprintf("%s/%s/%s", outputDirPath, fileId, rawFileName+fileExt)
	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	bitrates := [3]uint16{160, 256, 320}
	errCh := make(chan error, len(bitrates))
	var wg sync.WaitGroup

	for _, bitrate := range bitrates {
		wg.Add(1)
		go func(bitrate uint16) {
			defer wg.Done()
			fs.processAudioFile(fileId, fileExt, bitrate, errCh)
		}(bitrate)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	// TODO: upload to bucket

	go fs.cleanUpProcessedFiles(fileId)

	return nil
}

func (fs *FileService) processAudioFile(fileId string, fileExt string, bitrate uint16, errCh chan<- error) {
	fileDirPath := fmt.Sprintf("%s/%s/%d", outputDirPath, fileId, bitrate)
	err := os.MkdirAll(fileDirPath, 0755)
	if err != nil {
		errCh <- err
		return
	}

	err = ffmpeg_go.Input(outputDirPath+"/"+fileId+"/"+rawFileName+fileExt).
		Output(fmt.Sprintf("%s/%s/%d/output%%03d.ts", outputDirPath, fileId, bitrate), ffmpeg_go.KwArgs{
			"c:a":            "libmp3lame",
			"b:a":            fmt.Sprintf("%dk", bitrate),
			"map":            "0:0",
			"f":              "segment",
			"segment_time":   audioSegmentTime,
			"segment_list":   fmt.Sprintf("%s/%s/outputlist.m3u8", outputDirPath, fileId),
			"segment_format": "mpegts",
		}).Run()
	errCh <- err
}

func (fs *FileService) cleanUpProcessedFiles(fileId string) {
	err := os.RemoveAll(outputDirPath + "/" + fileId)
	if err != nil {
		log.Printf("Error cleaning up processed files: %v", err)
	}
}
