package repositories

import (
	"context"
	"fmt"
	"goozinshe/config"
	"goozinshe/logger"
	"goozinshe/models"
	"regexp"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type MoviesRepository struct {
	db             *pgxpool.Pool
	youtubeService *youtube.Service
}

func NewMoviesRepository(conn *pgxpool.Pool, youtubeService *youtube.Service) *MoviesRepository {
	return &MoviesRepository{db: conn, youtubeService: youtubeService}
}

func (r *MoviesRepository) IsAdmin(c context.Context, id int) (bool, error) {
	var role string
	row := r.db.QueryRow(c, "select is_admin from users where id = $1", id)
	err := row.Scan(&role)
	if err != nil {
		return false, err
	}
	return role == "admin", nil
}

func (r *MoviesRepository) FindByIdAdmin(c context.Context, id int) (models.Movie, error) {
	sql := `
	SELECT 
        m.id,
        m.title,
        m.description,
        m.release_year,
        m.director,
        m.rating,
        m.is_favourite,
        m.trailer_url,
        m.poster_url,
		m.viewsyt,
		m.duration,
		m.video_url,
		m.views_count,
		m.screen_src,
		m.producer,
        g.id,
        g.title,
        g.poster_url,
        c.id,
        c.title,
        c.poster_url,
        a.id,
        a.age,
        a.poster_url,
        e.id, 
        e.series, 
        e.title,              
        e.trailer_url,
		e.duration,
		e.poster_url
    FROM movies m
    JOIN movies_genres mg ON mg.movie_id = m.id
    JOIN genres g ON mg.genre_id = g.id
    JOIN movies_categories mc ON mc.movie_id = m.id
    JOIN categories c ON mc.categorie_id = c.id
    JOIN movies_ages ma ON ma.movie_id = m.id
    JOIN ages a ON ma.age_id = a.id
    LEFT JOIN movies_allseries me ON me.movie_id = m.id
    LEFT JOIN allseries e ON me.allserie_id = e.id
	WHERE m.id = $1
	`

	sqlSeason := ` 
   SELECT 
        m.id,
        m.title,
        m.description,
        m.release_year,
        m.director,
        m.rating,
        m.is_favourite,
        m.trailer_url,
        m.poster_url,
		m.viewsyt,
		m.duration,
		m.video_url,
		m.views_count,
		m.screen_src,
		m.producer,
		s.id AS season_id,  
    	s.number AS season_number,
    	s.title AS season_title,
   		e2.id AS episode_id,  
        e2.series, 
        e2.title,              
        e2.trailer_url,
		e2.duration,
		e2.poster_url
    FROM movies m
	JOIN movies_seasons ms ON ms.movie_id = m.id
	JOIN season s ON ms.season_id = s.id
	LEFT JOIN seasons_allseries se ON se.season_id = s.id
	LEFT JOIN allseries e2 ON se.allserie_id = e2.id
	WHERE m.id = $1
	

    `

	//Producer     string      `form:"producer"`

	l := logger.GetLogger()

	rows, err := r.db.Query(c, sql, id)
	defer rows.Close()
	if err != nil {
		l.Error("Ошибка запроса к базе", zap.String("db_msg", err.Error()))
		return models.Movie{}, err
	}

	var movie models.Movie
	found := false

	categoriesMap := make(map[int]*models.Category)
	genresMap := make(map[int]*models.Genre)
	agesMap := make(map[int]*models.Age)
	allseriesMap := make(map[int]*models.AllSeries)

	for rows.Next() {
		var m models.Movie
		var g models.Genre
		var c models.Category
		var a models.Age
		var e models.AllSeries
		var views *int64

		if err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Description,
			&m.ReleaseYear,
			&m.Director,
			&m.Rating,
			&m.IsFavourite,
			&m.TrailerUrl,
			&m.PosterUrl,
			&views,
			&m.Duration,
			&m.VideoUrl,
			&m.ViewsCount,
			&m.ScreenSrc,
			&m.Producer,
			&g.Id,
			&g.Title,
			&g.PosterUrl,
			&c.Id,
			&c.Title,
			&c.PosterUrl,
			&a.Id,
			&a.Age,
			&a.PosterUrl,
			&e.Id,
			&e.Series,
			&e.Title,
			&e.TrailerUrl,
			&e.Duration,
			&e.PosterUrl,
		); err != nil {
			return models.Movie{}, err
		}

		if !found {
			movie = m
			found = true
		}

		if views != nil {
			movie.ViewsYouTube = views
		} else {
			movie.ViewsYouTube = new(int64)
		}

		videoID := extractVideoID(m.TrailerUrl)
		if videoID != "" {
			apiKey := config.Config.YouTubeAPIKey
			videoStats, err := getYouTubeVideoStats(apiKey, videoID)
			if err != nil {
				l.Warn("Ошибка получения статистики видео", zap.String("trailer", m.TrailerUrl), zap.Error(err))
			} else {
				viewsCount := int64(videoStats.Statistics.ViewCount)
				movie.ViewsYouTube = &viewsCount
				l.Info("Обновили views из YouTube API", zap.Int64("views", viewsCount))
			}
		}

		if _, exists := categoriesMap[c.Id]; !exists {
			categoriesMap[c.Id] = &c
		}

		if _, exists := genresMap[g.Id]; !exists {
			genresMap[g.Id] = &g
		}

		if _, exists := agesMap[a.Id]; !exists {
			agesMap[a.Id] = &a
		}

		if e.Id != nil {
			if _, exists := allseriesMap[*e.Id]; !exists {
				allseriesMap[*e.Id] = &e
			}
		}
	}

	if !found {
		l.Info("Фильм не найден!")
		return models.Movie{}, fmt.Errorf("фильм с ID %d не найден", id)
	}

	for _, v := range categoriesMap {
		movie.Category = append(movie.Category, *v)
	}

	for _, v := range genresMap {
		movie.Genres = append(movie.Genres, *v)
	}

	for _, v := range agesMap {
		movie.Ages = append(movie.Ages, *v)
	}

	for _, v := range allseriesMap {
		movie.AllSeries = append(movie.AllSeries, *v)
	}

	rows, err = r.db.Query(c, sqlSeason, id)
	defer rows.Close()
	if err != nil {
		l.Error("Ошибка запроса к базе", zap.String("db_msg", err.Error()))
		return models.Movie{}, err
	}
	//var season models.Season
	found = false
	seasonMap := make(map[int]*models.Season)

	for rows.Next() {
		var m models.Movie
		var s models.Season
		var e2 models.AllSeries
		var views *int64
		if err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Description,
			&m.ReleaseYear,
			&m.Director,
			&m.Rating,
			&m.IsFavourite,
			&m.TrailerUrl,
			&m.PosterUrl,
			&views, // используем указатель на int64
			&m.Duration,
			&m.VideoUrl,
			&m.ViewsCount,
			&m.ScreenSrc,
			&m.Producer,
			&s.Id,
			&s.Number,
			&s.Title,
			&e2.Id,
			&e2.Series,
			&e2.Title,
			&e2.TrailerUrl,
			&e2.Duration,
			&e2.PosterUrl,
		); err != nil {
			return models.Movie{}, err
		}
		// if !found {
		// 	season = s
		// 	found = true
		// }

		if _, exists := seasonMap[s.Id]; !exists {
			seasonMap[s.Id] = &s
		}

		if _, exists := allseriesMap[*e2.Id]; !exists {
			allseriesMap[*e2.Id] = &e2
		}
	}

	for _, v := range seasonMap {
		for _, t := range allseriesMap {
			v.AllSeries = append(v.AllSeries, *t)
		}
		movie.Season = append(movie.Season, *v)
	}

	return movie, nil
}

