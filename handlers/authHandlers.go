package handlers

import (
	"fmt"
	"goozinshe/config"
	"goozinshe/logger"
	"goozinshe/models"
	"goozinshe/repositories"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var l = logger.GetLogger()

type AuthHandlers struct {
	userRepo *repositories.UsersRepository
}

func NewAuthHandlers(userRepo *repositories.UsersRepository) *AuthHandlers {
	return &AuthHandlers{userRepo: userRepo}
}

type signInRequest struct {
	Email    string
	Password string
}

type signUpRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type updateProfileRequset struct {
	Name        *string               `form:"name"`
	PhoneNumber *string               `form:"phonenumber"`
	Birthday    *string               `form:"birthday"`
	PosterUrl   *multipart.FileHeader `form:"posterUrl"`
}

func formatPhoneNumber(input string, defaultRegion string) (string, error) {
	num, err := phonenumbers.Parse(input, defaultRegion)
	if err != nil {
		l.Error("неверный формат номера", zap.Error(err))
		return "", fmt.Errorf("неверный формат номера: %w", err)
	}

	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("номер невалиден")
	}

	countryCode := fmt.Sprintf("+%d", num.GetCountryCode())
	nationalNumber := fmt.Sprintf("%d", num.GetNationalNumber())

	if len(nationalNumber) == 10 {
		return fmt.Sprintf("%s (%s) - %s - %s - %s",
			countryCode,
			nationalNumber[:3],
			nationalNumber[3:6],
			nationalNumber[6:8],
			nationalNumber[8:],
		), nil
	}

	return phonenumbers.Format(num, phonenumbers.INTERNATIONAL), nil
}

// SignUp godoc
// @Tags /auth/signUp
// @Summary     Авторизация (Регистрация пользователя)
// @Accept      json
// @Produce     json
// @Param       request body handlers.signUpRequest true "Данные пользователя"
// @Success     200 {object} object{id=int, profileIncomplete=bool} "OK"
// @Failure     400 {object} models.ApiError "Passwords do not match"
// @Failure     409 {object} models.ApiError "User already exists"
// @Failure     500 {object} models.ApiError "Database error"
// @Router      /auth/signUp [post]
func (h *AuthHandlers) SignUp(c *gin.Context) {
	var request signUpRequest
	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Database error"))
		return
	}

	if request.Password != request.ConfirmPassword {
		c.JSON(http.StatusBadRequest, models.NewApiError("Passwords do not match"))
		return
	}

	user, err := h.userRepo.FindByEmail(c, request.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Database error"))
		return
	}
	if user.Id != 0 {
		c.JSON(http.StatusConflict, models.NewApiError("User already exists"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to hash password"))
		return
	}

	user = models.User{
		Email:    request.Email,
		Password: string(passwordHash),
	}

	userId, err := h.userRepo.Create(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not create user"))
		return
	}
	l.Info("пожалуйста заполните профиль")
	profileIncomplete := user.Name == "" || user.Birthday == nil || user.PhoneNumber == nil
	l.Info("Пользователь зарегистрирован")
	c.JSON(http.StatusOK, gin.H{
		"id":                userId,
		"profileIncomplete": profileIncomplete,
	})
}

// UpdateUserProfile godoc
// @Tags /auth/signIn
// @Summary     Авторизация (Заполнение профиля пользователя)
// @Accept      json
// @Produce     json
// @Param 		name body string true "имя пользователя"
// @Param 		phonenumber body string true "моб.номер пользователя"
// @Param 		birthday body string true "дата рождения пользователя"
// @Success     200 {object} object{id=int, profileIncomplete=bool} "OK"
// @Failure     400 {object} models.ApiError "Invalid user Id"
// @Failure     400 {object} models.ApiError "Database error"
// @Failure     404 {object} models.ApiError "User not found"
// @Failure     500 {object} models.ApiError "could not create user"
// @Router      /auth/signIn/:userId [post]
func (h *AuthHandlers) UpdateUserProfile(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid user Id"))
		return
	}

	var prorequest updateProfileRequset
	err = c.Bind(&prorequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Database error"))
		return
	}

	user, err := h.userRepo.FindById(c, userId)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewApiError("User not found"))
		return
	}

	l.Debug("Request body:", zap.Any("prorequest", prorequest.PosterUrl))
	if prorequest.PosterUrl == nil {
		l.Warn("Poster file is required")
		c.JSON(http.StatusBadRequest, "Poster file is required")
		return
	}
	filename, err := h.saveProfileImage(c, prorequest.PosterUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	if prorequest.PhoneNumber == nil || *prorequest.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, "PhoneNumber file is required")
		return
	}

	formattedPhone, err := formatPhoneNumber(*prorequest.PhoneNumber, "KZ")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid phone number"))
		return
	}

	birthday, err := time.Parse("02.01.2006", *prorequest.Birthday)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid birthday format. Use DD.MM.YYYY"))
		return
	}

	user.Name = *prorequest.Name
	user.PhoneNumber = &formattedPhone
	user.Birthday = &birthday
	user.PosterUrl = &filename

	err = h.userRepo.UpdateUserProfiile(c, userId, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not create user"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":        user.Name,
		"phonenumber": user.PhoneNumber,
		"birthday":    user.Birthday,
		"posterUrl":   user.PosterUrl,
	})
}

