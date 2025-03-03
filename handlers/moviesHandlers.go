package handlers

import (
	"fmt"
	"goozinshe/models"
	"goozinshe/prometheus"
	"goozinshe/repositories"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go.uber.org/zap"
)

type MoviesHandler struct {
	moviesRepo    *repositories.MoviesRepository
	genresRepo    *repositories.GenresRepository
	categoryRepo  *repositories.CategoryRepository
	ageRepo       *repositories.AgeRepository
	allseriesRepo *repositories.AllSeriesRepository
	seasonRepo    *repositories.SeasonRepository
}

type createMovieRequest struct {
	Title        string                `form:"title"`
	Description  string                `form:"description"`
	ReleaseYear  int                   `form:"releaseYear"`
	Director     string                `form:"director"`
	Producer     string                `form:"producer"`
	TrailerUrl   string                `form:"trailerUrl"`
	PosterUrl    *multipart.FileHeader `form:"posterUrl"`
	Views        *int64                `form:"viewsYT"`
	Duration     string                `form:"duration"`
	VideoUrl     *multipart.FileHeader `form:"videoUrl"`
	ViewsCount   *int                  `form:"views_count"`
	ScreenSrc    *multipart.FileHeader `form:"screen_src"`
	GenreIds     []int                 `form:"genreIds"`
	CategoryIds  []int                 `form:"categoryIds"`
	AgeIds       []int                 `form:"ageIds"`
	AllSeriesIds []int                 `form:"allseriesIds"`
}

// Title:       request.Title,
// 		Description: request.Description,
// 		ReleaseYear: request.ReleaseYear,
// 		Director:    request.Director,
// 		TrailerUrl:  request.TrailerUrl,
// 		Duration:    &request.Duration,
// 		PosterUrl:   posterFilename,
// 		VideoUrl:    &videoFilename,
// 		ScreenSrc:   &screenFilename,
// 		Genres:      genres,
// 		Category:    categories,
// 		Ages:        ages,
// 		AllSeries:   allseries,

type updateMovieRequest struct {
	Title       string                `form:"title"`
	Description string                `form:"description"`
	ReleaseYear int                   `form:"releaseYear"`
	Director    string                `form:"director"`
	TrailerUrl  string                `form:"trailerUrl"`
	PosterUrl   *multipart.FileHeader `form:"posterUrl"`
	Duration    string                `form:"duration"`
	VideoUrl    *multipart.FileHeader `form:"videoUrl"`
	ScreenSrc   *multipart.FileHeader `form:"screen_src"`
	GenreIds    []int                 `form:"genreIds"`
	CategoryIds []int                 `form:"categoryIds"`
	AgeIds      []int                 `form:"ageIds"`
}

func NewMoviesHandler(
	moviesRepo *repositories.MoviesRepository,
	genreRepo *repositories.GenresRepository,
	categoryRepo *repositories.CategoryRepository,
	ageRepo *repositories.AgeRepository,
	allseriesRepo *repositories.AllSeriesRepository,
	seasonRepo *repositories.SeasonRepository,
) *MoviesHandler {
	return &MoviesHandler{
		moviesRepo:    moviesRepo,
		genresRepo:    genreRepo,
		categoryRepo:  categoryRepo,
		ageRepo:       ageRepo,
		allseriesRepo: allseriesRepo,
		seasonRepo:    seasonRepo,
	}
}

func (h *MoviesHandler) AdminMiddleware(c *gin.Context) {
	l.Info("AdminMiddleware: проверка пользователя")
	userId, exists := c.Get("userId")
	l.Info("DEBUG:", zap.Int("userId in context:", userId.(int)), zap.Bool("Exists:", exists))
	if !exists {
		l.Info("AdminMiddleware: userId not found")
		c.JSON(http.StatusUnauthorized, models.NewApiError("User not authenticated"))
		c.Abort()
		return
	}

	isAdmin, err := h.moviesRepo.IsAdmin(c, userId.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("could not check user role"))
		c.Abort()
		return
	}

	if !isAdmin {
		l.Info("AdminMiddleware: userId not found")
		c.JSON(http.StatusForbidden, models.NewApiError("Access forbidden"))
		c.Abort()
		return
	}
	l.Info("DEBUG: AdminMiddleware передает управление дальше")
	c.Next()
	l.Info("DEBUG: Вернулись в AdminMiddleware после c.Next()")
}

