Приложение 

У меня идея была такая, чтобы приложение работал по направлениям youtube и локальным хостом

1.Включила интеграцию c ютубом ////youtubeApi v3//
в итоге  добавила
ViewYouTube -  количество просмотров для трейлеров//при большом просмотре фильмов можно открыть пользователям
VideoUrl -  при отсутствии фильма на локальном сервере, можно смотреть через ютубм

video url для сериалов будет null


2.Чтобы просматривать количество просмотров от Пользователей User воспользовалась библиотекой Promethei


3.Добавила isAdmin в age, genre, category, users, movies (для всех ендпоинтов)

4.Для пользователей доступны только 2 ендпоинта
	authorized.GET("/movies/user", moviesHandler.FindAllforUsers)
	authorized.GET("/movies/user/:movieId", moviesHandler.FindByIdforUsers)


5.Так же есть папка удалленых файлов и кодов, и функции CRUD доступны только админам
_________________________________________________________________________________________________________

6.Нурзат я удалила повторяющиеся функции в хэндлерах (movies, genres, categories, ages, allseries)
userId, exists := c.Get("userId")
if !exists {
    c.JSON(http.StatusUnauthorized, models.NewApiError("User not authenticated"))
    return
}

isAdmin, err := h.moviesRepo.IsAdmin(c, userId.(int))
if err != nil {
    c.JSON(http.StatusInternalServerError, models.NewApiError("could not check user role"))
    return
}

if !isAdmin {
    c.JSON(http.StatusForbidden, models.NewApiError("Access forbidden"))
    return
}

__________________________________________________________________________________________________________
cделала одну функцию для видео, скринов и изображении
func saveUploadedFile(c *gin.Context, file *multipart.FileHeader, folder string) (string, error) {
	filename := fmt.Sprintf("%s%s", uuid.NewString(), filepath.Ext(file.Filename))
	filepath := fmt.Sprintf("%s/%s", folder, filename)
	err := c.SaveUploadedFile(file, filepath)
	return filename, err
}
_______________________________________________________________________________________________________

поменяла ендпоинты с allseries на /movies/allseries
    admin.POST("/movies/allseries", allseriesHandlers.Create)
	admin.PUT("/movies/allseries/:movieId", allseriesHandlers.Update)
	admin.DELETE("/movies/allseries/:movieId", allseriesHandlers.Delete)

___________________________________________________________________________________________________
