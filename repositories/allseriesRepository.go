package repositories

import (
	"context"
	"fmt"
	"goozinshe/logger"
	"goozinshe/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AllSeriesRepository struct {
	db *pgxpool.Pool
}

func NewAllSeriesRepository(conn *pgxpool.Pool) *AllSeriesRepository {
	return &AllSeriesRepository{db: conn}
}

func (r *AllSeriesRepository) IsAdmin(c context.Context, id int) (bool, error) {
	var role string
	row := r.db.QueryRow(c, "select is_admin from users where id = $1", id)
	err := row.Scan(&role)
	if err != nil {
		return false, err
	}

	return role == "admin", nil // проверяем, является ли роль "admin"
}

func (r *AllSeriesRepository) FindAllByIds(c context.Context, ids []int) ([]models.AllSeries, error) {
	rows, err := r.db.Query(c, "select id, series,  title,  trailer_url, duration from allseries where id = any($1)", ids)
	defer rows.Close()
	if err != nil {
		l := logger.GetLogger()
		l.Error(err.Error())
		return nil, err
	}

	allseries := make([]models.AllSeries, 0)

	for rows.Next() {
		var allserie models.AllSeries
		err = rows.Scan(
			&allserie.Id,
			&allserie.Series,
			&allserie.Title,
			&allserie.TrailerUrl,
			&allserie.Duration)
		if err != nil {
			l := logger.GetLogger()
			l.Error(err.Error())
			return nil, err
		}

		allseries = append(allseries, allserie)
	}

	return allseries, nil
}

func (r *AllSeriesRepository) Create(c context.Context, serie models.AllSeries) (int, error) {
	var id int
	// tx, err := r.db.Begin(c)

	row := r.db.QueryRow(c, "insert into allseries (series, title, trailer_url, duration) values($1, $2, $3, $4) returning id", serie.Series, serie.Title, serie.TrailerUrl, serie.Duration)
	err := row.Scan(&id)
	if err != nil {
		l := logger.GetLogger()
		l.Error(err.Error())
		return 0, err
	}
	return id, nil
}

func (r *AllSeriesRepository) FindById(c context.Context, movieId int) (models.AllSeries, error) {
	var allserie models.AllSeries
	row := r.db.QueryRow(c, "select id, series,  title, trailer_url, duration from allseries where id = $1", movieId)
	err := row.Scan(
		&allserie.Id,
		&allserie.Series,
		&allserie.Title,
		&allserie.TrailerUrl,
		&allserie.Duration)
	if err != nil {
		l := logger.GetLogger()
		l.Error(err.Error())
		return models.AllSeries{}, err
	}
	return allserie, nil
}

func (r *AllSeriesRepository) FindAll(c context.Context) ([]models.AllSeries, error) {
	rows, err := r.db.Query(c, "select id, series, title, trailer_url, duration from allseries")
	defer rows.Close()
	if err != nil {
		l := logger.GetLogger()
		l.Error(err.Error())
		return nil, err
	}

	allseries := make([]models.AllSeries, 0)
	for rows.Next() {
		var allserie models.AllSeries
		err := rows.Scan(
			&allserie.Id,
			&allserie.Series,
			&allserie.Title,
			&allserie.TrailerUrl,
			&allserie.Duration)
		if err != nil {
			l := logger.GetLogger()
			l.Error(err.Error())
			return nil, err
		}

		allseries = append(allseries, allserie)
	}

	return allseries, nil
}

func (r *AllSeriesRepository) Update(c context.Context, movieId int, allserie models.AllSeries) error {
	_, err := r.db.Exec(c, `update allseries set 
							series = $1 ,
							title = $2, 
							trailer_url = $3,
							duration = $4,
							poster_url = $5
							where id = $6`,
		&allserie.Series,
		&allserie.Title,
		&allserie.TrailerUrl,
		&allserie.Duration,
		&allserie.PosterUrl,
		movieId)
	if err != nil {
		l := logger.GetLogger()
		l.Error(err.Error())
		return err
	}

	return nil
}

//series, title, description, release_year, director, rating, trailer_url

func (r *AllSeriesRepository) Delete(c context.Context, movieId int) error {
	l := logger.GetLogger()

	var serieTitle string
	row := r.db.QueryRow(c, "select title from allseries where id = $1", movieId)
	err := row.Scan(&serieTitle)
	if err != nil {
		return err
	}

	l.Warn(fmt.Sprintf("Вы действительно хотите удалить %s серию?", serieTitle))

	_, err = r.db.Exec(c, "delete from allseries where id = $1", movieId)
	if err != nil {
		return err
	}

	l.Info(fmt.Sprintf("серия %s удалена", serieTitle))

	return nil
}
