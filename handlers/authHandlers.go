package handlers

import (
	"fmt"
	"goozinshe/config"
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
	"golang.org/x/crypto/bcrypt"
)

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

type signUpRequset struct {
	Name        string                `form:"name"`
	Email       string                `form:"email"`
	Password    string                `form:"password"`
	PhoneNumber string                `form:"phonenumber"`
	Birthday    string                `form:"birthday"`
	PosterUrl   *multipart.FileHeader `form:"poster_url"`
}

func formatPhoneNumber(input string, defaultRegion string) (string, error) {
	num, err := phonenumbers.Parse(input, defaultRegion)
	if err != nil {
		return "", fmt.Errorf("неверный формат номера: %v", err)
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

func (h *AuthHandlers) SignUp(c *gin.Context) {
	var request signUpRequset
	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Database error"))
		return
	}

	user, err := h.userRepo.FindByEmail(c, request.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Database error"))
		return
	}
	if user.Id != 0 { // Если Id заполнен, значит пользователь уже есть
		c.JSON(http.StatusConflict, models.NewApiError("User already exists"))
		return
	}

	if request.PosterUrl == nil {
		c.JSON(http.StatusBadRequest, "Poster file is required")
		return
	}
	filename, err := h.saveGenrePoster(c, request.PosterUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	birthday, err := time.Parse("02.01.2006", request.Birthday)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid birthday format. Use DD.MM.YYYY"))
		return
	}

	formattedPhone, err := formatPhoneNumber(request.PhoneNumber, "KZ") // "KZ" — код страны по умолчанию
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid phone number"))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to hash password"))
		return
	}

	user = models.User{
		Name:        request.Name,
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

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func (h *AuthHandlers) SignIn(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (h *AuthHandlers) saveGenrePoster(c *gin.Context, poster *multipart.FileHeader) (string, error) {
	filename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(poster.Filename))
	filePath := fmt.Sprintf("images/%s", filename)
	err := c.SaveUploadedFile(poster, filePath)

	return filename, err
}

func (h *AuthHandlers) SignOut(c *gin.Context) {
	c.Status(http.StatusOK)
}

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
