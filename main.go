package main

import (
	"context"
	"goozinshe/config"
	"goozinshe/docs"
	"goozinshe/handlers"
	"goozinshe/logger"
	"goozinshe/middlewares"
	"goozinshe/prometheus"
	"goozinshe/repositories"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// @title           OZINSHE	API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8081
// @BasePath  /
// @securityDefinitions.basic  BasicAuth
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	r := gin.Default()
	prometheus.InitPrometheus()
	logger := logger.GetLogger()
	r.Use(
		ginzap.Ginzap(logger, time.RFC3339, true),
		ginzap.RecoveryWithZap(logger, true),
	)

	corsConfig := cors.Config{
		AllowAllOrigins: true,
		AllowHeaders:    []string{"*"},
		AllowMethods:    []string{"*"},
	}
	r.Use(cors.New(corsConfig))

	err := loadConfig()
	if err != nil {
		panic(err)
	}

	conn, err := connectToDb()
	if err != nil {
		panic(err)
	}

	youtubeService, err := connectToYouTube()
	if err != nil {
		panic(err)
	}

	moviesRepository := repositories.NewMoviesRepository(conn, youtubeService)
	genresRepostiroy := repositories.NewGenresRepository(conn)
	categoryRepository := repositories.NewCategoryRepository(conn)
	ageRepository := repositories.NewAgeRepository(conn)
	usersRepository := repositories.NewUsersRepository(conn)
	allseriesRepository := repositories.NewAllSeriesRepository(conn)
	selectedRepository := repositories.NewSelectedlistRepository(conn)
	seasonRepository := repositories.NewSeasonRepository(conn)

	moviesHandler := handlers.NewMoviesHandler(
		moviesRepository,
		genresRepostiroy,
		categoryRepository,
		ageRepository,
		allseriesRepository,
		seasonRepository,
	)

	selectedHandlers := handlers.NewSelectedlistHandler(moviesRepository, selectedRepository)
	genresHandler := handlers.NewGenreHanlers(genresRepostiroy)
	imageHandlers := handlers.NewImageHandlers()
	videoHandlers := handlers.NewVideoHandlers()
	categoryHandlers := handlers.NewCategoryHandlers(categoryRepository)
	agesHandlers := handlers.NewAgeHandler(ageRepository)
	usersHandlers := handlers.NewUsersHandlers(usersRepository)
	authHandlers := handlers.NewAuthHandlers(usersRepository)
	allseriesHandlers := handlers.NewAllSeriesHandlers(allseriesRepository)
	SeasonsHandlers := handlers.NewSeasonsHandlers(seasonRepository, allseriesRepository)

	authorized := r.Group("")
	authorized.Use(middlewares.AuthMiddleware)

	admin := r.Group("/admin")
	admin.Use(middlewares.AuthMiddleware)

	admin.GET("/movies", moviesHandler.FindAll)
	admin.POST("/movies", moviesHandler.Create)
	admin.PUT("/movies/:id", moviesHandler.Update)
	admin.DELETE("/movies/:id", moviesHandler.Delete)
	admin.GET("/movies/:id", moviesHandler.FindByIdAdmin)
	admin.POST("/genres", genresHandler.Create)
	admin.PUT("/genres/:id", genresHandler.Update)
	admin.DELETE("/genres/:id", genresHandler.Delete)
	admin.POST("/categories", categoryHandlers.Create)
	admin.DELETE("/categories/:id", categoryHandlers.Delete)
	admin.PUT("/categories/:id", categoryHandlers.Update)
	admin.POST("/ages", agesHandlers.HandleAddAge)
	admin.PUT("/ages/:id", agesHandlers.Update)
	admin.DELETE("/ages/:id", agesHandlers.Delete)
	admin.POST("/movies/allseries", allseriesHandlers.Create)
	admin.PUT("/movies/allseries/:movieId", allseriesHandlers.Update)
	admin.DELETE("/movies/allseries/:movieId", allseriesHandlers.Delete)
	admin.POST("/movies/seasons", SeasonsHandlers.Create)
	admin.PUT("/movies/seasons/:seasonId", SeasonsHandlers.Update)
	admin.DELETE("/movies/seasons/:seasonId", SeasonsHandlers.Delete)

	admin.PATCH("/users/:id/changePassword", usersHandlers.ChangePassword)
	admin.POST("/users", usersHandlers.Create)
	admin.PUT("/users/:id", usersHandlers.Update)
	admin.DELETE("/users/:id", usersHandlers.Delete)
	admin.GET("/users", usersHandlers.FindAll)
	admin.GET("/users/:id", usersHandlers.FindById)

	//Users//
	authorized.GET("/movies", moviesHandler.FindAllforUsers)
	authorized.GET("/movies/:movieId", moviesHandler.FindByIdforUsers)
	authorized.GET("/genres/:id", genresHandler.FindById)
	authorized.GET("/genres", genresHandler.FindAll)
	authorized.GET("/categories", categoryHandlers.FindAll)
	authorized.GET("/categories/:id", categoryHandlers.FindById)
	authorized.GET("/ages", agesHandlers.FindAll)
	authorized.GET("/ages/:id", agesHandlers.FindById)
	authorized.GET("/movies/allseries/:movieId", allseriesHandlers.FindById)
	authorized.GET("/movies/allseries", allseriesHandlers.FindAll)
	authorized.POST("/selected/:movieId", selectedHandlers.HandleAddMovie)
	authorized.GET("/selected", selectedHandlers.HandleGetMoviesAndSeries)
	authorized.GET("/movies/seasons", SeasonsHandlers.FindAll)
	authorized.GET("/movies/seasons/:seasonId", SeasonsHandlers.FindById)

	authorized.PATCH("/users/:id/changePassword", usersHandlers.ChangePassword)

	authorized.POST("/auth/signOut", authHandlers.SignOut)     //http://localhost:8081/auth/signOut
	authorized.GET("/auth/userInfo", authHandlers.GetUserInfo) //http://localhost:8081/auth/userInfo
	unauthorized := r.Group("")
	unauthorized.GET("/images/:imageId", imageHandlers.HandleGetImageById)
	unauthorized.GET("/video/:videoId", videoHandlers.HandleGetVideoById)

	unauthorized.POST("/auth/signIn", authHandlers.SignIn) //http://localhost:8081/auth/signIn
	unauthorized.POST("/auth/signUp", authHandlers.SignUp)
	authorized.PUT("/auth/signIn/:userId", authHandlers.UpdateUserProfile) /// заполнение профиля

	docs.SwaggerInfo.BasePath = "/"
	unauthorized.GET("/swagger/*any", swagger.WrapHandler(swaggerfiles.Handler))

	logger.Info("Application starting...")
	r.GET("/metrics", prometheus.MetricsHandler()) //http://localhost:8081/metrics
	r.Run(config.Config.AppHost)
}

func loadConfig() error {
	logger := logger.GetLogger()
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	var mapConfig config.MapConfig
	err = viper.Unmarshal(&mapConfig)
	if err != nil {
		return err
	}

	config.Config = &mapConfig
	if config.Config.Prometheus.Enabled {
		logger.Info("Prometheus monitoring is enabled")
	} else {
		logger.Info("Prometheus monitoring is disabled")
	}

	return nil
}

func connectToDb() (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), config.Config.DbConnectionString)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func connectToYouTube() (*youtube.Service, error) {
	youtubeService, err := youtube.NewService(context.Background(), option.WithAPIKey(config.Config.YouTubeAPIKey))
	if err != nil {
		return nil, err
	}
	return youtubeService, nil

}
