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

type AgeHandler struct {
	ageRepo *repositories.AgeRepository
}

type createAgeRequest struct {
	Age    string                `form:"age"`
	Poster *multipart.FileHeader `form:"poster"`
}

type updateAgeRequest struct {
	Age    string                `form:"age"`
	Poster *multipart.FileHeader `form:"poster"`
}

func NewAgeHandler(ageRepo *repositories.AgeRepository) *AgeHandler {
	return &AgeHandler{
		ageRepo: ageRepo,
	}
}

func (a *AgeHandler) AdminMiddleware(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewApiError("User not authenticated"))
		c.Abort()
		return
	}
	isAdmin, err := a.ageRepo.IsAdmin(c, userId.(int))
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

// HandleAddAge godoc
// @Summary      HandleAddAge age
// @Tags 		 ages
// @Accept       json
// @Produce      json
// @Param request body models.Age true "Age model"
// @Success      200  {object} object{id=int}  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid request age"
// @Failure   	 500  {object} models.ApiError
// @Router       /admin/ages [post]
func (a *AgeHandler) HandleAddAge(c *gin.Context) {
	var request createAgeRequest
	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("invalid age"))
		return
	}

	if request.Poster == nil {
		c.JSON(http.StatusBadRequest, "Poster file is required")
		return
	}

	filename, err := a.saveAgePoster(c, request.Poster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	age := models.Age{
		Age:       request.Age,
		PosterUrl: filename,
	}

	id, err := a.ageRepo.Create(c, age)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (a *AgeHandler) saveAgePoster(c *gin.Context, poster *multipart.FileHeader) (string, error) {
	filename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(poster.Filename))
	filepath := fmt.Sprintf("images/%s", filename)
	err := c.SaveUploadedFile(poster, filepath)

	return filename, err
}

// FindAll godoc
// @Summary      Get all ages
// @Tags         ages
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.Age "List of age"
// @Failure      500  {object}  models.ApiError "Internal Server Error"
// @Router       /ages [get]
func (a *AgeHandler) FindAll(c *gin.Context) {
	ages, err := a.ageRepo.FindAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}
	c.JSON(http.StatusOK, ages)
}

// FindById godoc
// @Summary      Find by id
// @Tags         ages
// @Accept       json
// @Produce      json
// @Param        id path int true "Ages id"
// @Success      200  {object}  models.Age "Ok"
// @Failure      400  {object}  models.ApiError "Invalid age Id"
// @Router       /ages/{id} [get]
func (a *AgeHandler) FindById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("inavalid age id"))
	}

	age, err := a.ageRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, age)

}

// Update godoc
// @Summary      Update age
// @Tags 		 ages
// @Accept       json
// @Produce      json
// @Param request body models.Age true "Age model"
// @Success      200  {object} object{id=int}  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid Age Id"
// @Failure   	 500  {object} models.ApiError
// @Router       /admin/ages/{id} [put]
func (a *AgeHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Age Id"))
		return
	}

	_, err = a.ageRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var request updateAgeRequest
	err = c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}

	filename, err := a.saveAgePoster(c, request.Poster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	age := models.Age{
		Age:       request.Age,
		PosterUrl: filename,
	}

	a.ageRepo.Update(c, id, age)

	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary      Delete age
// @Tags         ages
// @Accept       json
// @Produce      json
// @Param        id path int true "Ages id"
// @Success      200  {object}  models.Age "Ok"
// @Failure      400  {object}  models.ApiError "Invalid age Id"
// @Router       /admin/ages/{id} [delete]
func (a *AgeHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid age Id"))
		return
	}

	_, err = a.ageRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = a.ageRepo.Delete(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
