package repositories

import (
	"context"
	"goozinshe/logger"
	"goozinshe/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SelectedlistRepository struct {
	db *pgxpool.Pool
}

func NewSelectedlistRepository(db *pgxpool.Pool) *SelectedlistRepository {
	return &SelectedlistRepository{db: db}
}

func (r *SelectedlistRepository) GetMoviesFromSelectedlist(c context.Context) ([]models.Movie, error) {
	sql :=
		`
SELECT 
        m.id,
        m.title,
        m.description,
        m.release_year,
        m.director,
        m.rating,
        CASE WHEN sl.movie_id IS NOT NULL THEN TRUE ELSE FALSE END AS is_favourite,
        m.trailer_url,
        m.poster_url,
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
        e.trailer_url
    FROM selected sl
	JOIN movies m on sl.movie_id = m.id
    JOIN movies_genres mg ON mg.movie_id = m.id
    JOIN genres g ON mg.genre_id = g.id
    JOIN movies_categories mc ON mc.movie_id = m.id
    JOIN categories c ON mc.categorie_id = c.id
    JOIN movies_ages ma ON ma.movie_id = m.id
    JOIN ages a ON ma.age_id = a.id
    left JOIN movies_allseries me ON me.movie_id = m.id
    left JOIN allseries e ON me.allserie_id = e.id
order by sl.added_at

 `

	l := logger.GetLogger()
	rows, err := r.db.Query(c, sql)
	if err != nil {
		l.Error(err.Error())
		return nil, err
	}

	movies := make([]*models.Movie, 0)
	moviesMap := make(map[int]*models.Movie)
	allseriesMap := make(map[int]*models.AllSeries, 0)

	for rows.Next() {
		var m models.Movie
		var g models.Genre
		var c models.Category
		var a models.Age
		var e models.AllSeries

		err := rows.Scan(
			&m.Id,
			&m.Title,
			&m.Description,
			&m.ReleaseYear,
			&m.Director,
			&m.Rating,
			&m.IsFavourite,
			&m.TrailerUrl,
			&m.PosterUrl,
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
			// &e.Description,
			// &e.ReleaseYear,
			// &e.Director,
			// &e.Rating,
			&e.TrailerUrl,
		)
		if err != nil {
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
		for _, exisitingAges := range moviesMap[m.Id].Ages {
			if exisitingAges.Id == a.Id {
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

func (r *SelectedlistRepository) AddToSelectedMovie(c context.Context, movieId int) error {
	_, err := r.db.Exec(c, "insert into selected (movie_id,  added_at) values($1, $2)", movieId, time.Now())
	if err != nil {
		l := logger.GetLogger()
		l.Error(err.Error())
	}

	return err
}

func (r *SelectedlistRepository) RemoveFromSelectedlist(c context.Context, movieId int) error {
	if movieId != 0 {
		_, err := r.db.Exec(c, "DELETE FROM selected WHERE movie_id = $1", movieId)
		if err != nil {
			l := logger.GetLogger()
			l.Error("Error deleting movie from selected list: " + err.Error())
			return err
		}
	}

	return nil
}
