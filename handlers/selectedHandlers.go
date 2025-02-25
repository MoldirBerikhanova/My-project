package handlers

import (
	"goozinshe/models"
	"goozinshe/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SelectedlistHandler struct {
	moviesRepo       *repositories.MoviesRepository
	SelectedlistRepo *repositories.SelectedlistRepository
}

func NewSelectedlistHandler(moviesRepo *repositories.MoviesRepository, SelectedlistRepo *repositories.SelectedlistRepository) *SelectedlistHandler {
	return &SelectedlistHandler{moviesRepo: moviesRepo, SelectedlistRepo: SelectedlistRepo}
}

// HandleGetMovies godoc
// @Summary      получение списка проектов на главной
// @Tags 		 проекты на главную
// @Accept       json
// @Produce      json
// @Success      200 {array} models.Movie "OK"
// @Failure   	 500  {object} models.ApiError
// @Router       /selected [get]
func (h *SelectedlistHandler) HandleGetMoviesAndSeries(c *gin.Context) {
	movies, err := h.SelectedlistRepo.GetMoviesFromSelectedlist(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	response := map[string]interface{}{
		"movies": movies,
	}

	c.JSON(http.StatusOK, response)
}

// HandleAddMovie godoc
// @Summary      Добавление проектов на главную
// @Tags         проекты на главную
// @Accept       json
// @Produce      json
// @Param movieId path int true "Movie id"
// @Success      200  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid data"
// @Failure   	 500  {object} models.ApiError
// @Router       /selected/:movieId [post]
func (h *SelectedlistHandler) HandleAddMovie(c *gin.Context) {
	movieIdStr := c.Param("movieId")
	movieId, err := strconv.Atoi(movieIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie id"))
		return
	}

	_, err = h.moviesRepo.FindByIdAdmin(c, movieId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}
	err = h.SelectedlistRepo.AddToSelectedMovie(c, movieId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}
}

// HandleRemoveMovie godoc
// @Summary      Удаление проектов с главной
// @Tags         проекты на главную
// @Accept       json
// @Produce      json
// @Param movieId path int true "Movie id"
// @Success      200 "OK"
// @Failure   	 400  {object} models.ApiError "Invalid data"
// @Failure   	 500  {object} models.ApiError
// @Router       /selected/:movieId [delete]
func (h *SelectedlistHandler) HandleRemoveMovie(c *gin.Context) {
	movieIdStr := c.Param("movieId")
	movieId, err := strconv.Atoi(movieIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid movie id"))
		return
	}

	_, err = h.moviesRepo.FindByIdAdmin(c, movieId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	err = h.SelectedlistRepo.RemoveFromSelectedlist(c, movieId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