// FindByIdAdmin godoc
// Преимущество ViewsCount, ViewsYoutube, VideoUrl
// для
// @Summary      поиск доступен только админам
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        id path int true "Movie id"
// @Success      200  {object}  models.Movie "Ok"
// @Failure      400  {object}  models.ApiError "Invalid Movie Id"
// @Router        /admin/movies/{id} [get]
func (h *MoviesHandler) FindByIdAdmin(c *gin.Context) {
	start := time.Now()
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Movie Id"))
		prometheus.HttpDuration.WithLabelValues("GET").Observe(time.Since(start).Seconds())
		return
	}

	l.Info("Поиск фильма с ID:", zap.Int("movie_id", id))

	movie, err := h.moviesRepo.FindByIdAdmin(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		prometheus.HttpDuration.WithLabelValues("GET").Observe(time.Since(start).Seconds())
		return
	}

	err = h.moviesRepo.IncrementViewsCount(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to increment view count"))
		prometheus.HttpDuration.WithLabelValues("GET").Observe(time.Since(start).Seconds())
		return
	}
	l.Info("ViewsCount обновлён для фильма", zap.Int("movie_id", id))
	c.JSON(http.StatusOK, movie)

	prometheus.HttpDuration.WithLabelValues("GET").Observe(time.Since(start).Seconds())
}

// FindAll godoc
// @Summary      Get all movies for ADMIN
// @Tags         movies
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.Movie "List of movies"
// @Failure      500  {object}  models.ApiError "Internal Server Error"
// @Router        /admin/movies [get]
func (h *MoviesHandler) FindAll(c *gin.Context) {
	start := time.Now()

	filters := models.MovieFilters{
		SearchTerm: c.Query("search"),
		IsWatched:  c.Query("iswatched"),
		GenreId:    c.Query("genreids"),
		AgeId:      c.Query("ageids"),
		CategoryId: c.Query("categoryids"),
		Sort:       c.Query("sort"),
	}
	l.Info("DEBUG Перед вызовом h.moviesRepo.FindAll я просто тестирую")
	movies, err := h.moviesRepo.FindAll(c, filters)
	if err != nil {
		l.Info("ERROR: Ошибка при получении фильмов:", zap.Error(err))
		c.JSON(http.StatusInternalServerError, err)
		prometheus.HttpDuration.WithLabelValues("GET").Observe(time.Since(start).Seconds())
		return
	}
	l.Info("Успешно получили фильмы, отправляем ответ")
	c.JSON(http.StatusOK, movies)
	prometheus.HttpDuration.WithLabelValues("GET").Observe(time.Since(start).Seconds())
}