// if e2.Id != nil && *e2.Id != 0 {
// 	found := false
// 	for _, existingSeries := range seasonMap[s.Id].AllSeries {
// 		if existingSeries.Id != nil && *existingSeries.Id == *e2.Id {
// 			found = true
// 			break
// 		}
// 	}
// 	if !found {
// 		seasonMap[s.Id].AllSeries = append(seasonMap[s.Id].AllSeries, e2)
// 	}
// }
// }
// for _, movie := range movies {
// seasonExists := make(map[int]bool)
// for _, season := range seasonMap {
// 	if !seasonExists[season.Id] {
// 		movie.Season = append(movie.Season, *season)
// 		seasonExists[season.Id] = true
// 	}
// }

func (r *MoviesRepository) FindAll(c context.Context, filters models.MovieFilters) ([]models.Movie, error) {

	sql := ` 
   SELECT 
        m.id,
        m.title,
        m.description,
        m.release_year,
        m.director,
        m.rating,
        m.is_favourite,
        m.trailer_url,
        m.poster_url,
		m.viewsyt,
		m.duration,
		m.video_url,
		m.views_count,
		m.screen_src,
		m.producer,
        g.id,
        g.title,
        g.poster_url,
        c.id,
        c.title,
        c.poster_url,
        a.id,
        a.age,
        a.poster_url,
        e.id, 
        e.series, 
        e.title,              
        e.trailer_url,
		e.duration,
		e.poster_url
    FROM movies m
    JOIN movies_genres mg ON mg.movie_id = m.id
    JOIN genres g ON mg.genre_id = g.id
    JOIN movies_categories mc ON mc.movie_id = m.id
    JOIN categories c ON mc.categorie_id = c.id
    JOIN movies_ages ma ON ma.movie_id = m.id
    JOIN ages a ON ma.age_id = a.id
    left JOIN movies_allseries me ON me.movie_id = m.id
    left JOIN allseries e ON me.allserie_id = e.id
	

    `
	sqlSeason := ` 
   SELECT 
        m.id,
        m.title,
        m.description,
        m.release_year,
        m.director,
        m.rating,
        m.is_favourite,
        m.trailer_url,
        m.poster_url,
		m.viewsyt,
		m.duration,
		m.video_url,
		m.views_count,
		m.screen_src,
		m.producer,
		s.id AS season_id,  
    	s.number AS season_number,
    	s.title AS season_title,
   		e2.id AS episode_id,  
        e2.series, 
        e2.title,              
        e2.trailer_url,
		e2.duration,
		e2.poster_url
    FROM movies m
	JOIN movies_seasons ms ON ms.movie_id = m.id
	JOIN season s ON ms.season_id = s.id
	LEFT JOIN seasons_allseries se ON se.season_id = s.id
	LEFT JOIN allseries e2 ON se.allserie_id = e2.id;
    `

	params := pgx.NamedArgs{}

	if filters.SearchTerm != "" {
		//	'%%%s%%' => '%поиск%'
		sql = fmt.Sprintf("%s and m.title ilike @s", sql)
		params["s"] = fmt.Sprintf("%%%s%%", filters.SearchTerm)
	}
	if filters.GenreId != "" {
		sql = fmt.Sprintf("%s and g.id = @genreId", sql)
		params["genreId"] = filters.GenreId
	}

	if filters.AgeId != "" {
		sql = fmt.Sprintf("%s and a.id = @ageId", sql)
		params["AgeId"] = filters.AgeId
	}

	if filters.CategoryId != "" {
		sql = fmt.Sprintf("%s and c.id = @categoryId", sql)
		params["categoryId"] = filters.CategoryId
	}

	if filters.IsWatched != "" {
		isWatched, _ := strconv.ParseBool(filters.IsWatched)

		sql = fmt.Sprintf("%s and m.is_watched = @isWatched", sql)
		params["isWatched"] = isWatched
	}
	if filters.Sort != "" {
		identifier := pgx.Identifier{filters.Sort}
		sql = fmt.Sprintf("%s order by m.%s", sql, identifier.Sanitize())
	}

	l := logger.GetLogger()
	l.Info("Executing SQL Query in FindAll", zap.String("query", sql))
	rows, err := r.db.Query(c, sql)
	if err != nil {
		l.Error(err.Error())
		return nil, err
	}
	l.Info("SQL Query executed successfully")

	movies := make([]*models.Movie, 0)
	moviesMap := make(map[int]*models.Movie)
	allseriesMap := make(map[int]*models.AllSeries, 0)

	for rows.Next() {
		var m models.Movie
		var g models.Genre
		var c models.Category
		var a models.Age
		var e models.AllSeries
		var views *int64
		if err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Description,
			&m.ReleaseYear,
			&m.Director,
			&m.Rating,
			&m.IsFavourite,
			&m.TrailerUrl,
			&m.PosterUrl,
			&views, // используем указатель на int64
			&m.Duration,
			&m.VideoUrl,
			&m.ViewsCount,
			&m.ScreenSrc,
			&m.Producer,
			&g.Id,
			&g.Title,
			&g.PosterUrl,
			&c.Id,
			&c.Title,
			&c.PosterUrl,
			&a.Id,
			&a.Age,
			&a.PosterUrl,
			&e.Id,
			&e.Series,
			&e.Title,
			&e.TrailerUrl,
			&e.Duration,
			&e.PosterUrl,
		); err != nil {
			return nil, err
		}

		if views != nil {
			m.ViewsYouTube = views
		} else {
			m.ViewsYouTube = nil
		}

		if _, exists := moviesMap[m.Id]; !exists {
			moviesMap[m.Id] = &m
			movies = append(movies, &m)
		}

		videoID := extractVideoID(m.TrailerUrl)
		if videoID != "" {
			apiKey := config.Config.YouTubeAPIKey
			videoStats, err := getYouTubeVideoStats(apiKey, videoID)
			if err != nil {
				l.Warn("Failed to fetch video stats for trailer", zap.String("trailer", m.TrailerUrl), zap.Error(err))
			} else {
				views := int64(videoStats.Statistics.ViewCount)
				moviesMap[m.Id].ViewsYouTube = &views
			}
		}
		genreExists := false
		for _, existingGenres := range moviesMap[m.Id].Genres {
			if existingGenres.Id == g.Id {
				genreExists = true
				break
			}
		}
		if !genreExists {
			moviesMap[m.Id].Genres = append(moviesMap[m.Id].Genres, g)
		}

		categoryExists := false
		for _, existingCategory := range moviesMap[m.Id].Category {
			if existingCategory.Id == c.Id {
				categoryExists = true
				break
			}
		}
		if !categoryExists {
			moviesMap[m.Id].Category = append(moviesMap[m.Id].Category, c)
		}

		ageExists := false
		for _, existingAge := range moviesMap[m.Id].Ages {
			if existingAge.Id == a.Id {
				ageExists = true
				break
			}
		}
		if !ageExists {
			moviesMap[m.Id].Ages = append(moviesMap[m.Id].Ages, a)
		}

		if e.Id != nil && *e.Id != 0 {
			if _, exists := allseriesMap[*e.Id]; !exists {
				allseriesMap[*e.Id] = &e
				moviesMap[m.Id].AllSeries = append(moviesMap[m.Id].AllSeries, e)
			}
		}
	}
	rows, err = r.db.Query(c, sqlSeason)
	if err != nil {
		l.Error(err.Error())
		return nil, err
	}
	//Season := make([]*models.Season, 0)
	seasonMap := make(map[int]*models.Season)

	for rows.Next() {
		var m models.Movie
		var s models.Season
		var e2 models.AllSeries
		var views *int64
		if err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Description,
			&m.ReleaseYear,
			&m.Director,
			&m.Rating,
			&m.IsFavourite,
			&m.TrailerUrl,
			&m.PosterUrl,
			&views, // используем указатель на int64
			&m.Duration,
			&m.VideoUrl,
			&m.ViewsCount,
			&m.ScreenSrc,
			&m.Producer,
			&s.Id,
			&s.Number,
			&s.Title,
			&e2.Id,
			&e2.Series,
			&e2.Title,
			&e2.TrailerUrl,
			&e2.Duration,
			&e2.PosterUrl,
		); err != nil {
			return nil, err
		}
		if _, exists := moviesMap[m.Id]; !exists {
			moviesMap[m.Id] = &m
			movies = append(movies, &m)
		}
		if _, exists := seasonMap[s.Id]; !exists {
			seasonMap[s.Id] = &s
			moviesMap[m.Id].Season = append(moviesMap[m.Id].Season, s)
		}

		if e2.Id != nil && *e2.Id != 0 {
			found := false
			for _, existingSeries := range seasonMap[s.Id].AllSeries {
				if existingSeries.Id != nil && *existingSeries.Id == *e2.Id {
					found = true
					break
				}
			}
			if !found {
				seasonMap[s.Id].AllSeries = append(seasonMap[s.Id].AllSeries, e2)
			}
		}
	}
	for _, movie := range movies {
		seasonExists := make(map[int]bool)
		for _, season := range seasonMap {
			if !seasonExists[season.Id] {
				movie.Season = append(movie.Season, *season)
				seasonExists[season.Id] = true
			}
		}
	}

	// if _, exists := seasonMap[s.Id]; !exists {
	// 	seasonMap[s.Id] = &s
	// 	Season = append(Season, &s)
	// }

	// if e2.Id != nil {
	// 	exists := false
	// 	for _, existingAllSeries := range seasonMap[s.Id].AllSeries {
	// 		if existingAllSeries.Id == e2.Id {
	// 			exists = true
	// 			break
	// 		}
	// 	}
	// 	if !exists {
	// 		seasonMap[s.Id].AllSeries = append(seasonMap[s.Id].AllSeries, e2)
	// 	}
	// }

	err = rows.Err()
	if err != nil {
		l.Error(err.Error())
		return nil, err
	}

	concreteMovies := make([]models.Movie, 0, len(movies))
	for _, v := range movies {
		concreteMovies = append(concreteMovies, *v)
	}

	return concreteMovies, nil
}

