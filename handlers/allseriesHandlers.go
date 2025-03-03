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

type AllSeriesHandlers struct {
	allseriesRepo *repositories.AllSeriesRepository
}

type createAllSeriesRequest struct {
	Series     *int                  `form:"series"`
	Title      *string               `form:"title"`
	TrailerUrl *string               `form:"trailer_url"`
	Duration   *string               `form:"duration"`
	PosterUrl  *multipart.FileHeader `form:"poster_url"`
}

type updateAllSeriesRequest struct {
	Series     *int                  `form:"series"`
	Title      *string               `form:"title"`
	TrailerUrl *string               `form:"trailer_url"`
	Duration   *string               `form:"duration"`
	PosterUrl  *multipart.FileHeader `form:"poster_url"`
}

func NewAllSeriesHandlers(allseriesRepo *repositories.AllSeriesRepository) *AllSeriesHandlers {
	return &AllSeriesHandlers{
		allseriesRepo: allseriesRepo,
	}
}

func (h *AllSeriesHandlers) AdminMiddleware(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewApiError("User not authenticated"))
		c.Abort()
		return
	}
	isAdmin, err := h.allseriesRepo.IsAdmin(c, userId.(int))
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

// Create godoc
// @Summary      Create allseries
// @Tags 		 allseries - это эндпоинты для каждой серии
// @Accept       json
// @Produce      json
// @Param request body models.AllSeries true "AllSeries model"
// @Success      200  {object} object{id=int}  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid request AllSeries"
// @Failure   	 500  {object} models.ApiError
// @Router       /admin/movies/allseries [post]
func (h *AllSeriesHandlers) Create(c *gin.Context) {
	var request createAllSeriesRequest
	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request AllSeries")
		return
	}

	allserie := models.AllSeries{
		Series: request.Series,
		Title:  request.Title,
		// Description: request.Description,
		// ReleaseYear: request.ReleaseYear,
		// Director:    request.Director,
		// Rating:      request.Rating,
		TrailerUrl: request.TrailerUrl,
		Duration:   request.Duration,
	}

	id, err := h.allseriesRepo.Create(c, allserie)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

// FindById godoc
// @Summary      Find by id allseries
// @Tags         allseries - это эндпоинты для каждой серии
// @Accept       json
// @Produce      json
// @Param        id path int true "AllSeries id"
// @Success      200  {object}  models.AllSeries "Ok"
// @Failure      400  {object}  models.ApiError "Invalid allseries id"
// @Router       /movies/allseries/{id} [get]
func (h *AllSeriesHandlers) FindById(c *gin.Context) {
	idStr := c.Param("movieId")
	movieId, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid allseries id"))
		return
	}

	allseries, err := h.allseriesRepo.FindById(c, movieId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, allseries)

}

// FindAll godoc
// @Summary      Get all allseries
// @Tags          allseries - это эндпоинты для каждой серии
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.AllSeries "List of allseries"
// @Failure      500  {object}  models.ApiError "Internal Server Error"
// @Router       /movies/allseries [get]
func (h *AllSeriesHandlers) FindAll(c *gin.Context) {

	allseries, err := h.allseriesRepo.FindAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, allseries)
}

func (h *AllSeriesHandlers) saveMoviesPoster(c *gin.Context, poster *multipart.FileHeader) (*string, error) {
	filename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(poster.Filename))
	filepath := fmt.Sprintf("images/%s", filename)
	err := c.SaveUploadedFile(poster, filepath)

	return &filename, err
}

// Update godoc
// @Summary      Update allseries
// @Tags 		 allseries - это эндпоинты для каждой серии
// @Accept       json
// @Produce      json
// @Param request body models.AllSeries true "AllSeries model"
// @Success      200  {object} object{id=int}  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid AllSeries Id"
// @Failure   	 500  {object} models.ApiError
// @Router       /admin/movies/allseries/{id} [put]
func (h *AllSeriesHandlers) Update(c *gin.Context) {
	idStr := c.Param("movieId")
	movieId, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid AllSeries Id"))
		return
	}

	_, err = h.allseriesRepo.FindById(c, movieId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var request updateAllSeriesRequest
	err = c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}

	filename, err := h.saveMoviesPoster(c, request.PosterUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}
	allserie := models.AllSeries{
		Series:     request.Series,
		Title:      request.Title,
		TrailerUrl: request.TrailerUrl,
		Duration:   request.Duration,
		PosterUrl:  filename,
	}

	err = h.allseriesRepo.Update(c, movieId, allserie)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary      Delete allseries
// @Tags          allseries - это эндпоинты для каждой серии
// @Accept       json
// @Produce      json
// @Param        id path int true "Allseries id"
// @Success      200  {object}  models.AllSeries "Ok"
// @Failure      400  {object}  models.ApiError "Invalid AllSeries Id"
// @Router       /admin/movies/allseries/{id} [delete]
func (h *AllSeriesHandlers) Delete(c *gin.Context) {
	idStr := c.Param("movieId")
	movieId, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid AllSeries Id"))
		return
	}

	_, err = h.allseriesRepo.FindById(c, movieId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.allseriesRepo.Delete(c, movieId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
