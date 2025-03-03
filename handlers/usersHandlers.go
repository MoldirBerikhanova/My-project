package handlers

import (
	"fmt"
	"goozinshe/models"
	"goozinshe/repositories"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UsersHandlers struct {
	userRepo *repositories.UsersRepository
}

func NewUsersHandlers(userRepo *repositories.UsersRepository) *UsersHandlers {
	return &UsersHandlers{userRepo: userRepo}
}

type createUserRequest struct {
	Name            *string               `form:"name"`
	Email           string                `form:"email"`
	Password        string                `form:"password"`
	ConfirmPassword string                `form:"confirm_password"`
	PhoneNumber     *string               `form:"phonenumber"`
	Birthday        *string               `form:"birthday"`
	Poster          *multipart.FileHeader `form:"posterUrl"`
}

type updateUserRequest struct {
	Name        *string               `form:"name"`
	Email       string                `form:"email"`
	Password    string                `form:"password"`
	PhoneNumber *string               `form:"phonenumber"`
	Birthday    *string               `form:"birthday"`
	Poster      *multipart.FileHeader `form:"posterUrl"`
}

type changePasswordRequest struct {
	Password string
}

type userResponse struct {
	Id          int        `form:"id"`
	Name        string     `form:"name"`
	Email       string     `form:"email"`
	PhoneNumber *int       `form:"phonenumber"`
	Birthday    *time.Time `form:"birthday"`
	Poster      string     `form:"posterUrl"`
}

func (h *UsersHandlers) AdminMiddleware(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.NewApiError("User not authenticated"))
		c.Abort()
		return
	}
	isAdmin, err := h.userRepo.IsAdmin(c, userId.(int))
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
// @Tags users
// @Summary      Find users by id
// @Accept       json
// @Produce      json
// @Param id path int true "User id"
// @Success      200  {array} handlers.userResponse "OK"
// @Failure   	 400  {object} models.ApiError "Invalid user id"
// @Failure   	 404  {object} models.ApiError "User not found"
// @Failure   	 500  {object} models.ApiError
// @Router       /users/{id} [get]
func (h *UsersHandlers) FindById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user Id"))
		return
	}

	user, err := h.userRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError("User not found"))
		return
	}
	c.JSON(http.StatusOK, user)
}

// FindAll godoc
// @Tags users
// @Summary      Get users list
// @Accept       json
// @Produce      json
// @Success      200  {array} handlers.userResponse "OK"
// @Failure   	 500  {object} models.ApiError
// @Router       /users [get]
func (h *UsersHandlers) FindAll(c *gin.Context) {
	users, err := h.userRepo.FindAll(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not load users"))
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *UsersHandlers) saveProfileImage(c *gin.Context, poster *multipart.FileHeader) (string, error) {
	filename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(poster.Filename))
	filepath := fmt.Sprintf("images/%s", filename)
	err := c.SaveUploadedFile(poster, filepath)

	return filename, err
}

// Create godoc
// @Tags users
// @Summary      Create user
// @Accept       json
// @Produce      json
// @Param 		name body string true "имя пользователя"
// @Param 		email body string true "эл.почта пользователя"
// @Param 		password body string true "пароль пользователя"
// @Param 		confirm_passowrd body string true "подтверждение пароля"
// @Param 		phonenumber body string true "моб.номер пользователя"
// @Param 		birthday body string true "дата рождения пользователя"
// @Success      200  {object} object{id=int} "OK"
// @Failure   	 400  {object} models.ApiError "Invalid data"
// @Failure   	 500  {object} models.ApiError
// @Router       /users [post]
func (h *UsersHandlers) Create(c *gin.Context) {
	var request createUserRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid payload"))
		return
	}

	user, err := h.userRepo.FindByEmail(c, request.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("User already exists"))
		return
	}

	if request.Poster == nil {
		log.Println("Poster file is required")
		c.JSON(http.StatusBadRequest, "Poster file is required")
		return
	}
	filename, err := h.saveProfileImage(c, request.Poster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	if request.PhoneNumber == nil || *request.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, "PhoneNumber file is required")
		return
	}

	formattedPhone, err := formatPhoneNumber(*request.PhoneNumber, "KZ")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid phone number"))
		return
	}

	birthday, err := time.Parse("02.01.2006", *request.Birthday)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid birthday format. Use DD.MM.YYYY"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to hash password"))
		return
	}

	user = models.User{
		Name:        *request.Name,
		Email:       request.Email,
		Password:    string(passwordHash),
		PhoneNumber: &formattedPhone,
		Birthday:    &birthday,
		PosterUrl:   &filename,
	}

	id, err := h.userRepo.Create(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not create user"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// Update godoc
// @Tags users
// @Summary      Update user
// @Accept       json
// @Produce      json
// @Param 		 name body string true "имя пользователя"
// @Param 		 email body string true "эл.почта пользователя"
// @Param 		 password body string true "пароль пользователя"
// @Param 		 phonenumber body string true "моб.номер пользователя"
// @Param 		 birthday body string true "дата рождения пользователя"
// @Success      200  {object} object{id=int} "OK"
// @Failure   	 400  {object} models.ApiError "Invalid data"
// @Failure   	 400  {object} models.ApiError "Invalid request payload"
// @Failure   	 404  {object} models.ApiError "User not found"
// @Failure   	 500  {object} models.ApiError
// @Router       /users/{id} [put]
func (h *UsersHandlers) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user Id"))
		return
	}

	var request updateUserRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request payload"))
		return
	}

	user, err := h.userRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError("User not found"))
		return
	}

	if request.Poster == nil {
		log.Println("Poster file is required")
		c.JSON(http.StatusBadRequest, "Poster file is required")
		return
	}
	filename, err := h.saveProfileImage(c, request.Poster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	if request.PhoneNumber == nil || *request.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, "PhoneNumber file is required")
		return
	}

	formattedPhone, err := formatPhoneNumber(*request.PhoneNumber, "KZ")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid phone number"))
		return
	}

	birthday, err := time.Parse("02.01.2006", *request.Birthday)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid birthday format. Use DD.MM.YYYY"))
		return
	}

	user.Name = *request.Name
	user.Email = request.Email
	user.Password = request.Password
	user.PhoneNumber = &formattedPhone
	user.Birthday = &birthday
	user.PhoneNumber = &filename

	err = h.userRepo.Update(c, id, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

// Delete godoc
// @Tags users
// @Summary      Delete user
// @Accept       json
// @Produce      json
// @Param id path int true "User id"
// @Success      200  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid data"
// @Failure   	 404  {object} models.ApiError "User not found"
// @Failure   	 500  {object} models.ApiError
// @Router       /users/{id} [delete]
func (h *UsersHandlers) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user Id"))
		return
	}

	_, err = h.userRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError("User not found"))
		return
	}

	err = h.userRepo.Delete(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}

// ChangePassword godoc
// @Tags users
// @Summary      Change user password
// @Accept       json
// @Produce      json
// @Param id path int true "User id"
// @Param request body handlers.changePasswordRequest true "Password data"
// @Success      200  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid data"
// @Failure   	 404  {object} models.ApiError "User not found"
// @Failure   	 500  {object} models.ApiError
// @Router       /users/{id}/changePassword [patch]
func (h *UsersHandlers) ChangePassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user Id"))
		return
	}

	var request changePasswordRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request payload"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to hash password"))
		return
	}

	user, err := h.userRepo.FindById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError("User not found"))
		return
	}

	user.Password = string(passwordHash)

	err = h.userRepo.Update(c, id, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.Status(http.StatusOK)
}