func getYouTubeVideoStats(apiKey, videoID string) (*youtube.Video, error) {
	service, err := youtube.NewService(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать сервис YouTube: %w", err)
	}

	call := service.Videos.List([]string{"snippet", "statistics"}).Id(videoID)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе данных: %w", err)
	}
	if len(response.Items) == 0 {
		return nil, fmt.Errorf("видео не найдено")
	}
	return response.Items[0], nil
}

func extractVideoID(url string) string {
	re := regexp.MustCompile(`(?:youtube\.com\/(?:[^\/\n\s]+\/[^\/\n\s]+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([a-zA-Z0-9_-]{11})`)
	match := re.FindStringSubmatch(url)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
func (r *MoviesRepository) Create(c context.Context, movie models.Movie) (int, error) {
	l := logger.GetLogger()
	var id int

	tx, err := r.db.Begin(c)
	if err != nil {
		l.Error(err.Error())
		return 0, err
	}

	defer func() {
		if err != nil {
			tx.Rollback(c)
		}
	}()

	row := tx.QueryRow(c,
		` 
    insert into movies(title, description, release_year, director, trailer_url, poster_url, duration, video_url, screen_src, producer)
    values($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    returning id
    `,
		movie.Title,
		movie.Description,
		movie.ReleaseYear,
		movie.Director,
		movie.TrailerUrl,
		movie.PosterUrl,
		movie.Duration,
		movie.VideoUrl,
		movie.ScreenSrc,
		movie.Producer)

	err = row.Scan(&id)
	if err != nil {
		l.Error(err.Error())
		return 0, err
	}
	for _, genre := range movie.Genres {
		_, err = tx.Exec(c, "insert into movies_genres(movie_id, genre_id) values($1, $2)", id, genre.Id)
		if err != nil {
			l.Error(err.Error())
			return 0, err
		}
	}

	for _, category := range movie.Category {
		_, err = tx.Exec(c, "insert into movies_categories(movie_id, categorie_id) values($1, $2)", id, category.Id)
		if err != nil {
			l.Error(err.Error())
			return 0, err
		}
	}

	for _, age := range movie.Ages {
		_, err = tx.Exec(c, "insert into movies_ages(movie_id, age_id) values($1, $2)", id, age.Id)
		if err != nil {
			l.Error(err.Error())
			return 0, err
		}
	}

	for _, allserie := range movie.AllSeries {
		_, err = tx.Exec(c, "insert into movies_allseries(movie_id, allserie_id) values($1, $2)", id, allserie.Id)
		if err != nil {
			l.Error(err.Error())
			return 0, err
		}
	}

	l.Info(fmt.Sprintf("проект %s добавлен успешно", movie.Title))

	err = tx.Commit(c)
	if err != nil {
		l.Error(err.Error())
		return 0, nil
	}

	return id, nil
}

func (r *MoviesRepository) IncrementViewsCount(c context.Context, id int) error {
	_, err := r.db.Exec(c, "UPDATE movies SET views_count = views_count + 1 WHERE id = $1", id)
	return err
}

func (r *MoviesRepository) Update(c context.Context, id int, updatedMovie models.Movie) error {
	l := logger.GetLogger()
	tx, err := r.db.Begin(c)
	if err != nil {
		l.Error(err.Error())
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(c)
		}
	}()

	_, err = tx.Exec(
		c,
		`
        update movies
        set 
            title = $1,
            description = $2,
            release_year = $3,
            director = $4,
            trailer_url = $5,
            poster_url = $6,
			duration = $7,
			video_url = $8,
			screen_src = $9
        where id = $10
        `,
		updatedMovie.Title,
		updatedMovie.Description,
		updatedMovie.ReleaseYear,
		updatedMovie.Director,
		updatedMovie.TrailerUrl,
		updatedMovie.PosterUrl,
		updatedMovie.Duration,
		updatedMovie.VideoUrl,
		updatedMovie.ScreenSrc,
		id)

	if err != nil {
		l.Error(err.Error())
		return err
	}

	_, err = tx.Exec(c, "DELETE FROM movies_genres WHERE movie_id = $1", id)
	if err != nil {
		l.Error(err.Error())
		return err
	}

	_, err = tx.Exec(c, "DELETE FROM movies_categories WHERE movie_id = $1", id)
	if err != nil {
		l.Error(err.Error())
		return err
	}

	_, err = tx.Exec(c, "DELETE FROM movies_ages WHERE movie_id = $1", id)
	if err != nil {
		l.Error(err.Error())
		return err
	}

	for _, genre := range updatedMovie.Genres {
		_, err = tx.Exec(c, "insert into movies_genres(movie_id, genre_id) values($1, $2)", id, genre.Id)
		if err != nil {
			l.Error(err.Error())
			return err
		}
	}

	for _, category := range updatedMovie.Category {
		_, err = tx.Exec(c, "insert into movies_categories(movie_id, categorie_id) values($1, $2)", id, category.Id)
		if err != nil {
			l.Error(err.Error())
			return err
		}
	}

	for _, age := range updatedMovie.Ages {
		_, err = tx.Exec(c, "insert into movies_ages(movie_id, age_id) values($1, $2)", id, age.Id)
		if err != nil {
			l.Error(err.Error())
			return err
		}
	}

	err = tx.Commit(c)
	if err != nil {
		l.Error(err.Error())
		return err
	}

	return nil
}

