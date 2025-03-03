package handlers

import (
	"fmt"
	"goozinshe/models"
	"goozinshe/repositories"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GenreHandlers struct {
	repo *repositories.GenresRepository
}

type createGenreRequest struct {
	Title  string                `form:"title"`
	Poster *multipart.FileHeader `form:"poster"`
}

type updateGenreRequest struct {
	Title  string                `form:"title"`
	Poster *multipart.FileHeader `form:"poster"`
}

func NewGenreHanlers(repo *repositories.GenresRepository) *GenreHandlers {
	return &GenreHandlers{
		repo: repo,
	}
}
func (h *GenreHandlers) AdminMiddleware(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewApiError("User not authenticated"))
		c.Abort()
		return
	}
	isAdmin, err := h.repo.IsAdmin(c, userId.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not check user role"))
		c.Abort()
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, models.NewApiError("Access forbidden"))
		c.Abort()
		return
	}

	c.Next()
}

// FindById godoc
// @Summary      Find by id
// @Tags         genres
// @Accept       json
// @Produce      json
// @Param        id path int true "Genre id"
// @Success      200  {object}  models.Genre "Ok"
// @Failure      400  {object}  models.ApiError "Invalid Movie Id"
// @Router       /admin/genres/{id} [get]
func (h *GenreHandlers) FindById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Genre Id"))
		return
	}

	genre, err := h.repo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, genre)
}

// FindAll godoc
// @Summary      Get all genres
// @Tags         genres
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.Genre "List of genres"
// @Failure      500  {object}  models.ApiError "Internal Server Error"
// @Router       /genres [get]
func (h *GenreHandlers) FindAll(c *gin.Context) {
	genres, err := h.repo.FindAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, genres)
}

// Create godoc
// @Summary      Create genre
// @Tags 		 genres
// @Accept       json
// @Produce      json
// @Param request body models.Genre true "Genre model"
// @Success      200  {object} object{id=int}  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid request category"
// @Failure   	 500  {object} models.ApiError
// @Router        /admin/genres [post]
func (h *GenreHandlers) Create(c *gin.Context) {
	var request createGenreRequest
	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}

	if request.Poster == nil {
		c.JSON(http.StatusBadRequest, "Poster file is required")
		return
	}

	filename, err := h.saveGenrePoster(c, request.Poster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	genre := models.Genre{
		Title:     request.Title,
		PosterUrl: filename,
	}

	id, err := h.repo.Create(c, genre)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (h *GenreHandlers) saveGenrePoster(c *gin.Context, poster *multipart.FileHeader) (string, error) {
	filename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(poster.Filename))
	filepath := fmt.Sprintf("images/%s", filename)
	err := c.SaveUploadedFile(poster, filepath)

	return filename, err
}

// Update godoc
// @Summary      Update genre
// @Tags 		 genres
// @Accept       json
// @Produce      json
// @Param request body models.Genre true "Genre model"
// @Success      200  {object} object{id=int}  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid Genre Id"
// @Failure   	 500  {object} models.ApiError
// @Router        /admin/genres/{id} [put]
func (h *GenreHandlers) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Genre Id"))
		return
	}

	_, err = h.repo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var request updateGenreRequest
	err = c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}

	filename, err := h.saveGenrePoster(c, request.Poster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	genre := models.Genre{
		Title:     request.Title,
		PosterUrl: filename,
	}

	h.repo.Update(c, id, genre)

	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary      Delete genre
// @Tags         genres
// @Accept       json
// @Produce      json
// @Param        id path int true "Genre id"
// @Success      200  {object}  models.Genre "Ok"
// @Failure      400  {object}  models.ApiError "Invalid genre Id"
// @Router        /genres/{id} [delete]
func (h *GenreHandlers) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Genre Id"))
		return
	}

	_, err = h.repo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.repo.Delete(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
