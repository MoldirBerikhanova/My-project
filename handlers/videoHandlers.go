package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type videoHandlers struct{}

func NewVideoHandlers() *videoHandlers {
	return &videoHandlers{}
}

// HandleGetVideoById godoc
// @Summary      Download video
// @Tags video
// @Accept       json
// @Produce      application/octet-stream
// @Param imageId path int true "video id"
// @Success      200  {string} string "video to download"
// @Failure 400 {object} models.ApiError "Invalid image id"
// @Failure   	 500  {object} models.ApiError
// @Router       /video/:videoId [get]
func (h *videoHandlers) HandleGetVideoById(c *gin.Context) {
	videoId := c.Param("videoId")
	if videoId == "" {
		c.JSON(http.StatusBadRequest, "Invalid video id")
		return
	}

	fileName := filepath.Base(videoId)
	byteFile, err := os.ReadFile(fmt.Sprintf("video/%s", videoId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ext := filepath.Ext(videoId)
	var contentType string
	switch ext {
	case ".mp4":
		contentType = "video/mp4"
	case ".avi":
		contentType = "video/x-msvideo"
	case ".mkv":
		contentType = "video/x-matroska"
	default:
		contentType = "application/octet-stream"
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Data(http.StatusOK, contentType, byteFile)
}