func (r *MoviesRepository) Delete(c context.Context, id int) error {
	l := logger.GetLogger()
	tx, err := r.db.Begin(c)
	if err != nil {
		l.Error(err.Error())
		return err
	}

	var movieTitle string
	row := r.db.QueryRow(c, "select title from movies where id = $1", id)
	err = row.Scan(&movieTitle)
	if err != nil {
		return err
	}

	l.Warn(fmt.Sprintf("Вы действительно хотите удалить %s?", movieTitle))

	_, err = tx.Exec(c, "DELETE FROM movies_genres WHERE movie_id = $1", id)
	if err != nil {
		l.Error(err.Error())
		tx.Rollback(c)
		return err
	}

	_, err = tx.Exec(c, "DELETE FROM movies_categories WHERE movie_id = $1", id)
	if err != nil {
		l.Error(err.Error())
		tx.Rollback(c)
		return err
	}

	_, err = tx.Exec(c, "DELETE FROM movies_ages WHERE movie_id = $1", id)
	if err != nil {
		l.Error(err.Error())
		tx.Rollback(c)
		return err
	}

	_, err = tx.Exec(c, "DELETE FROM movies_allseries WHERE movie_id = $1", id)
	if err != nil {
		l.Error(err.Error())
		tx.Rollback(c)
		return err
	}

	_, err = tx.Exec(c, "DELETE FROM movies WHERE id = $1", id)
	if err != nil {
		l.Error(err.Error())
		tx.Rollback(c)
		return err
	}

	l.Info(fmt.Sprintf("Фильм %s удален:", movieTitle))

	err = tx.Commit(c)
	if err != nil {
		l.Error(err.Error())
		return err
	}

	return nil
}