// Create godoc
// @Summary      Create movie
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        title body string true "Title of the movie"
// @Param        description body string true "Description of the movie"
// @Param        releaseYear body int true "ReleaseYear of the movie"
// @Param        director body string true "Director"
// @Param        trailerUrl body string true "TrailerUrl"
// @Param      	 genreIds body []int true "Genre ids"
// @Param		 categoryIds body []int true "Category ids"
// @Param        ageIds body []int true "Age ids"
// @Success      200  {object}  object{id=int} "OK"
// @Failure      400  {object}  models.ApiError "Could not bind json"
// @Failure      500  {object}  models.ApiError
// @Router        /admin/movies [post]
func (h *MoviesHandler) Create(c *gin.Context) {
	l.Info("Перехожу к методу create")
	var request createMovieRequest
	err := c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Could not bind json"))
		return
	}

	genres, err := h.genresRepo.FindAllByIds(c, request.GenreIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	categories, err := h.categoryRepo.FindAllByIds(c, request.CategoryIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	ages, err := h.ageRepo.FindAllByIds(c, request.AgeIds)
	if err != nil {
		l.Error("Ошибка при получении возрастных ограничений", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	allseries, err := h.allseriesRepo.FindAllByIds(c, request.AllSeriesIds)
	if err != nil {
		l.Error("Ошибка при получении списка фильмов", zap.Error(err))
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	if request.PosterUrl == nil {
		c.JSON(http.StatusBadRequest, "Poster file is required")
		return
	}

	posterFilename, err := saveUploadedFile(c, request.PosterUrl, "images")
	if err != nil {
		l.Error("Ошибка загрузки постера", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	var videoFilename *string
	if request.VideoUrl != nil {
		filename, err := saveUploadedFile(c, request.VideoUrl, "video")
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
			return
		}
		videoFilename = &filename
	}

	screenFilename, err := saveUploadedFile(c, request.ScreenSrc, "screen")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	movie := models.Movie{
		Title:       request.Title,
		Description: request.Description,
		ReleaseYear: request.ReleaseYear,
		Director:    request.Director,
		Producer:    &request.Producer,
		TrailerUrl:  request.TrailerUrl,
		Duration:    &request.Duration,
		PosterUrl:   posterFilename,
		VideoUrl:    videoFilename,
		ScreenSrc:   &screenFilename,
		Genres:      genres,
		Category:    categories,
		Ages:        ages,
		AllSeries:   allseries,
	}

	id, err := h.moviesRepo.Create(c, movie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	l.Info("Фильм создан успешно")

	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
}

func saveUploadedFile(c *gin.Context, file *multipart.FileHeader, folder string) (string, error) {
	filename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(file.Filename))
	filepath := fmt.Sprintf("%s/%s", folder, filename)
	err := c.SaveUploadedFile(file, filepath)
	return filename, err
}

// Update godoc
// @Summary      Update movie
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        title body string true "Title of the movie"
// @Param        description body string true "Description of the movie"
// @Param        releaseYear body int true "ReleaseYear of the movie"
// @Param        director body string true "Director"
// @Param        trailerUrl body string true "TrailerUrl"
// @Param      	 genreIds body []int true "Genre ids"
// @Param		 categoryIds body []int true "Category ids"
// @Param        ageIds body []int true "Age ids"
// @Success      200  {object}  object{id=int} "OK"
// @Failure      400  {object}  models.ApiError "Could not bind json"
// @Failure      500  {object}  models.ApiError
// @Router        /admin/movies/{id} [put]
func (h *MoviesHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Movie Id"))
		return
	}

	_, err = h.moviesRepo.FindByIdAdmin(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	var request updateMovieRequest
	err = c.Bind(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Could not bind json"))
		return
	}

	genres, err := h.genresRepo.FindAllByIds(c, request.GenreIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}
	categories, err := h.categoryRepo.FindAllByIds(c, request.CategoryIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	ages, err := h.ageRepo.FindAllByIds(c, request.AgeIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	posterFilename, err := saveUploadedFile(c, request.PosterUrl, "images")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}
	videoFilename, err := saveUploadedFile(c, request.VideoUrl, "video")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	screenFilename, err := saveUploadedFile(c, request.ScreenSrc, "screen")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError(err.Error()))
		return
	}

	movie := models.Movie{
		Title:       request.Title,
		Description: request.Description,
		ReleaseYear: request.ReleaseYear,
		Director:    request.Director,
		TrailerUrl:  request.TrailerUrl,
		PosterUrl:   posterFilename,
		VideoUrl:    &videoFilename,
		ScreenSrc:   &screenFilename,
		Genres:      genres,
		Category:    categories,
		Ages:        ages,
	}

	err = h.moviesRepo.Update(c, id, movie)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to update movie"))
		return
	}

	l.Info("Фильм обновлен успешно")
	c.Status(http.StatusOK)
}

// Delete godoc
// @Summary      Delete movie
// @Tags         movies
// @Accept       json
// @Produce      json
// @Param        id path int true "Movie id"
// @Success      200  "OK"
// @Failure   	 400  {object} models.ApiError "Invalid data"
// @Failure   	 500  {object} models.ApiError
// @Router        /admin/movies/{id} [delete]
func (h *MoviesHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Movie Id"))
		return
	}

	_, err = h.moviesRepo.FindByIdAdmin(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		return
	}

	h.moviesRepo.Delete(c, id)

	c.Status(http.StatusNoContent)
}

// FindAllforUsers godoc
// @Summary      Get all movies for Users
// @Tags         movies для пользователей
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.Movie "List of movies"
// @Failure      500  {object}  models.ApiError "Internal Server Error"
// @Router       /movies/user [get]
func (h *MoviesHandler) FindAllforUsers(c *gin.Context) {
	filters := models.MovieFilters{
		SearchTerm: c.Query("search"),
		IsWatched:  c.Query("iswatched"),
		GenreId:    c.Query("genreids"),
		AgeId:      c.Query("ageids"),
		CategoryId: c.Query("categoryids"),
		Sort:       c.Query("sort"),
	}
	movies, err := h.moviesRepo.FindAllforUsers(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, movies)
}

// FindAllforUsers godoc
// @Summary      Get all movies for Users
// @Tags         movies для пользователей
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.Movie "List of movies"
// @Failure      500  {object}  models.ApiError "Internal Server Error"
// @Router       /movies/:movieId [get]
func (h *MoviesHandler) FindByIdforUsers(c *gin.Context) {
	start := time.Now()
	movieIdStr := c.Param("movieId")
	movieId, err := strconv.Atoi(movieIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError("Invalid Movie Id"))
		prometheus.HttpDuration.WithLabelValues("GET").Observe(time.Since(start).Seconds())
		return
	}

	l.Info("Поиск фильма с ID:", zap.Int("movie_id", movieId))

	movie, err := h.moviesRepo.FindByIdUser(c, movieId)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewApiError(err.Error()))
		prometheus.HttpDuration.WithLabelValues("GET").Observe(time.Since(start).Seconds())
		return
	}

	err = h.moviesRepo.IncrementViewsCount(c, movieId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewApiError("Failed to increment view count"))
		prometheus.HttpDuration.WithLabelValues("GET").Observe(time.Since(start).Seconds())
		return
	}
	l.Info("ViewsCount обновлён для фильма", zap.Int("movie_id", movieId))
	c.JSON(http.StatusOK, movie)

	prometheus.HttpDuration.WithLabelValues("GET").Observe(time.Since(start).Seconds())
}