// SignIn godoc
// @Tags /auth/signIn
// @Summary     Авторизация (вход в систему)
// @Accept      json
// @Produce     json
// @Param 	    email body string true "эл.почта пользователя"
// @Param 		password body string true "пароль пользователя"
// @Success     200 {object} object{id=int, profileIncomplete=bool} "OK"
// @Failure     400 {object} models.ApiError "Invalid request payload""
// @Failure     401 {object} models.ApiError "Invalid credials"
// @Failure     500 {object} models.ApiError "could not generate jwt token"
// @Router      /auth/signIn [post]
func (h *AuthHandlers) SignIn(c *gin.Context) {
	l.Info("Добро пожаловать в систему")
	var request signInRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid request payload"))
		return
	}

	user, err := h.userRepo.FindByEmail(c, request.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		l.Warn("Ошибка авторизации", zap.String("email", request.Email))
		c.JSON(http.StatusUnauthorized, models.NewApiError("Invalid credials"))
		return
	}

	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(user.Id),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.Config.JwtExpiresIn)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.Config.JwtSecretKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not generate jwt token"))
		return
	}

	profileIncomplete := user.Name == "" || user.Birthday == nil || user.PhoneNumber == nil
	if profileIncomplete {
		l.Info("пожалуйста заполните профиль")
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString, "profileIncomplete": profileIncomplete})

}

func (h *AuthHandlers) saveProfileImage(c *gin.Context, poster *multipart.FileHeader) (string, error) {
	filename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(poster.Filename))
	filePath := fmt.Sprintf("images/%s", filename)
	err := c.SaveUploadedFile(poster, filePath)

	return filename, err
}

// SignOut godoc
// @Tags /auth/signOut
// @Summary     Авторизация (выход из системы)
// @Accept      json
// @Success     200 {object} object{id=int, profileIncomplete=bool} "OK"
// @Router      /auth/signOut [post]
func (h *AuthHandlers) SignOut(c *gin.Context) {
	c.Status(http.StatusOK)
}

// GetUserInfo godoc
// @Tags /auth/userInfo
// @Summary     Авторизация (получение информации о пользователе)
// @Accept      json
// @Produce     json
// @Param 		name body string true "имя пользователя"
// @Param 	    email body string true "эл.почта пользователя"
// @Success     200 {object} object{id=int, profileIncomplete=bool} "OK"
// @Failure     500 {object} models.ApiError "could not generate jwt token"
// @Router      /auth/userInfo [get]
func (h *AuthHandlers) GetUserInfo(c *gin.Context) {
	userId := c.GetInt("userId")
	user, err := h.userRepo.FindById(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, userResponse{
		Id:    user.Id,
		Email: user.Email,
		Name:  user.Name,
	})
}