func (r *MoviesRepository) FindAllforUsers(c context.Context, filters models.MovieFilters) ([]models.MovieUser, error) {
	sql := ` 
    SELECT 
        m.id,
        m.title,
        m.description,
        m.release_year,
        m.director,
        m.trailer_url,
        m.poster_url,
		m.producer,
        g.id,
        g.title,
        g.poster_url,
        c.id,
        c.title,
        c.poster_url,
        a.id,
        a.age,
        a.poster_url,
        e.id, 
        e.series, 
        e.title,              
        e.trailer_url,
		e.duration,
		e.poster_url
    FROM movies m
    JOIN movies_genres mg ON mg.movie_id = m.id
    JOIN genres g ON mg.genre_id = g.id
    JOIN movies_categories mc ON mc.movie_id = m.id
    JOIN categories c ON mc.categorie_id = c.id
    JOIN movies_ages ma ON ma.movie_id = m.id
    JOIN ages a ON ma.age_id = a.id
    left JOIN movies_allseries me ON me.movie_id = m.id
    left JOIN allseries e ON me.allserie_id = e.id
    `
	sqlSeason := ` 
SELECT 
		 m.id,
        m.title,
        m.description,
        m.release_year,
        m.director,
        m.trailer_url,
        m.poster_url,
		m.producer,
		 s.id AS season_id,  
		 s.number AS season_number,
		 s.title AS season_title,
			e2.id AS episode_id,  
		 e2.series, 
		 e2.title,              
		 e2.trailer_url,
		 e2.duration,
		 e2.poster_url
	 FROM movies m
	 JOIN movies_seasons ms ON ms.movie_id = m.id
	 JOIN season s ON ms.season_id = s.id
	 LEFT JOIN seasons_allseries se ON se.season_id = s.id
	 LEFT JOIN allseries e2 ON se.allserie_id = e2.id
	 `

	params := pgx.NamedArgs{}

	if filters.SearchTerm != "" {
		//	'%%%s%%' => '%поиск%'
		sql = fmt.Sprintf("%s and m.title ilike @s", sql)
		params["s"] = fmt.Sprintf("%%%s%%", filters.SearchTerm)
	}
	if filters.GenreId != "" {
		sql = fmt.Sprintf("%s and g.id = @genreId", sql)
		params["genreId"] = filters.GenreId
	}

	if filters.AgeId != "" {
		sql = fmt.Sprintf("%s and a.id = @ageId", sql)
		params["AgeId"] = filters.AgeId
	}

	if filters.CategoryId != "" {
		sql = fmt.Sprintf("%s and c.id = @categoryId", sql)
		params["categoryId"] = filters.CategoryId
	}

	if filters.IsWatched != "" {
		isWatched, _ := strconv.ParseBool(filters.IsWatched)

		sql = fmt.Sprintf("%s and m.is_watched = @isWatched", sql)
		params["isWatched"] = isWatched
	}
	if filters.Sort != "" {
		identifier := pgx.Identifier{filters.Sort}
		sql = fmt.Sprintf("%s order by m.%s", sql, identifier.Sanitize())
	}

	l := logger.GetLogger()
	rows, err := r.db.Query(c, sql, params)
	if err != nil {
		l.Error(err.Error())
		return nil, err
	}

	movies := make([]*models.MovieUser, 0)
	moviesMap := make(map[int]*models.MovieUser)

	allseriesMap := make(map[int]*models.AllSeries, 0)

	for rows.Next() {
		var m models.MovieUser
		var g models.Genre
		var c models.Category
		var a models.Age
		var e models.AllSeries
		if err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Description,
			&m.ReleaseYear,
			&m.Director,
			&m.TrailerUrl,
			&m.PosterUrl,
			&m.Producer,
			&g.Id,
			&g.Title,
			&g.PosterUrl,
			&c.Id,
			&c.Title,
			&c.PosterUrl,
			&a.Id,
			&a.Age,
			&a.PosterUrl,
			&e.Id,
			&e.Series,
			&e.Title,
			&e.TrailerUrl,
			&e.Duration,
			&e.PosterUrl,
		); err != nil {
			return nil, err
		}

		if _, exists := moviesMap[m.Id]; !exists {
			moviesMap[m.Id] = &m
			movies = append(movies, &m)
		}

		genreExists := false
		for _, existingGenres := range moviesMap[m.Id].Genres {
			if existingGenres.Id == g.Id {
				genreExists = true
				break
			}
		}
		if !genreExists {
			moviesMap[m.Id].Genres = append(moviesMap[m.Id].Genres, g)
		}

		categoryExists := false
		for _, existingCategory := range moviesMap[m.Id].Category {
			if existingCategory.Id == c.Id {
				categoryExists = true
				break
			}
		}
		if !categoryExists {
			moviesMap[m.Id].Category = append(moviesMap[m.Id].Category, c)
		}

		ageExists := false
		for _, existingAge := range moviesMap[m.Id].Ages {
			if existingAge.Id == a.Id {
				ageExists = true
				break
			}
		}
		if !ageExists {
			moviesMap[m.Id].Ages = append(moviesMap[m.Id].Ages, a)
		}

		if e.Id != nil {
			if _, exists := allseriesMap[*e.Id]; !exists {
				allseriesMap[*e.Id] = &e
				moviesMap[m.Id].AllSeries = append(moviesMap[m.Id].AllSeries, e)
			}
		}
	}
	rows, err = r.db.Query(c, sqlSeason)
	if err != nil {
		l.Error(err.Error())
		return nil, err
	}
	//Season := make([]*models.Season, 0)
	seasonMap := make(map[int]*models.Season)

	for rows.Next() {
		var m models.MovieUser
		var s models.Season
		var e2 models.AllSeries
		if err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Description,
			&m.ReleaseYear,
			&m.Director,
			&m.TrailerUrl,
			&m.PosterUrl,
			&m.Producer,
			&s.Id,
			&s.Number,
			&s.Title,
			&e2.Id,
			&e2.Series,
			&e2.Title,
			&e2.TrailerUrl,
			&e2.Duration,
			&e2.PosterUrl,
		); err != nil {
			return nil, err
		}
		if _, exists := moviesMap[m.Id]; !exists {
			moviesMap[m.Id] = &m
			movies = append(movies, &m)
		}
		if _, exists := seasonMap[s.Id]; !exists {
			seasonMap[s.Id] = &s
			moviesMap[m.Id].Season = append(moviesMap[m.Id].Season, s)
		}
		if e2.Id != nil && *e2.Id != 0 {
			found := false
			for _, existingSeries := range seasonMap[s.Id].AllSeries {
				if existingSeries.Id != nil && *existingSeries.Id == *e2.Id {
					found = true
					break
				}
			}
			if !found {
				seasonMap[s.Id].AllSeries = append(seasonMap[s.Id].AllSeries, e2)
			}
		}
	}
	for _, movie := range movies {
		seasonExists := make(map[int]bool)
		for _, season := range seasonMap {
			if !seasonExists[season.Id] {
				movie.Season = append(movie.Season, *season)
				seasonExists[season.Id] = true
			}
		}
	}
	err = rows.Err()
	if err != nil {
		l.Error(err.Error())
		return nil, err
	}

	concreteMovies := make([]models.MovieUser, 0, len(movies))
	for _, v := range movies {
		concreteMovies = append(concreteMovies, *v)
	}

	return concreteMovies, nil
}

