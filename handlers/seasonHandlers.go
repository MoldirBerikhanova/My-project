package handlers

import (
	"goozinshe/models"
	"goozinshe/repositories"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SeasonsHandlers struct {
	SeasonsRepo   *repositories.SeasonRepository
	allseriesRepo *repositories.AllSeriesRepository
}

type createSeasonsRequest struct {
	Number *int    `json:"number"`
	Title  *string `json:"title"`
}

type updateSeasonsRequest struct {
	Number *int    `json:"number"`
	Title  *string `json:"title"`
}

func NewSeasonsHandlers(SeasonsRepo *repositories.SeasonRepository, allseriesRepo *repositories.AllSeriesRepository) *SeasonsHandlers {
	return &SeasonsHandlers{
		SeasonsRepo:   SeasonsRepo,
		allseriesRepo: allseriesRepo,
	}
}

func (h *SeasonsHandlers) AdminMiddleware(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewApiError("User not authenticated"))
		c.Abort()
		return
	}
	isAdmin, err := h.SeasonsRepo.IsAdmin(c, userId.(int))
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
// @Summary      Create Seasons
// @Tags 		 Seasons - это эндпоинты для каждого сезона
// @Accept       json
// @Produce      json
// @Param request body models.Season true "Seasons model"
// @Success      200  {object} object{id=int}  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid request Seasons"
// @Failure   	 500  {object} models.ApiError
// @Router       /admin/movies/seasons [post]
func (h *SeasonsHandlers) Create(c *gin.Context) {
	var request createSeasonsRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request Seasons")
		return
	}

	season := models.Season{
		Number: *request.Number,
		Title:  *request.Title,
	}

	id, err := h.SeasonsRepo.Create(c, season)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

// FindById godoc
// @Summary      Find by id Seasons
// @Tags         Seasons - это эндпоинты для каждого сезона
// @Accept       json
// @Produce      json
// @Param        id path int true "Seasons id"
// @Success      200  {object}  models.Season "Ok"
// @Failure      400  {object}  models.ApiError "Invalid Seasons id"
// @Router       /movies/seasons/{id} [get]
func (h *SeasonsHandlers) FindById(c *gin.Context) {
	idStr := c.Param("seasonId")
	seasonId, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Seasons id"))
		return
	}

	Seasons, err := h.SeasonsRepo.FindById(c, seasonId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, Seasons)

}

// FindAll godoc
// @Summary      Get all Seasons
// @Tags          Seasons - это эндпоинты для каждого сезона
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.Season "List of Seasons"
// @Failure      500  {object}  models.ApiError "Internal Server Error"
// @Router       /movies/seasons [get]
func (h *SeasonsHandlers) FindAll(c *gin.Context) {

	Seasons, err := h.SeasonsRepo.FindAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, Seasons)
}

// Update godoc
// @Summary      Update Seasons
// @Tags 		 Seasons - это эндпоинты для каждого сезона
// @Accept       json
// @Produce      json
// @Param request body models.Season true "Seasons model"
// @Success      200  {object} object{id=int}  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid Seasons Id"
// @Failure   	 500  {object} models.ApiError
// @Router       /admin/movies/seasons{id} [put]
func (h *SeasonsHandlers) Update(c *gin.Context) {
	log.Println("обновление")
	idStr := c.Param("seasonId")
	seasonId, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Seasons Id"))
		return
	}

	_, err = h.SeasonsRepo.FindById(c, seasonId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var request updateSeasonsRequest
	err = c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}
	season := models.Season{
		Number: *request.Number,
		Title:  *request.Title,
	}

	err = h.SeasonsRepo.Update(c, seasonId, season)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary      Delete Seasons
// @Tags          Seasons - это эндпоинты для каждого сезона
// @Accept       json
// @Produce      json
// @Param        id path int true "Seasons id"
// @Success      200  {object}  models.Season "Ok"
// @Failure      400  {object}  models.ApiError "Invalid Seasons Id"
// @Router       /admin/movies/seasons{id} [delete]
func (h *SeasonsHandlers) Delete(c *gin.Context) {
	idStr := c.Param("seasonId")
	seasonId, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Seasons Id"))
		return
	}

	_, err = h.SeasonsRepo.FindById(c, seasonId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.SeasonsRepo.Delete(c, seasonId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
