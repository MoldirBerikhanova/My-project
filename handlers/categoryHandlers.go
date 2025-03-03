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

type CategoryHandlers struct {
	categoryRepo *repositories.CategoryRepository
}

type createCategoryRequest struct {
	Title  string                `form:"title"`
	Poster *multipart.FileHeader `form:"poster"`
}

type updateCategoryRequest struct {
	Title  string                `form:"title"`
	Poster *multipart.FileHeader `form:"poster"`
}

func NewCategoryHandlers(categoryRepo *repositories.CategoryRepository) *CategoryHandlers {
	return &CategoryHandlers{
		categoryRepo: categoryRepo,
	}
}

func (h *CategoryHandlers) AdminMiddleware(c *gin.Context) {
	fmt.Println("AdminMiddleware started")
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewApiError("User not authenticated"))
		c.Abort()
		return
	}
	isAdmin, err := h.categoryRepo.IsAdmin(c, userId.(int))
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
// @Summary      Create category
// @Tags 		 categories
// @Accept       json
// @Produce      json
// @Param request body models.Category true "Category model"
// @Success      200  {object} object{id=int}  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid request category"
// @Failure   	 500  {object} models.ApiError
// @Router       /admin/categories [post]
func (h *CategoryHandlers) Create(c *gin.Context) {
	var request updateCategoryRequest
	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request category")
		return
	}

	if request.Poster == nil {
		c.JSON(http.StatusBadRequest, "Poster file is required")
		return
	}

	filename, err := h.saveCategoryPoster(c, request.Poster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	category := models.Category{
		Title:     request.Title,
		PosterUrl: filename,
	}

	id, err := h.categoryRepo.Create(c, category)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

// FindById godoc
// @Summary      Find by id
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id path int true "Category id"
// @Success      200  {object}  models.Category "Ok"
// @Failure      400  {object}  models.ApiError "Invalid Movie Id"
// @Router       /categories/{id} [get]
func (h *CategoryHandlers) FindById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid  categoryId"))
		return
	}

	category, err := h.categoryRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandlers) saveCategoryPoster(c *gin.Context, poster *multipart.FileHeader) (string, error) {
	filename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(poster.Filename))
	filepath := fmt.Sprintf("images/%s", filename)
	err := c.SaveUploadedFile(poster, filepath)

	return filename, err
}

// FindAll godoc
// @Summary      Get all categories
// @Tags         categories
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.Movie "List of movies"
// @Failure      500  {object}  models.ApiError "Internal Server Error"
// @Router       /categories [get]
func (h *CategoryHandlers) FindAll(c *gin.Context) {
	categories, err := h.categoryRepo.FindAll(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, categories)
}

// Update godoc
// @Summary      Update category
// @Tags 		 categories
// @Accept       json
// @Produce      json
// @Param request body models.Category true "Category model"
// @Success      200  {object} object{id=int}  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid Category Id"
// @Failure   	 500  {object} models.ApiError
// @Router       /admin/categories/{id} [put]
func (h *CategoryHandlers) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Genre Id"))
		return
	}

	_, err = h.categoryRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var request updateCategoryRequest
	err = c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}

	filename, err := h.saveCategoryPoster(c, request.Poster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	category := models.Category{
		Title:     request.Title,
		PosterUrl: filename,
	}

	h.categoryRepo.Update(c, id, category)

	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary      Delete category
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id path int true "Category id"
// @Success      200  {object}  models.Category "Ok"
// @Failure      400  {object}  models.ApiError "Invalid category Id"
// @Router       /admin/categories/{id} [delete]
func (h *CategoryHandlers) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid category Id"))
		return
	}

	_, err = h.categoryRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = h.categoryRepo.Delete(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