func (r *MoviesRepository) FindByIdUser(c context.Context, movieId int) (models.MovieUser, error) {
	sql :=
		`
 SELECT 
        m.id,
        m.title,
        m.description,
        m.release_year,
        m.director,
        m.trailer_url,
        m.poster_url,
		m.producer,
        g.id,
        g.title,
        g.poster_url,
        c.id,
        c.title,
        c.poster_url,
        a.id,
        a.age,
        a.poster_url,
        e.id, 
        e.series, 
        e.title,              
        e.trailer_url,
		e.duration,
		e.poster_url
    FROM movies m
    JOIN movies_genres mg ON mg.movie_id = m.id
    JOIN genres g ON mg.genre_id = g.id
    JOIN movies_categories mc ON mc.movie_id = m.id
    JOIN categories c ON mc.categorie_id = c.id
    JOIN movies_ages ma ON ma.movie_id = m.id
    JOIN ages a ON ma.age_id = a.id
    left JOIN movies_allseries me ON me.movie_id = m.id
    left JOIN allseries e ON me.allserie_id = e.id
where m.id = $1
	`
	sqlSeason := ` 
	SELECT 
		 m.id,
        m.title,
        m.description,
        m.release_year,
        m.director,
        m.trailer_url,
        m.poster_url,
		m.producer,
		 s.id AS season_id,  
		 s.number AS season_number,
		 s.title AS season_title,
			e2.id AS episode_id,  
		 e2.series, 
		 e2.title,              
		 e2.trailer_url,
		 e2.duration,
		 e2.poster_url
	 FROM movies m
	 JOIN movies_seasons ms ON ms.movie_id = m.id
	 JOIN season s ON ms.season_id = s.id
	 LEFT JOIN seasons_allseries se ON se.season_id = s.id
	 LEFT JOIN allseries e2 ON se.allserie_id = e2.id
	 WHERE m.id = $1
	 
 
	 `

	l := logger.GetLogger()

	rows, err := r.db.Query(c, sql, movieId)
	defer rows.Close()
	if err != nil {
		l.Error("Could not query database", zap.String("db_msg", err.Error()))
		return models.MovieUser{}, err
	}
	var movie *models.MovieUser

	categoriesMap := make(map[int]*models.Category, 0)
	category := make([]*models.Category, 0)

	genresMap := make(map[int]*models.Genre, 0)
	genre := make([]*models.Genre, 0)

	agesMap := make(map[int]*models.Age, 0)
	age := make([]*models.Age, 0)

	allseriesMap := make(map[int]*models.AllSeries, 0)
	allserie := make([]*models.AllSeries, 0)

	for rows.Next() {
		var m models.MovieUser
		var g models.Genre
		var c models.Category
		var a models.Age
		var e models.AllSeries
		if err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Description,
			&m.ReleaseYear,
			&m.Director,
			&m.TrailerUrl,
			&m.PosterUrl,
			&m.Producer,
			&g.Id,
			&g.Title,
			&g.PosterUrl,
			&c.Id,
			&c.Title,
			&c.PosterUrl,
			&a.Id,
			&a.Age,
			&a.PosterUrl,
			&e.Id,
			&e.Series,
			&e.Title,
			&e.TrailerUrl,
			&e.Duration,
			&e.PosterUrl,
		); err != nil {
			return models.MovieUser{}, err
		}

		if movie == nil {
			movie = &m
		}

		if _, exists := categoriesMap[c.Id]; !exists {
			categoriesMap[c.Id] = &c
			category = append(category, &c)
		}

		if _, exists := genresMap[g.Id]; !exists {
			genresMap[g.Id] = &g
			genre = append(genre, &g)
		}

		if _, exists := agesMap[a.Id]; !exists {
			agesMap[a.Id] = &a
			age = append(age, &a)
		}

		if e.Id != nil {
			if _, exists := allseriesMap[*e.Id]; !exists {
				allseriesMap[*e.Id] = &e
				allserie = append(allserie, &e)
			}
		}
	}

	err = rows.Err()
	if err != nil {
		return models.MovieUser{}, err
	}

	var categories []models.Category
	for _, cat := range category {
		categories = append(categories, *cat)
	}

	var genres []models.Genre
	for _, gen := range genre {
		genres = append(genres, *gen)
	}

	var ages []models.Age
	for _, age := range age {
		ages = append(ages, *age)
	}

	var allseries []models.AllSeries
	for _, allserie := range allserie {
		allseries = append(allseries, *allserie)
	}

	movie.Category = categories
	movie.Genres = genres
	movie.Ages = ages
	movie.AllSeries = allseries

	rows, err = r.db.Query(c, sqlSeason, movieId)
	defer rows.Close()
	if err != nil {
		l.Error("Ошибка запроса к базе", zap.String("db_msg", err.Error()))
		return models.MovieUser{}, err
	}
	//var season models.Season

	seasonMap := make(map[int]*models.Season)

	for rows.Next() {
		var m models.Movie
		var s models.Season
		var e2 models.AllSeries

		if err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Description,
			&m.ReleaseYear,
			&m.Director,
			&m.TrailerUrl,
			&m.PosterUrl,
			&m.Producer,
			&s.Id,
			&s.Number,
			&s.Title,
			&e2.Id,
			&e2.Series,
			&e2.Title,
			&e2.TrailerUrl,
			&e2.Duration,
			&e2.PosterUrl,
		); err != nil {
			return models.MovieUser{}, err
		}

		if _, exists := seasonMap[s.Id]; !exists {
			seasonMap[s.Id] = &s
		}

		if _, exists := allseriesMap[*e2.Id]; !exists {
			allseriesMap[*e2.Id] = &e2
		}
	}

	for _, v := range seasonMap {
		for _, t := range allseriesMap {
			v.AllSeries = append(v.AllSeries, *t)
		}
		movie.Season = append(movie.Season, *v)
	}

	return *movie, nil
}
